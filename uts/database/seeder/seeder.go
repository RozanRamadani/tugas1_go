package main

import (
	"fmt"
	"log"

	"siakad-mini/config"
	"siakad-mini/domain"

	"golang.org/x/crypto/bcrypt"
)

// hashPassword menggunakan algoritma bcrypt untuk mengenkripsi password plain-text
func hashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Gagal melakukan hashing password: %v", err)
	}
	return string(bytes)
}

func main() {
	config.LoadConfig()
	config.ConnectDB()
	db := config.DB

	log.Println("Memulai proses seeding database...")

	// ==========================================
	// 1. Buat Data Admin (Minimal 1)
	// ==========================================
	var adminCount int64
	db.Model(&domain.User{}).Where("role = ?", domain.RoleAdmin).Count(&adminCount)
	if adminCount == 0 {
		admin := domain.User{
			Username: "admin_super",
			Password: hashPassword("admin123"),
			Role:     domain.RoleAdmin,
		}
		db.Create(&admin)
		log.Println("✅ 1 Admin berhasil dibuat (Username: admin_super, Pass: admin123)")
	} else {
		log.Println("⚠️ Admin sudah ada, skip pembuatan admin.")
	}

	// ==========================================
	// 2. Buat Data Mahasiswa (Minimal 20)
	// ==========================================
	var studentCount int64
	db.Model(&domain.Student{}).Count(&studentCount)
	if studentCount == 0 {
		// Hashing password default mahasiswa (hanya di-hash sekali untuk optimalisasi waktu)
		defaultPass := hashPassword("mhs123")
		
		var users []domain.User
		var students []domain.Student

		for i := 1; i <= 20; i++ {
			nim := fmt.Sprintf("112233%02d", i) // Contoh NIM: 11223301 s/d 11223320
			
			// Siapkan entitas User
			user := domain.User{
				Username: nim, // Biasanya username mahasiswa menggunakan NIM
				Password: defaultPass,
				Role:     domain.RoleMahasiswa,
			}
			users = append(users, user)
		}

		// Insert batch untuk users agar cepat
		db.Create(&users)

		// Setelah user di-insert, ID-nya akan otomatis terisi. Sekarang buat profil mahasiswa.
		namaRealistis := []string{
			"Budi Santoso", "Siti Aminah", "Andi Wijaya", "Rina Marlina", "Dewi Lestari",
			"Agus Setiawan", "Ayu Wandira", "Reza Rahadian", "Dina Fitriani", "Fajar Nugraha",
			"Eka Putra", "Maya Sari", "Rizky Ramadhan", "Dian Sastro", "Hendra Gunawan",
			"Putri Handayani", "Iqbal Ramadhan", "Tari Haryanti", "Wahyu Hidayat", "Anita Rachman",
		}

		for i, user := range users {
			student := domain.Student{
				NIM:    user.Username,
				Name:   namaRealistis[i],
				UserID: user.ID,
			}
			students = append(students, student)
		}

		db.Create(&students) // Batch insert
		log.Println("✅ 20 Mahasiswa berhasil dibuat (Password default: mhs123)")
	} else {
		log.Println("⚠️ Mahasiswa sudah ada, skip pembuatan mahasiswa.")
	}

	// ==========================================
	// 3. Buat Data Mata Kuliah (Minimal 10)
	// ==========================================
	var courseCount int64
	db.Model(&domain.Course{}).Count(&courseCount)
	if courseCount == 0 {
		courses := []domain.Course{
			{Code: "IF101", Name: "Algoritma dan Pemrograman", Credits: 3},
			{Code: "IF102", Name: "Struktur Data", Credits: 3},
			{Code: "IF103", Name: "Basis Data", Credits: 4},
			{Code: "IF104", Name: "Pemrograman Web Dasar", Credits: 3},
			{Code: "IF105", Name: "Pemrograman Backend Lanjut", Credits: 4},
			{Code: "IF106", Name: "Kecerdasan Buatan", Credits: 3},
			{Code: "IF107", Name: "Jaringan Komputer", Credits: 3},
			{Code: "IF108", Name: "Sistem Operasi", Credits: 3},
			{Code: "IF109", Name: "Keamanan Informasi", Credits: 3},
			{Code: "IF110", Name: "Rekayasa Perangkat Lunak", Credits: 4},
		}
		db.Create(&courses)
		log.Println("✅ 10 Mata Kuliah berhasil dibuat")
	} else {
		log.Println("⚠️ Mata kuliah sudah ada, skip pembuatan mata kuliah.")
	}

	log.Println("🎉 Proses seeding selesai!")
}
