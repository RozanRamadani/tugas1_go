package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"latihan-fiber/config"
	"latihan-fiber/database"
	"latihan-fiber/helper"
)

var (
	baseURL   = "http://localhost:3000/api/v1"
	client    = &http.Client{Timeout: 10 * time.Second}
	pool      *pgxpool.Pool
	jwtSecret string
	jwtMgr    *helper.JWTManager
)

func main() {
	testName := flag.String("test", "all", "c1 | c2 | c3 | c4 | c5 | c6 | d2 | d3 | all")
	flag.Parse()

	config.LoadEnv()
	ctx := context.Background()

	var err error
	pool, err = database.NewPool(ctx)
	if err != nil {
		log.Fatalf("db pool: %v", err)
	}
	defer pool.Close()

	jwtSecret = config.GetEnv("JWT_SECRET", "supersecretloadtestkey123456789012")
	jwtMgr = helper.NewJWTManager(jwtSecret, config.GetEnv("JWT_ISSUER", "praktikum-backend"), 24*time.Hour)

	// Ensure users exist
	mhsujiToken := ensureUser(ctx, "mhsuji", "mhsuji@student.unair.ac.id", "user")
	salahalamatToken := ensureUser(ctx, "salahalamat", "salah@kampus.invalid", "user")
	adminToken := ensureUser(ctx, "admin", "admin@ex.com", "admin")

	fmt.Printf("[OK] mhsujiToken: %s...\n", mhsujiToken[:20])
	fmt.Printf("[OK] salahalamatToken: %s...\n", salahalamatToken[:20])
	fmt.Printf("[OK] adminToken: %s...\n", adminToken[:20])

	switch *testName {
	case "c1":
		runC1(ctx, mhsujiToken)
	case "c2":
		runC2(ctx, mhsujiToken, adminToken)
	case "c3":
		runC3(ctx, mhsujiToken)
	case "c4":
		runC4(ctx, salahalamatToken, adminToken)
	case "c5_prep":
		runC5Prep(ctx, mhsujiToken)
	case "c5_check":
		runC5Check(ctx)
	case "c6":
		runC6(ctx, mhsujiToken)
	case "d2":
		runD2(ctx, adminToken)
	case "d3":
		runD3(ctx, mhsujiToken)
	default:
		fmt.Println("Gunakan flag -test=c1|c2|c3|c4|c5|c6|d2")
	}
}

func ensureUser(ctx context.Context, username, email, role string) string {
	var id int
	err := pool.QueryRow(ctx, "SELECT id FROM users WHERE username = $1", username).Scan(&id)
	if err != nil {
		hash, _ := helper.HashPassword("password123")
		err = pool.QueryRow(ctx, `
			INSERT INTO users (username, email, password, role, is_active)
			VALUES ($1, $2, $3, $4, true)
			RETURNING id`, username, email, hash, role).Scan(&id)
		if err != nil {
			log.Fatalf("insert user %s: %v", username, err)
		}
	} else {
		// update email/role
		_, _ = pool.Exec(ctx, "UPDATE users SET email = $1, role = $2 WHERE id = $3", email, role, id)
	}

	token, err := jwtMgr.GenerateAccessToken(id, role)
	if err != nil {
		log.Fatalf("generate token for %s: %v", username, err)
	}
	return token
}

func runResetSQL() {
	cmd := exec.Command("psql", "-U", "postgres", "-d", "praktikum_backend", "-f", "reset11.sql")
	cmd.Env = append(os.Environ(), "PGPASSWORD=admin")
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("reset11.sql failed: %v, output: %s", err, string(out))
	}
	fmt.Println("[RESET] reset11.sql executed successfully.")
}

func clearMailLog() {
	_ = os.MkdirAll("logs", 0o755)
	_ = os.WriteFile("logs/mail.log", []byte(""), 0o600)
	fmt.Println("[LOG] logs/mail.log cleared.")
}

func countMailLog(pattern string) int {
	data, err := os.ReadFile("logs/mail.log")
	if err != nil {
		return 0
	}
	lines := strings.Split(string(data), "\n")
	cnt := 0
	for _, l := range lines {
		if strings.Contains(l, pattern) {
			cnt++
		}
	}
	return cnt
}

func httpReq(method, path, token string, body any) (int, map[string]any, error) {
	var bodyReader io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		bodyReader = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, baseURL+path, bodyReader)
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	respBytes, _ := io.ReadAll(resp.Body)
	var res map[string]any
	_ = json.Unmarshal(respBytes, &res)
	return resp.StatusCode, res, nil
}

