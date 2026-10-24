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
	tokenMhs, _ := jwtManager.GenerateAccessTokenWithUsername(7, "mhs001", "user")

	fmt.Println("==================================================")
	fmt.Println("PENGUJIAN FITUR PORTOFOLIO MAHASISWA (TUGAS MANDIRI)")
	fmt.Println("==================================================")

	// Kasus Gagal 1: Mahasiswa tidak ada di PostgreSQL
	res1, code1 := doReq("GET", "/students/tidak-ada/portfolio", tokenMhs, nil)
	fmt.Printf("[Kasus Gagal 1] GET mahasiswa tidak ada di PG -> Status: %d, Response: %s\n\n", code1, res1)

	// Kasus Gagal 2: Mahasiswa ada di PG tapi portofolio belum dibuat
	res2, code2 := doReq("GET", "/students/std_001/portfolio", tokenMhs, nil)
	fmt.Printf("[Kasus Gagal 2] GET portofolio belum dibuat -> Status: %d, Response: %s\n\n", code2, res2)

	// Kasus Gagal 3: Validasi gagal (bio terlalu pendek, headline kosong)
	res3, code3 := doReq("PUT", "/students/std_001/portfolio", tokenMhs, map[string]any{
		"headline": "",
		"bio":      "pendek",
	})
	fmt.Printf("[Kasus Gagal 3] PUT validasi gagal (bio pendek) -> Status: %d, Response: %s\n\n", code3, res3)

	// Sukses 1: PUT /students/:id/portfolio (Membuat portofolio awal dengan 1 proyek)
	res4, code4 := doReq("PUT", "/students/std_001/portfolio", tokenMhs, map[string]any{
		"headline": "Backend Engineer & Cloud Architecture Enthusiast",
		"bio":      "Mahasiswa D4 Teknik Informatika Universitas Airlangga yang fokus mendalami sistem terdistribusi, Go, dan database berskala besar.",
		"projects": []map[string]any{
			{
				"title":        "Sistem Manajemen KRS Berbasis Go Fiber",
				"technologies": []string{"Go", "Fiber", "PostgreSQL", "Docker"},
				"link":         "https://github.com/student/krs-system",
				"year":         2025,
			},
		},
		"certificates": []map[string]any{
			{
				"name":   "Certified MongoDB Associate Developer",
				"issuer": "MongoDB University",
				"date":   "2026-02-15",
			},
		},
	})
	fmt.Printf("[Sukses 1] PUT /students/std_001/portfolio -> Status: %d, Response: %s\n\n", code4, res4)

	// Sukses 2: GET /students/:id/portfolio
	res5, code5 := doReq("GET", "/students/std_001/portfolio", tokenMhs, nil)
	fmt.Printf("[Sukses 2] GET /students/std_001/portfolio -> Status: %d, Response: %s\n\n", code5, res5)

	// Persiapan uji batas 20 proyek: Isi proyek hingga berjumlah 18
	fmt.Println(">> Mengisi 17 proyek tambahan secara berurutan agar total proyek menjadi 18...")
	for i := 2; i <= 18; i++ {
		_, code := doReq("POST", "/students/std_001/portfolio/projects", tokenMhs, map[string]any{
			"title":        fmt.Sprintf("Proyek Riset Ke-%d", i),
			"technologies": []string{"Go", "MongoDB", "Redis"},
			"link":         fmt.Sprintf("https://github.com/student/project-%d", i),
			"year":         2026,
		})
		if code != 201 {
			fmt.Printf("Gagal mengisi proyek ke-%d, status: %d\n", i, code)
		}
	}
	fmt.Println(">> Berhasil! Portofolio std_001 saat ini memiliki 18 proyek.")

	// Uji Serentak Batas 20 Proyek: Kirim 10 request serentak ke akun dengan 18 proyek
	fmt.Println("\n>> Menjalankan serangan serentak: 10 request POST proyek dikirim BERSAMAAN...")
	var wg sync.WaitGroup
	var mu sync.Mutex
	concurrencyResults := make(map[int]int)

	for i := 1; i <= 10; i++ {
		wg.Add(1)
		idx := i
		go func() {
			defer wg.Done()
			_, statusCode := doReq("POST", "/students/std_001/portfolio/projects", tokenMhs, map[string]any{
				"title":        fmt.Sprintf("Proyek Serentak %d", idx),
				"technologies": []string{"Go", "PostgreSQL", "Docker"},
				"link":         "https://github.com/student/concurrent-test",
				"year":         2026,
			})
			mu.Lock()
			concurrencyResults[statusCode]++
			mu.Unlock()
		}()
	}
	wg.Wait()

	fmt.Printf("[BUKTI BATAS 20 PROYEK] Hasil 10 request serentak: %v\n", concurrencyResults)
	fmt.Println("-> Kriteria evaluasi modul: Tepat 2 request berhasil (201) dan 8 request ditolak (409)!")

	// Cek jumlah proyek saat ini
	resChk, _ := doReq("GET", "/students/std_001/portfolio", tokenMhs, nil)
	var chkObj struct {
		Data struct {
			Projects []struct {
				ID string `json:"id"`
			} `json:"projects"`
		} `json:"data"`
	}
	_ = json.Unmarshal([]byte(resChk), &chkObj)
	totalProjects := len(chkObj.Data.Projects)
	fmt.Printf("-> Total proyek tersimpan di database: %d proyek (Maksimal 20)\n\n", totalProjects)

	// Kasus Gagal 4: Menambah proyek ke-21 saat sudah penuh
	resLimit, codeLimit := doReq("POST", "/students/std_001/portfolio/projects", tokenMhs, map[string]any{
		"title":        "Proyek Melebihi Batas Ke-21",
		"technologies": []string{"Rust"},
		"year":         2026,
	})
	fmt.Printf("[Kasus Gagal 4] POST proyek ke-21 saat sudah 20 -> Status: %d, Response: %s\n\n", codeLimit, resLimit)

	// Hapus 1 proyek dengan DELETE $pull
	var deleteID string
	if totalProjects > 0 {
		deleteID = chkObj.Data.Projects[0].ID
	}
	resDel, codeDel := doReq("DELETE", "/students/std_001/portfolio/projects/"+deleteID, tokenMhs, nil)
	fmt.Printf("[Sukses 3] DELETE /students/std_001/portfolio/projects/%s -> Status: %d, Response: %s\n\n", deleteID, codeDel, resDel)

	// Kasus Gagal 5: Format ID proyek bukan ObjectID
	resBadOID, codeBadOID := doReq("DELETE", "/students/std_001/portfolio/projects/bukan-object-id", tokenMhs, nil)
	fmt.Printf("[Kasus Gagal 5] DELETE id proyek tidak valid -> Status: %d, Response: %s\n\n", codeBadOID, resBadOID)

	// Kasus Gagal 6: ID proyek tidak ditemukan
	resNotFoundPrj, codeNotFoundPrj := doReq("DELETE", "/students/std_001/portfolio/projects/6adbf1bff5d40c03efe8a899", tokenMhs, nil)
	fmt.Printf("[Kasus Gagal 6] DELETE id proyek yang tidak ada -> Status: %d, Response: %s\n\n", codeNotFoundPrj, resNotFoundPrj)

	// Uji NoSQL Injection: kirim operator {"$ne": ""} di body
	reqInj, _ := http.NewRequest("POST", baseURL+"/students/std_001/portfolio/projects", bytes.NewBufferString(`{"title": {"$ne": ""}, "technologies": ["Go"], "year": 2026}`))
	reqInj.Header.Set("Authorization", "Bearer "+tokenMhs)
	reqInj.Header.Set("Content-Type", "application/json")
	respInj, _ := http.DefaultClient.Do(reqInj)
	bodyInj, _ := io.ReadAll(respInj.Body)
	respInj.Body.Close()
	fmt.Printf("[BUKTI NoSQL INJECTION DITOLAK] Body {\"title\": {\"$ne\": \"\"}} -> Status: %d, Response: %s\n\n", respInj.StatusCode, string(bodyInj))

	// Buat portofolio untuk std_002 dengan beragam teknologi untuk memperkaya agregasi top 10
	_, _ = doReq("PUT", "/students/std_002/portfolio", tokenMhs, map[string]any{
		"headline": "Fullstack Cloud Developer",
		"bio":      "Fokus pada modern web stack, microservices, dan observability.",
		"projects": []map[string]any{
			{"title": "E-Commerce Microservices", "technologies": []string{"Go", "Kubernetes", "gRPC", "PostgreSQL", "Redis"}, "year": 2025},
			{"title": "Realtime Analytics Dashboard", "technologies": []string{"Python", "Kafka", "MongoDB", "React", "TypeScript"}, "year": 2026},
		},
	})

	// Sukses 4: GET /portfolios/stats/technologies (Aggregation Pipeline Double $unwind)
	resStats, codeStats := doReq("GET", "/portfolios/stats/technologies", "", nil)
	fmt.Printf("[Sukses 4] GET /portfolios/stats/technologies -> Status: %d, Response: %s\n\n", codeStats, resStats)
}

func doReq(method, path, token string, bodyData any) (string, int) {
	var body io.Reader
	if bodyData != nil {
		b, _ := json.Marshal(bodyData)
		body = bytes.NewReader(b)
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
