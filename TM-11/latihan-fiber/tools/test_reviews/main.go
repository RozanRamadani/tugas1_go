package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"latihan-fiber/helper"
)

const baseURL = "http://localhost:3000/api/v1"
const jwtSecret = "supersecretloadtestkey123456789012"

func main() {
	jwtManager := helper.NewJWTManager(jwtSecret, "latihan-fiber", 1*time.Hour, 24*time.Hour)

	// Token untuk masing-masing user
	tokenMhs1, _ := jwtManager.GenerateAccessTokenWithUsername(7, "mhs001", "user")
	tokenMhs2, _ := jwtManager.GenerateAccessTokenWithUsername(8, "mhs002", "user")
	tokenMhs3, _ := jwtManager.GenerateAccessTokenWithUsername(9, "mhs003", "user")
	tokenAdmin, _ := jwtManager.GenerateAccessTokenWithUsername(4, "admin", "admin")

	fmt.Println("==================================================")
	fmt.Println("PENGUJIAN ENDPOINT ULASAN KELAS (LANGKAH 9)")
	fmt.Println("==================================================")

	// 1. mhs003 menulis ulasan IF101 -> 403 NOT_ENROLLED
	res1, code1 := doReq("POST", "/courses/1/reviews", tokenMhs3, map[string]any{
		"rating":  5,
		"comment": "Mencoba menulis ulasan padahal belum terdaftar",
		"tags":    []string{"coba"},
	})
	fmt.Printf("[1] mhs003 menulis ulasan (tidak terdaftar) -> Status: %d, Response: %s\n", code1, res1)

	// 2. mhs001 menulis ulasan dengan tag [Tugas-Banyak, dosen jelas, tugas-banyak] -> 201
	res2, code2 := doReq("POST", "/courses/1/reviews", tokenMhs1, map[string]any{
		"rating":  5,
		"comment": "Materi Go-nya jelas, tugasnya cukup banyak",
		"tags":    []string{"Tugas-Banyak", " dosen jelas ", "tugas-banyak"},
	})
	fmt.Printf("[2] mhs001 menulis ulasan valid -> Status: %d, Response: %s\n", code2, res2)

	var reviewData struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	_ = json.Unmarshal([]byte(res2), &reviewData)
	reviewID := reviewData.Data.ID

	// 3. mhs001 menulis ulasan kedua di kelas yang sama -> 409 ALREADY_REVIEWED
	res3, code3 := doReq("POST", "/courses/1/reviews", tokenMhs1, map[string]any{
		"rating":  4,
		"comment": "Ulasan kedua saya yang seharusnya ditolak database",
	})
	fmt.Printf("[3] mhs001 menulis ulasan kedua (duplikat) -> Status: %d, Response: %s\n", code3, res3)

	// 4. Validasi: rating 9, komentar pendek, 6 tags -> 422
	res4, code4 := doReq("POST", "/courses/1/reviews", tokenMhs2, map[string]any{
		"rating":  9,
		"comment": "pendek",
		"tags":    []string{"1", "2", "3", "4", "5", "6"},
	})
	fmt.Printf("[4] Validasi gagal (rating 9, comment pendek, 6 tag) -> Status: %d, Response: %s\n", code4, res4)

	// 5. Ulasan untuk kelas 99999 -> 404
	res5, code5 := doReq("POST", "/courses/99999/reviews", tokenMhs1, map[string]any{
		"rating":  4,
		"comment": "Ulasan kelas yang tidak ada di PostgreSQL",
	})
	fmt.Printf("[5] Ulasan untuk kelas 99999 -> Status: %d, Response: %s\n", code5, res5)

	// 6. GET /reviews/bukan-objectid -> 400
	res6, code6 := doReq("GET", "/reviews/bukan-objectid", tokenMhs1, nil)
	fmt.Printf("[6] GET /reviews/bukan-objectid -> Status: %d, Response: %s\n", code6, res6)

	// 7. mhs002 mengubah ulasan mhs001 -> 403 Forbidden
	newRating := 3
	res7, code7 := doReq("PATCH", "/reviews/"+reviewID, tokenMhs2, map[string]any{
		"rating": &newRating,
	})
	fmt.Printf("[7] mhs002 mengubah ulasan mhs001 (bukan pemilik) -> Status: %d, Response: %s\n", code7, res7)

	// 8. mhs001 mengubah ulasannya sendiri -> 200
	res8, code8 := doReq("PATCH", "/reviews/"+reviewID, tokenMhs1, map[string]any{
		"rating": &newRating,
	})
	fmt.Printf("[8] mhs001 mengubah ulasan sendiri -> Status: %d, Response: %s\n", code8, res8)

	// 9. mhs001 menilai ulasannya sendiri -> 409 OWN_REVIEW
	res9, code9 := doReq("POST", "/reviews/"+reviewID+"/helpful", tokenMhs1, nil)
	fmt.Printf("[9] mhs001 menilai ulasan sendiri -> Status: %d, Response: %s\n", code9, res9)

	// 10. mhs002 menilai membantu -> 200, lalu mengulanginya -> 409 ALREADY_VOTED
	res10a, code10a := doReq("POST", "/reviews/"+reviewID+"/helpful", tokenMhs2, nil)
	fmt.Printf("[10a] mhs002 menilai ulasan (pertama) -> Status: %d, Response: %s\n", code10a, res10a)
	res10b, code10b := doReq("POST", "/reviews/"+reviewID+"/helpful", tokenMhs2, nil)
	fmt.Printf("[10b] mhs002 menilai ulasan (kedua) -> Status: %d, Response: %s\n", code10b, res10b)

	// 11. Balasan ulasan (Reply)
	res11, code11 := doReq("POST", "/reviews/"+reviewID+"/replies", tokenMhs2, map[string]any{
		"body": "Setuju, materinya sangat mendalam dan aplikatif",
	})
	fmt.Printf("[11] mhs002 membalas ulasan -> Status: %d, Response: %s\n", code11, res11)

	// 12. Ringkasan ulasan (Summary)
	res12, code12 := doReq("GET", "/courses/1/reviews/summary", tokenMhs1, nil)
	fmt.Printf("[12] GET /courses/1/reviews/summary -> Status: %d, Response: %s\n", code12, res12)

	// 13. Daftar ulasan (ListByCourse)
	res13, code13 := doReq("GET", "/courses/1/reviews?limit=5", tokenMhs1, nil)
	fmt.Printf("[13] GET /courses/1/reviews -> Status: %d, Response: %s\n", code13, res13)

	// 14. Uji Serentak: 1 mahasiswa menilai 20 kali serentak -> tepat 1x 200 dan 19x 409
	tokenMhs3Vote, _ := jwtManager.GenerateAccessTokenWithUsername(9, "mhs003", "user")
	var wg sync.WaitGroup
	var mu sync.Mutex
	statusCounts := make(map[int]int)
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, code := doReq("POST", "/reviews/"+reviewID+"/helpful", tokenMhs3Vote, map[string]any{})
			mu.Lock()
			statusCounts[code]++
			mu.Unlock()
		}()
	}
	wg.Wait()
	fmt.Printf("[14] Uji serentak (mhs003 menilai 20x serentak): %v (Harus tepat 1 kali 200 dan 19 kali 409)\n", statusCounts)

	// 15. NoSQL Injection test: Body {"comment": {"$ne": ""}} -> 400
	reqBadJSON, _ := http.NewRequest("POST", baseURL+"/courses/1/reviews", bytes.NewBufferString(`{"comment": {"$ne": ""}, "rating": 5}`))
	reqBadJSON.Header.Set("Authorization", "Bearer "+tokenMhs2)
	reqBadJSON.Header.Set("Content-Type", "application/json")
	respInj, _ := http.DefaultClient.Do(reqBadJSON)
	bodyInj, _ := io.ReadAll(respInj.Body)
	respInj.Body.Close()
	fmt.Printf("[15] NoSQL injection body {\"$ne\": \"\"} -> Status: %d, Response: %s\n", respInj.StatusCode, string(bodyInj))

	// 16. NoSQL query param: ?tag={"$ne":""} -> 200 teks biasa
	res16, code16 := doReq("GET", `/courses/1/reviews?tag={"$ne":""}`, tokenMhs1, nil)
	fmt.Printf("[16] GET /courses/1/reviews?tag={\"$ne\":\"\"} -> Status: %d, Response: %s\n", code16, res16)

	// 17. mhs003 menghapus ulasan mhs001 -> 403
	res17, code17 := doReq("DELETE", "/reviews/"+reviewID, tokenMhs3, nil)
	fmt.Printf("[17] mhs003 menghapus ulasan mhs001 -> Status: %d, Response: %s\n", code17, res17)

	// 18. Admin menghapus ulasan mhs001 (review:moderate) -> 204
	res18, code18 := doReq("DELETE", "/reviews/"+reviewID, tokenAdmin, nil)
	fmt.Printf("[18] Admin menghapus ulasan mhs001 -> Status: %d, Response: %s\n", code18, res18)

	// 19. Menghapus ulasan yang sudah terhapus -> 404
	res19, code19 := doReq("DELETE", "/reviews/"+reviewID, tokenAdmin, nil)
	fmt.Printf("[19] Menghapus ulasan yang sudah terhapus -> Status: %d, Response: %s\n", code19, res19)
}

func doReq(method, path, token string, bodyData any) (string, int) {
	var body io.Reader
	if bodyData != nil {
		b, _ := json.Marshal(bodyData)
		body = bytes.NewReader(b)
	} else if method == "POST" || method == "PATCH" || method == "PUT" {
		body = bytes.NewReader([]byte("{}"))
	}
	req, _ := http.NewRequest(method, baseURL+path, body)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err.Error(), 0
	}
	defer resp.Body.Close()
	respBytes, _ := io.ReadAll(resp.Body)
	return string(respBytes), resp.StatusCode
}