// ---------------------------------------------------------------------
// C.1 Jalur normal
// ---------------------------------------------------------------------
func runC1(ctx context.Context, token string) {
	fmt.Println("==================================================")
	fmt.Println("MENJALANKAN UJI C.1: Jalur Normal")
	fmt.Println("==================================================")
	runResetSQL()
	clearMailLog()

	// 1. POST /courses/2/enroll
	status1, _, _ := httpReq("POST", "/courses/2/enroll", token, nil)
	fmt.Printf("Langkah 1: POST /courses/2/enroll -> Status: %d (expected 201)\n", status1)

	// 2. POST /courses/2/enroll (diulang)
	status2, res2, _ := httpReq("POST", "/courses/2/enroll", token, nil)
	fmt.Printf("Langkah 2: POST /courses/2/enroll (diulang) -> Status: %d, Code: %v (expected 409 ALREADY_ENROLLED)\n", status2, res2["code"])

	// 3. DELETE /courses/2/enroll
	status3, _, _ := httpReq("DELETE", "/courses/2/enroll", token, nil)
	fmt.Printf("Langkah 3: DELETE /courses/2/enroll -> Status: %d (expected 204)\n", status3)

	// 4. POST /courses/2/enroll
	status4, _, _ := httpReq("POST", "/courses/2/enroll", token, nil)
	fmt.Printf("Langkah 4: POST /courses/2/enroll -> Status: %d (expected 201)\n", status4)

	// 5. POST /enrollments/swap {"from_course_id": 2, "to_course_id": 4}
	status5, _, _ := httpReq("POST", "/enrollments/swap", token, map[string]any{"from_course_id": 2, "to_course_id": 4})
	fmt.Printf("Langkah 5: POST /enrollments/swap -> Status: %d (expected 200)\n", status5)

	fmt.Println("Menunggu 5 detik agar worker dan relay menyelesaikan task...")
	time.Sleep(5 * time.Second)

	// Verifikasi 1: Tabel outbox untuk mhsuji berisi TEPAT empat baris
	rows, err := pool.Query(ctx, `
		SELECT o.id, o.event_type, o.published_at IS NOT NULL
		FROM outbox o
		WHERE (payload->>'user_id')::INT IN (SELECT id FROM users WHERE username = 'mhsuji')
		ORDER BY o.id ASC`)
	if err != nil {
		log.Fatalf("query outbox: %v", err)
	}
	defer rows.Close()

	fmt.Println("\n--- Tabel Outbox untuk mhsuji ---")
	outboxCount := 0
	for rows.Next() {
		var id int64
		var evt string
		var pub bool
		_ = rows.Scan(&id, &evt, &pub)
		outboxCount++
		fmt.Printf("  Row %d: id=%d event_type=%s published=%v\n", outboxCount, id, evt, pub)
	}
	fmt.Printf("Jumlah baris outbox mhsuji: %d (expected 4)\n", outboxCount)

	// Verifikasi 2: GET /notifications/mine
	statusN, resN, _ := httpReq("GET", "/notifications/mine", token, nil)
	fmt.Printf("\n--- GET /notifications/mine (Status: %d) ---\n", statusN)
	if dataList, ok := resN["data"].([]any); ok {
		fmt.Printf("Jumlah notifikasi: %d (expected 4)\n", len(dataList))
		for i, item := range dataList {
			m := item.(map[string]any)
			fmt.Printf("  [%d] subject: %q (outbox_id: %v)\n", i+1, m["subject"], m["outbox_id"])
		}
	}

	// Verifikasi 3: logs/mail.log
	mailCount := countMailLog("mhsuji@")
	fmt.Printf("\n--- logs/mail.log untuk mhsuji@: %d baris (expected 4) ---\n", mailCount)
	logData, _ := os.ReadFile("logs/mail.log")
	fmt.Println(string(logData))
}

