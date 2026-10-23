package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"latihan-fiber/config"
	"latihan-fiber/database"
	"latihan-fiber/helper"
)

func main() {
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

	// Buat token mhs001 (role user) dan admin (id 1, role admin)
	var mhsID int
	_ = pool.QueryRow(ctx, "SELECT id FROM users WHERE username = 'mhs001'").Scan(&mhsID)
	mhsToken, _ := jwtManager.GenerateAccessToken(mhsID, "user")

	var adminID int
	_ = pool.QueryRow(ctx, "SELECT id FROM users WHERE role = 'admin' LIMIT 1").Scan(&adminID)
	if adminID == 0 {
		adminID = 1
	}
	adminToken, _ := jwtManager.GenerateAccessToken(adminID, "admin")

	var if301ID, if302ID, if399ID int
	_ = pool.QueryRow(ctx, "SELECT id FROM courses WHERE code = 'IF301'").Scan(&if301ID)
	_ = pool.QueryRow(ctx, "SELECT id FROM courses WHERE code = 'IF302'").Scan(&if302ID)
	_ = pool.QueryRow(ctx, "SELECT id FROM courses WHERE code = 'IF399'").Scan(&if399ID)

	client := &http.Client{Timeout: 5 * time.Second}
	baseURL := "http://localhost:3000/api/v1"

	// Reset dulu
	_, _ = pool.Exec(ctx, "DELETE FROM enrollments WHERE user_id = $1", mhsID)
	_, _ = pool.Exec(ctx, "UPDATE courses SET enrolled_count = (SELECT COUNT(*) FROM enrollments WHERE course_id = courses.id)")

	fmt.Println("--- MEMULAI PENGUJIAN FUNGSIONAL C.4 ---")

	// 1. Daftar ke kelas yang masih kosong -> 201
	status, code, _ := doReq(client, http.MethodPost, fmt.Sprintf("%s/courses/%d/enroll", baseURL, if301ID), mhsToken, nil)
	fmt.Printf("1. Daftar kelas kosong: Status=%d (Harus 201)\n", status)

	// 2. Daftar ke kelas yang sama sekali lagi -> 409 ALREADY_ENROLLED
	status, code, _ = doReq(client, http.MethodPost, fmt.Sprintf("%s/courses/%d/enroll", baseURL, if301ID), mhsToken, nil)
	fmt.Printf("2. Daftar kelas ganda: Status=%d Code=%s (Harus 409 ALREADY_ENROLLED)\n", status, code)

	// 3. Batalkan kelas yang tidak pernah diambil -> 409 NOT_ENROLLED
	status, code, _ = doReq(client, http.MethodDelete, fmt.Sprintf("%s/courses/%d/enroll", baseURL, if302ID), mhsToken, nil)
	fmt.Printf("3. Drop kelas belum diambil: Status=%d Code=%s (Harus 409 NOT_ENROLLED)\n", status, code)

	// 4. Pindah dengan from_course_id sama dengan to_course_id -> 422 VALIDATION_ERROR
	status, code, _ = doReq(client, http.MethodPost, fmt.Sprintf("%s/enrollments/swap", baseURL), mhsToken, map[string]int{
		"from_course_id": if301ID,
		"to_course_id":   if301ID,
	})
	fmt.Printf("4. Swap ID sama: Status=%d Code=%s (Harus 422 VALIDATION_ERROR)\n", status, code)

	// 5. Pindah ke kelas yang masih kosong -> 200 OK
	status, code, _ = doReq(client, http.MethodPost, fmt.Sprintf("%s/enrollments/swap", baseURL), mhsToken, map[string]int{
		"from_course_id": if301ID,
		"to_course_id":   if302ID,
	})
	fmt.Printf("5. Swap kelas sah: Status=%d (Harus 200)\n", status)

	// 6. Mahasiswa mencoba ubah kuota -> 403 FORBIDDEN
	status, code, _ = doReq(client, http.MethodPatch, fmt.Sprintf("%s/courses/%d", baseURL, if301ID), mhsToken, map[string]int{
		"quota":   50,
		"version": 1,
	})
	fmt.Printf("6. Mhs ubah kuota: Status=%d (Harus 403)\n", status)

	// Ambil version IF301 saat ini
	var ver301 int
	_ = pool.QueryRow(ctx, "SELECT version FROM courses WHERE id = $1", if301ID).Scan(&ver301)

	// 7. Admin ubah kuota dengan version terbaru -> 200 OK; version naik 1
	status, code, body := doReq(client, http.MethodPatch, fmt.Sprintf("%s/courses/%d", baseURL, if301ID), adminToken, map[string]int{
		"quota":   50,
		"version": ver301,
	})
	var resMap map[string]any
	_ = json.Unmarshal([]byte(body), &resMap)
	fmt.Printf("7. Admin ubah kuota valid: Status=%d Res=%v (Harus 200, version naik)\n", status, body)

	// 8. Admin ubah kuota dengan version lama -> 409 VERSION_CONFLICT
	status, code, _ = doReq(client, http.MethodPatch, fmt.Sprintf("%s/courses/%d", baseURL, if301ID), adminToken, map[string]int{
		"quota":   55,
		"version": ver301, // version lama!
	})
	fmt.Printf("8. Admin version lama: Status=%d Code=%s (Harus 409 VERSION_CONFLICT)\n", status, code)

	// 9. Admin menurunkan kuota di bawah jumlah peserta -> 409 QUOTA_BELOW_ENROLLED
	// IF302 sekarang punya 1 peserta (mhs001). Ubah kuota IF302 menjadi 0 (atau kurang dari enrolled_count)
	var ver302, enc302 int
	_ = pool.QueryRow(ctx, "SELECT version, enrolled_count FROM courses WHERE id = $1", if302ID).Scan(&ver302, &enc302)
	status, code, _ = doReq(client, http.MethodPatch, fmt.Sprintf("%s/courses/%d", baseURL, if302ID), adminToken, map[string]int{
		"quota":   enc302 - 1, // di bawah peserta! Tapi validator min=1, jadi jika enc302 >= 2 kita coba.
		"version": ver302,
	})
	// Jika enc302=1, min=1, kita bisa tambahkan 5 peserta dulu ke IF302, lalu coba set kuota=2
	_, _ = pool.Exec(ctx, "UPDATE courses SET enrolled_count = 5 WHERE id = $1", if302ID)
	status, code, _ = doReq(client, http.MethodPatch, fmt.Sprintf("%s/courses/%d", baseURL, if302ID), adminToken, map[string]int{
		"quota":   3,
		"version": ver302,
	})
	fmt.Printf("9. Admin kuota di bawah peserta: Status=%d Code=%s (Harus 409 QUOTA_BELOW_ENROLLED)\n", status, code)

	fmt.Println("--- SEMUA PENGUJIAN C.4 SELESAI ---")
}

func doReq(client *http.Client, method, url, token string, bodyData any) (int, string, string) {
	var bodyReader io.Reader
	if bodyData != nil {
		b, _ := json.Marshal(bodyData)
		bodyReader = bytes.NewBuffer(b)
	} else {
		bodyReader = bytes.NewBuffer([]byte("{}"))
	}
	req, _ := http.NewRequest(method, url, bodyReader)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return 0, "ERR", err.Error()
	}
	defer resp.Body.Close()

	b, _ := io.ReadAll(resp.Body)
	var errResp config.ErrorResponse
	_ = json.Unmarshal(b, &errResp)

	return resp.StatusCode, errResp.Code, string(b)
}
