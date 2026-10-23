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

type SupervisorInfo struct {
	ID              int
	Name            string
	Quota           int
	SupervisedCount int
	Version         int
}

func fire(n int, job func(i int) string) map[string]int {
	var wg sync.WaitGroup
	start := make(chan struct{})
	results := make(chan string, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
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

	// Ambil token admin untuk permission supervision:manage
	var adminID int
	_ = pool.QueryRow(ctx, "SELECT id FROM users WHERE role = 'admin' LIMIT 1").Scan(&adminID)
	if adminID == 0 {
		adminID = 1
	}
	adminToken, err := jwtManager.GenerateAccessToken(adminID, "admin")
	if err != nil {
		fmt.Printf("Gagal generate admin token: %v\n", err)
		os.Exit(1)
	}

	client := &http.Client{Timeout: 10 * time.Second}

	runSupervisorLoadTest(client, pool, adminToken, *baseURL)
}

func runSupervisorLoadTest(client *http.Client, pool repository.DBTX, adminToken, baseURL string) {
	ctx := context.Background()
	fmt.Println("30 mahasiswa serentak ditetapkan ke pembimbing ID 1 (kuota 3)...")

	// Ambil data pembimbing ID 1
	var sup SupervisorInfo
	err := pool.QueryRow(ctx, "SELECT id, name, quota, supervised_count, version FROM supervisors WHERE id = 1").
		Scan(&sup.ID, &sup.Name, &sup.Quota, &sup.SupervisedCount, &sup.Version)
	if err != nil {
		fmt.Printf("Gagal membaca pembimbing ID 1: %v\n", err)
		return
	}

	// Buat 30 student ID: std_001 s/d std_030
	n := 30
	studentIDs := make([]string, n)
	for i := 0; i < n; i++ {
		studentIDs[i] = fmt.Sprintf("std_%03d", i+1)
	}

	startTime := time.Now()
	tally := fire(n, func(i int) string {
		url := fmt.Sprintf("%s/supervisors/%d/assign", baseURL, sup.ID)
		payload, _ := json.Marshal(map[string]string{
			"student_id": studentIDs[i],
		})
		req, _ := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(payload))
		req.Header.Set("Authorization", "Bearer "+adminToken)
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
	var updated SupervisorInfo
	var count int
	_ = pool.QueryRow(ctx, "SELECT supervised_count FROM supervisors WHERE id = 1").Scan(&updated.SupervisedCount)
	_ = pool.QueryRow(ctx, "SELECT COUNT(*) FROM supervisions WHERE supervisor_id = 1").Scan(&count)

	statusKonsisten := "KONSISTEN"
	if updated.SupervisedCount != count || updated.SupervisedCount > sup.Quota {
		statusKonsisten = "TIDAK KONSISTEN"
	}
	fmt.Printf("Supervisor ID 1 (%s) kuota=%d supervised_count=%d baris supervisions=%d -> %s\n",
		sup.Name, sup.Quota, updated.SupervisedCount, count, statusKonsisten)
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