// ---------------------------------------------------------------------
// C.2 Redis mati, KRS tetap berjalan
// ---------------------------------------------------------------------
func runC2(ctx context.Context, token, adminToken string) {
	fmt.Println("==================================================")
	fmt.Println("MENJALANKAN UJI C.2: Redis mati, KRS tetap berjalan")
	fmt.Println("==================================================")
	// Langkah 1: Matikan Redis
	fmt.Println("1. Mematikan Redis...")
	_ = exec.Command("redis-cli", "shutdown").Run()
	time.Sleep(1 * time.Second)

	// Langkah 2: Daftar ke kelas 5, 6, dan 7
	fmt.Println("2. Mendaftar ke kelas 5, 6, dan 7 saat Redis mati:")
	t0 := time.Now()
	s5, _, _ := httpReq("POST", "/courses/5/enroll", token, nil)
	d5 := time.Since(t0)
	t1 := time.Now()
	s6, _, _ := httpReq("POST", "/courses/6/enroll", token, nil)
	d6 := time.Since(t1)
	t2 := time.Now()
	s7, _, _ := httpReq("POST", "/courses/7/enroll", token, nil)
	d7 := time.Since(t2)
	fmt.Printf("  POST /courses/5/enroll: %d (durasi: %v)\n", s5, d5)
	fmt.Printf("  POST /courses/6/enroll: %d (durasi: %v)\n", s6, d6)
	fmt.Printf("  POST /courses/7/enroll: %d (durasi: %v)\n", s7, d7)

	// Langkah 3: Tunggu 5 detik, periksa GET /admin/queue
	fmt.Println("3. Menunggu 5 detik lalu memeriksa GET /admin/queue...")
	time.Sleep(5 * time.Second)
	_, resQ, _ := httpReq("GET", "/admin/queue", adminToken, nil)
	dataQ, _ := json.MarshalIndent(resQ, "", "  ")
	fmt.Printf("GET /admin/queue response:\n%s\n", string(dataQ))

	// Langkah 4: Nyalakan Redis
	fmt.Println("4. Menyalakan kembali Redis server...")
	_ = exec.Command("powershell.exe", "-Command", "Start-Process -FilePath 'C:\\laragon\\bin\\redis\\redis-x64-5.0.14.1\\redis-server.exe' -WindowStyle Hidden").Run()
	time.Sleep(2 * time.Second)

	fmt.Println("Menunggu 8 detik agar relay memindahkan outbox ke Redis dan worker memprosesnya...")
	time.Sleep(8 * time.Second)

	// Periksa outbox_pending dan GET /notifications/mine
	_, resQ2, _ := httpReq("GET", "/admin/queue", adminToken, nil)
	dataQ2, _ := json.MarshalIndent(resQ2, "", "  ")
	fmt.Printf("GET /admin/queue setelah Redis kembali nyala:\n%s\n", string(dataQ2))

	_, resMine, _ := httpReq("GET", "/notifications/mine", token, nil)
	if list, ok := resMine["data"].([]any); ok {
		fmt.Printf("Jumlah notifikasi mhsuji sekarang: %d (expected 7: 4 dari C.1 + 3 dari C.2)\n", len(list))
	}
}

// ---------------------------------------------------------------------
// C.3 Temporary failure
// ---------------------------------------------------------------------
func runC3(ctx context.Context, token string) {
	fmt.Println("==================================================")
	fmt.Println("MENJALANKAN UJI C.3: Temporary Failure (MAILER_FAIL_EVERY=3)")
	fmt.Println("==================================================")
	runResetSQL()
	clearMailLog()

	fmt.Println("Mendaftarkan mhsuji ke kelas 1, 2, 3, 4, 5, dan 6...")
	for _, cid := range []int{1, 2, 3, 4, 5, 6} {
		st, _, _ := httpReq("POST", fmt.Sprintf("/courses/%d/enroll", cid), token, nil)
		fmt.Printf("  POST /courses/%d/enroll -> Status: %d (expected 201)\n", cid, st)
	}

	fmt.Println("Menunggu 20 detik agar mekanisme retry Asynq selesai...")
	time.Sleep(20 * time.Second)

	var notifCount int
	_ = pool.QueryRow(ctx, "SELECT COUNT(*) FROM notifications WHERE user_id = (SELECT id FROM users WHERE username = 'mhsuji')").Scan(&notifCount)

	_, resMine, _ := httpReq("GET", "/notifications/mine", token, nil)
	var apiCount int
	if list, ok := resMine["data"].([]any); ok {
		apiCount = len(list)
	}

	mailCount := countMailLog("mhsuji@")

	fmt.Println("\n--- HASIL UJI C.3 ---")
	fmt.Printf("1. Jumlah baris notifications di database: %d (expected 6)\n", notifCount)
	fmt.Printf("2. Jumlah notifikasi di GET /notifications/mine: %d (expected 6)\n", apiCount)
	fmt.Printf("3. Jumlah baris email di logs/mail.log: %d (expected 6)\n", mailCount)
	if notifCount == 6 && apiCount == 6 && mailCount == 6 {
		fmt.Println("[PASS] Ketiga angka SAMA-SAMA 6! Retried temporary failure sukses tanpa kehilangan notifikasi.")
	} else {
		fmt.Println("[FAIL] Angka belum sama!")
	}
}

