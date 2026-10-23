package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"

	"latihan-fiber/app/repository"
	"latihan-fiber/config"
	"latihan-fiber/database"
	"latihan-fiber/helper"
)

type UserAccount struct {
	ID       int
	Username string
	Role     string
	Token    string
}

type CourseInfo struct {
	ID            int
	Code          string
	Credits       int
	Quota         int
	EnrolledCount int
}

func fire(n int, job func(i int) string) map[string]int {
	var wg sync.WaitGroup
	start := make(chan struct{})
	results := make(chan string, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		// KESALAHAN SENGAJA (Bug #2 di Modul):
		// parameter out dideklarasikan sebagai <-chan string (receive-only),
		// sehingga perintah out <- job(i) di bawahnya ditolak compiler.
		go func(i int, out chan<- string) {
			defer wg.Done()
			<-start
			out <- job(i)
		}(i, results)
	}
	close(start) // buka gerbang
	wg.Wait()    // tunggu SEMUA goroutine selesai
	close(results)
	tally := map[string]int{}
	for r := range results {
		tally[r]++
	}
	return tally
}

func main() {
	mode := flag.String("mode", "crowd", "mode pengujian: crowd | credits | swap")
	baseURL := flag.String("url", "http://localhost:3000/api/v1", "base URL server")
	flag.Parse()

	config.LoadEnv()
	ctx := context.Background()
	pool, err := database.NewPool(ctx)
	if err != nil {
		fmt.Printf("Gagal koneksi database: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	jwtSecret := config.GetEnv("JWT_SECRET", "supersecretloadtestkey123456789012")
	jwtManager := helper.NewJWTManager(jwtSecret, "latihan-fiber", 1*time.Hour, 24*time.Hour)

	// Ambil 60 user mhs001 s/d mhs060
	rows, err := pool.Query(ctx, "SELECT id, username, role FROM users WHERE username LIKE 'mhs%' ORDER BY username")
	if err != nil {
		fmt.Printf("Gagal membaca users: %v\n", err)
		os.Exit(1)
	}
	defer rows.Close()

	var users []UserAccount
	for rows.Next() {
		var u UserAccount
		if err := rows.Scan(&u.ID, &u.Username, &u.Role); err != nil {
			fmt.Printf("Error scan user: %v\n", err)
			os.Exit(1)
		}
		token, err := jwtManager.GenerateAccessToken(u.ID, u.Role)
		if err != nil {
			fmt.Printf("Error generate token: %v\n", err)
			os.Exit(1)
		}
		u.Token = token
		users = append(users, u)
	}

	if len(users) < 60 {
		fmt.Printf("Peringatan: hanya ditemukan %d user mhs. Jalankan migration 006 terlebih dahulu!\n", len(users))
	}

	client := &http.Client{Timeout: 10 * time.Second}

	switch *mode {
	case "crowd":
		runCrowd(client, pool, users, *baseURL)
	case "credits":
		runCredits(client, pool, users, *baseURL)
	case "swap":
		runSwap(client, pool, users, *baseURL)
	default:
		fmt.Printf("Mode '%s' tidak dikenal. Pilih crowd, credits, atau swap.\n", *mode)
	}
}

func runCrowd(client *http.Client, pool repository.DBTX, users []UserAccount, baseURL string) {
	fmt.Println("60 mahasiswa berebut kelas IF399 secara serentak...")

	var if399 CourseInfo
	err := pool.QueryRow(context.Background(), "SELECT id, code, quota, enrolled_count FROM courses WHERE code = 'IF399'").
		Scan(&if399.ID, &if399.Code, &if399.Quota, &if399.EnrolledCount)
	if err != nil {
		fmt.Printf("Gagal membaca kelas IF399: %v\n", err)
		return
	}

	n := 60
	if len(users) < n {
		n = len(users)
	}

	startTime := time.Now()
	tally := fire(n, func(i int) string {
		url := fmt.Sprintf("%s/courses/%d/enroll", baseURL, if399.ID)
		req, _ := http.NewRequest(http.MethodPost, url, bytes.NewBuffer([]byte("{}")))
		req.Header.Set("Authorization", "Bearer "+users[i].Token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return "ERR: " + err.Error()
		}
		defer resp.Body.Close()

		bodyBytes, _ := io.ReadAll(resp.Body)
		if resp.StatusCode == http.StatusCreated {
			return "201 OK"
		}
		var errResp config.ErrorResponse
		_ = json.Unmarshal(bodyBytes, &errResp)
		if errResp.Code != "" {
			return fmt.Sprintf("%d %s", resp.StatusCode, errResp.Code)
		}
		return fmt.Sprintf("%d HTTP", resp.StatusCode)
	})
	duration := time.Since(startTime)

	printTally(tally)
	fmt.Printf("waktu total : %v\n", duration)

	// Intip database langsung
	var updated CourseInfo
	var count int
	_ = pool.QueryRow(context.Background(), "SELECT enrolled_count FROM courses WHERE id = $1", if399.ID).Scan(&updated.EnrolledCount)
	_ = pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM enrollments WHERE course_id = $1", if399.ID).Scan(&count)

	statusKonsisten := "KONSISTEN"
	if updated.EnrolledCount != count || updated.EnrolledCount > if399.Quota {
		statusKonsisten = "TIDAK KONSISTEN"
	}
	fmt.Printf("IF399 kuota=%d  enrolled_count=%d  baris enrollments=%d  -> %s\n", if399.Quota, updated.EnrolledCount, count, statusKonsisten)
}

func runCredits(client *http.Client, pool repository.DBTX, users []UserAccount, baseURL string) {
	fmt.Println("10 mahasiswa, masing-masing mendaftar ke 12 kelas 3 SKS secara serentak (batas 24 SKS)...")

	rows, err := pool.Query(context.Background(), "SELECT id, code, credits FROM courses WHERE code LIKE 'IF3%' OR code LIKE 'IF1%' ORDER BY id")
	if err != nil {
		fmt.Printf("Gagal membaca daftar kelas: %v\n", err)
		return
	}
	defer rows.Close()

	var courseIDs []int
	for rows.Next() {
		var id, cr int
		var cd string
		_ = rows.Scan(&id, &cd, &cr)
		courseIDs = append(courseIDs, id)
	}

	numStudents := 10
	if len(users) < numStudents {
		numStudents = len(users)
	}

	type JobParam struct {
		User     UserAccount
		CourseID int
	}
	var jobs []JobParam
	for i := 0; i < numStudents; i++ {
		for _, cid := range courseIDs {
			jobs = append(jobs, JobParam{User: users[i], CourseID: cid})
		}
	}

	startTime := time.Now()
	tally := fire(len(jobs), func(i int) string {
		param := jobs[i]
		url := fmt.Sprintf("%s/courses/%d/enroll", baseURL, param.CourseID)
		req, _ := http.NewRequest(http.MethodPost, url, bytes.NewBuffer([]byte("{}")))
		req.Header.Set("Authorization", "Bearer "+param.User.Token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return "ERR: " + err.Error()
		}
		defer resp.Body.Close()

		bodyBytes, _ := io.ReadAll(resp.Body)
		if resp.StatusCode == http.StatusCreated {
			return "201 OK"
		}
		var errResp config.ErrorResponse
		_ = json.Unmarshal(bodyBytes, &errResp)
		if errResp.Code != "" {
			return fmt.Sprintf("%d %s", resp.StatusCode, errResp.Code)
		}
		return fmt.Sprintf("%d HTTP", resp.StatusCode)
	})
	duration := time.Since(startTime)

	printTally(tally)
	fmt.Printf("waktu total : %v\n", duration)

	// Periksa SKS seluruh 10 mahasiswa
	var maxSKS int
	err = pool.QueryRow(context.Background(), `
		SELECT COALESCE(MAX(total_sks), 0) FROM (
			SELECT e.user_id, SUM(c.credits) as total_sks
			FROM enrollments e
			JOIN courses c ON c.id = e.course_id
			WHERE e.user_id IN (SELECT id FROM users WHERE username LIKE 'mhs%' ORDER BY username LIMIT 10)
			GROUP BY e.user_id
		) sub
	`).Scan(&maxSKS)

	konsisten := "KONSISTEN"
	if maxSKS > 24 {
		konsisten = fmt.Sprintf("TIDAK KONSISTEN (maksimal %d SKS)", maxSKS)
	}
	fmt.Printf("seluruh %d mahasiswa <= 24 SKS -> %s\n", numStudents, konsisten)
}

func runSwap(client *http.Client, pool repository.DBTX, users []UserAccount, baseURL string) {
	fmt.Println("20 pasang mahasiswa pindah kelas berlawanan arah secara serentak...")

	var if101, if102 CourseInfo
	_ = pool.QueryRow(context.Background(), "SELECT id, code, quota, enrolled_count FROM courses WHERE code = 'IF101'").
		Scan(&if101.ID, &if101.Code, &if101.Quota, &if101.EnrolledCount)
	_ = pool.QueryRow(context.Background(), "SELECT id, code, quota, enrolled_count FROM courses WHERE code = 'IF102'").
		Scan(&if102.ID, &if102.Code, &if102.Quota, &if102.EnrolledCount)

	// Persiapan: daftarkan 20 mhs pertama ke IF101, dan 20 mhs berikutnya ke IF102
	for i := 0; i < 20; i++ {
		_, _ = pool.Exec(context.Background(), "INSERT INTO enrollments (course_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", if101.ID, users[i].ID)
	}
	for i := 20; i < 40; i++ {
		_, _ = pool.Exec(context.Background(), "INSERT INTO enrollments (course_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", if102.ID, users[i].ID)
	}
	_, _ = pool.Exec(context.Background(), "UPDATE courses SET enrolled_count = (SELECT COUNT(*) FROM enrollments WHERE course_id = courses.id)")

	type SwapTask struct {
		User   UserAccount
		FromID int
		ToID   int
	}
	var tasks []SwapTask
	// 20 mhs pindah IF101 -> IF102
	for i := 0; i < 20; i++ {
		tasks = append(tasks, SwapTask{User: users[i], FromID: if101.ID, ToID: if102.ID})
	}
	// 20 mhs pindah IF102 -> IF101
	for i := 20; i < 40; i++ {
		tasks = append(tasks, SwapTask{User: users[i], FromID: if102.ID, ToID: if101.ID})
	}

	deadlocksBefore := getDeadlockCount(pool)

	startTime := time.Now()
	tally := fire(len(tasks), func(i int) string {
		task := tasks[i]
		url := fmt.Sprintf("%s/enrollments/swap", baseURL)
		payload, _ := json.Marshal(map[string]int{
			"from_course_id": task.FromID,
			"to_course_id":   task.ToID,
		})
		req, _ := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(payload))
		req.Header.Set("Authorization", "Bearer "+task.User.Token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return "ERR: " + err.Error()
		}
		defer resp.Body.Close()

		bodyBytes, _ := io.ReadAll(resp.Body)
		if resp.StatusCode == http.StatusOK {
			return "200 OK"
		}
		var errResp config.ErrorResponse
		_ = json.Unmarshal(bodyBytes, &errResp)
		if errResp.Code != "" {
			return fmt.Sprintf("%d %s", resp.StatusCode, errResp.Code)
		}
		return fmt.Sprintf("%d HTTP", resp.StatusCode)
	})
	duration := time.Since(startTime)

	deadlocksAfter := getDeadlockCount(pool)
	deadlocks := deadlocksAfter - deadlocksBefore

	printTally(tally)
	fmt.Printf("waktu total : %v\n", duration)
	fmt.Printf("deadlock terdeteksi PostgreSQL selama uji: %d\n", deadlocks)

	var c101, e101, c102, e102 int
	_ = pool.QueryRow(context.Background(), "SELECT enrolled_count FROM courses WHERE id = $1", if101.ID).Scan(&c101)
	_ = pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM enrollments WHERE course_id = $1", if101.ID).Scan(&e101)
	_ = pool.QueryRow(context.Background(), "SELECT enrolled_count FROM courses WHERE id = $1", if102.ID).Scan(&c102)
	_ = pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM enrollments WHERE course_id = $1", if102.ID).Scan(&e102)

	fmt.Printf("IF101 kuota=40 enrolled_count=%d baris enrollments=%d -> %s\n", c101, e101, status(c101 == e101 && c101 == 20))
	fmt.Printf("IF102 kuota=40 enrolled_count=%d baris enrollments=%d -> %s\n", c102, e102, status(c102 == e102 && c102 == 20))
}

func getDeadlockCount(pool repository.DBTX) int {
	var count int
	_ = pool.QueryRow(context.Background(), "SELECT deadlocks FROM pg_stat_database WHERE datname = 'praktikum_backend'").Scan(&count)
	return count
}

func status(ok bool) string {
	if ok {
		return "KONSISTEN"
	}
	return "TIDAK KONSISTEN"
}

func printTally(tally map[string]int) {
	fmt.Println("hasil request:")
	keys := make([]string, 0, len(tally))
	for k := range tally {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf("  %-25s %d\n", k, tally[k])
	}
}