// ---------------------------------------------------------------------
// C.4 Permanent failure
// ---------------------------------------------------------------------
func runC4(ctx context.Context, salahToken, adminToken string) {
	fmt.Println("==================================================")
	fmt.Println("MENJALANKAN UJI C.4: Permanent Failure (kampus.invalid)")
	fmt.Println("==================================================")
	_, resBefore, _ := httpReq("GET", "/admin/queue", adminToken, nil)
	var archivedBefore int
	if q, ok := resBefore["data"].(map[string]any)["queue"].(map[string]any); ok {
		archivedBefore = int(q["archived"].(float64))
	}
	fmt.Printf("queue.archived sebelum pendaftaran: %d\n", archivedBefore)

	// Daftar kelas 2 sebagai salahalamat
	status, _, _ := httpReq("POST", "/courses/2/enroll", salahToken, nil)
	fmt.Printf("POST /courses/2/enroll sebagai salahalamat -> Status: %d (expected 201)\n", status)

	fmt.Println("Menunggu 15 detik agar task diproses dan langsung masuk archived...")
	time.Sleep(15 * time.Second)

	_, resAfter, _ := httpReq("GET", "/admin/queue", adminToken, nil)
	dataAfter, _ := json.MarshalIndent(resAfter, "", "  ")
	fmt.Printf("GET /admin/queue setelah permanent failure:\n%s\n", string(dataAfter))
}

// ---------------------------------------------------------------------
// C.5 Idempotency
// ---------------------------------------------------------------------
func runC5Prep(ctx context.Context, token string) {
	fmt.Println("==================================================")
	fmt.Println("C.5 LANGKAH 1-2: Enroll kelas 8 saat worker mati")
	fmt.Println("==================================================")
	st, res, err := httpReq("POST", "/courses/8/enroll", token, nil)
	fmt.Printf("POST /courses/8/enroll -> Status: %d, Response: %v (err: %v)\n", st, res, err)

	var id int64
	var pubAt *time.Time
	err = pool.QueryRow(ctx, "SELECT id, published_at FROM outbox ORDER BY id DESC LIMIT 1").Scan(&id, &pubAt)
	if err != nil {
		log.Fatalf("get outbox id: %v", err)
	}
	fmt.Printf("Outbox event id: %d, published_at: %v (NULL = %v)\n", id, pubAt, pubAt == nil)
}

func runC5Check(ctx context.Context) {
	fmt.Println("==================================================")
	fmt.Println("C.5 LANGKAH 4: Verifikasi Idempotency setelah Worker Jalan")
	fmt.Println("==================================================")
	var id int64
	_ = pool.QueryRow(ctx, "SELECT id FROM outbox ORDER BY id DESC LIMIT 1").Scan(&id)

	var count int
	_ = pool.QueryRow(ctx, "SELECT COUNT(*) FROM notifications WHERE outbox_id = $1", id).Scan(&count)
	fmt.Printf("1. COUNT(*) FROM notifications WHERE outbox_id = %d: %d (expected 1)\n", id, count)

	mailCount := countMailLog("IF306")
	fmt.Printf("2. Jumlah baris logs/mail.log untuk kelas IF306: %d (expected 1)\n", mailCount)
}

// ---------------------------------------------------------------------
// C.6 Graceful shutdown
// ---------------------------------------------------------------------
func runC6(ctx context.Context, token string) {
	fmt.Println("==================================================")
	fmt.Println("MENJALANKAN UJI C.6: Graceful Shutdown")
	fmt.Println("==================================================")
	st, _, _ := httpReq("POST", "/courses/9/enroll", token, nil)
	fmt.Printf("POST /courses/9/enroll -> Status: %d (expected 201)\n", st)
}

// ---------------------------------------------------------------------
// D.2 Notifikasi Perubahan Kuota
// ---------------------------------------------------------------------
func runD2(ctx context.Context, adminToken string) {
	fmt.Println("==================================================")
	fmt.Println("MENJALANKAN UJI D.2: Notifikasi Perubahan Kuota")
	fmt.Println("==================================================")
	// Daftarkan 5 mahasiswa ke kelas 3 (Kecerdasan Buatan)
	courseID := 3
	fmt.Println("Mendaftarkan 5 mahasiswa ke kelas 3...")
	for i := 1; i <= 5; i++ {
		uname := fmt.Sprintf("mhs_kuota_%d", i)
		uemail := fmt.Sprintf("mhs_kuota_%d@student.unair.ac.id", i)
		uToken := ensureUser(ctx, uname, uemail, "user")
		st, _, _ := httpReq("POST", fmt.Sprintf("/courses/%d/enroll", courseID), uToken, nil)
		fmt.Printf("  Mendaftarkan %s -> status %d\n", uname, st)
	}

	// Ambil versi kelas 3 saat ini
	var version, quota int
	_ = pool.QueryRow(ctx, "SELECT version, quota FROM courses WHERE id = $1", courseID).Scan(&version, &quota)
	newQuota := quota + 10
	fmt.Printf("Mengubah kuota kelas 3 dari %d menjadi %d (version=%d)...\n", quota, newQuota, version)

	// Admin melakukan PATCH /courses/3
	st, res, err := httpReq("PATCH", fmt.Sprintf("/courses/%d", courseID), adminToken, map[string]any{
		"quota":   newQuota,
		"version": version,
	})
	fmt.Printf("PATCH /courses/3 -> Status: %d, Response: %v (err: %v)\n", st, res, err)

	fmt.Println("Menunggu 10 detik agar seluruh peserta menerima email notifikasi perubahan kuota...")
	time.Sleep(10 * time.Second)

	// Periksa tabel outbox untuk event course.quota_changed
	var outboxID int64
	var eventType string
	var payload []byte
	err = pool.QueryRow(ctx, "SELECT id, event_type, payload FROM outbox WHERE event_type = 'course.quota_changed' ORDER BY id DESC LIMIT 1").Scan(&outboxID, &eventType, &payload)
	if err != nil {
		fmt.Printf("[FAIL] Event course.quota_changed tidak ditemukan di outbox: %v\n", err)
		return
	}
	fmt.Printf("[OK] Outbox tercatat: id=%d event_type=%s payload=%s\n", outboxID, eventType, string(payload))

	// Periksa notifications untuk outbox_id tersebut
	rows, _ := pool.Query(ctx, "SELECT user_id, subject FROM notifications WHERE outbox_id = $1", outboxID)
	defer rows.Close()
	cnt := 0
	for rows.Next() {
		var uid int
		var sub string
		_ = rows.Scan(&uid, &sub)
		cnt++
		fmt.Printf("  Notifikasi %d: user_id=%d subject=%q\n", cnt, uid, sub)
	}
}

// ---------------------------------------------------------------------
// D.3 Percobaan Pembanding
// ---------------------------------------------------------------------
func runD3(ctx context.Context, token string) {
	fmt.Println("==================================================")
	fmt.Println("MENJALANKAN UJI D.3: Percobaan Pembanding Bug 4")
	fmt.Println("==================================================")
	runResetSQL()
	clearMailLog()

	// 1. Pendaftaran pertama (berhasil)
	st1, _, _ := httpReq("POST", "/courses/2/enroll", token, nil)
	fmt.Printf("Pendaftaran 1 (POST /courses/2/enroll) -> Status: %d\n", st1)

	// 2. Pendaftaran kedua yang sama (seharusnya gagal dan tidak meninggalkan jejak outbox)
	st2, res2, _ := httpReq("POST", "/courses/2/enroll", token, nil)
	fmt.Printf("Pendaftaran 2 (POST /courses/2/enroll diulang) -> Status: %d, Code: %v\n", st2, res2["code"])

	var outboxCount int
	_ = pool.QueryRow(ctx, "SELECT COUNT(*) FROM outbox WHERE (payload->>'user_id')::INT = (SELECT id FROM users WHERE username = 'mhsuji')").Scan(&outboxCount)
	fmt.Printf("Hasil Pemeriksaan Outbox:\n")
	fmt.Printf("Jumlah baris outbox mhsuji saat ini: %d\n", outboxCount)
	if outboxCount == 1 {
		fmt.Println("[KONDISI BENAR]: Baris outbox tepat 1. Pendaftaran gagal tidak meninggalkan catatan outbox palsu.")
	} else {
		fmt.Printf("[KONDISI BUG TERBUKTI]: Baris outbox ada %d! Pendaftaran yang di-rollback menyisakan event liar di outbox karena s.outbox.Insert tidak memakai WithTx(tx)!\n", outboxCount)
	}
}
