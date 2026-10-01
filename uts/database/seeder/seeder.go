package main

import (
	"fmt"
	"log"
	"math/rand"

	"siakad-mini/config"
	"siakad-mini/domain"

	"golang.org/x/crypto/bcrypt"
)

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

	// 1. Buat Data Admin
	var adminCount int64
	db.Model(&domain.User{}).Where("role = ?", domain.RoleAdmin).Count(&adminCount)
	if adminCount == 0 {
		admin := domain.User{
			Email:    "admin@siakad.com",
			Password: hashPassword("admin123"),
			Role:     domain.RoleAdmin,
		}
		db.Create(&admin)
		log.Println("✅ 1 Admin berhasil dibuat (Email: admin@siakad.com, Pass: admin123)")
	}

	// 2. Buat Data Mahasiswa
	var studentCount int64
	db.Model(&domain.Student{}).Count(&studentCount)
	if studentCount == 0 {
		defaultPass := hashPassword("mhs12345") // minimal 8 karakter sesuai requirement UTS
		
		var users []domain.User
		var students []domain.Student
		prodiList := []string{"Teknik Informatika", "Sistem Informasi", "Ilmu Komputer"}

		for i := 1; i <= 20; i++ {
			email := fmt.Sprintf("mhs%02d@student.com", i)
			
			user := domain.User{
				Email:    email,
				Password: defaultPass,
				Role:     domain.RoleMahasiswa,
			}
			users = append(users, user)
		}
		db.Create(&users)

		namaRealistis := []string{
			"Budi Santoso", "Siti Aminah", "Andi Wijaya", "Rina Marlina", "Dewi Lestari",
			"Agus Setiawan", "Ayu Wandira", "Reza Rahadian", "Dina Fitriani", "Fajar Nugraha",
			"Eka Putra", "Maya Sari", "Rizky Ramadhan", "Dian Sastro", "Hendra Gunawan",
			"Putri Handayani", "Iqbal Ramadhan", "Tari Haryanti", "Wahyu Hidayat", "Anita Rachman",
		}

		for i, user := range users {
			student := domain.Student{
				UserID:      user.ID,
				NIM:         fmt.Sprintf("112233%02d", i+1),
				Nama:        namaRealistis[i],
				Prodi:       prodiList[i%3],
				Angkatan:    2022 + (i % 2), // 2022 atau 2023
				IpkTerakhir: 3.0 + (rand.Float64() * 1.0), // IPK antara 3.0 - 4.0
			}
			students = append(students, student)
		}
		db.Create(&students)
		log.Println("✅ 20 Mahasiswa berhasil dibuat (Password default: mhs12345)")
	}

	// 3. Buat Data Mata Kuliah
	var courseCount int64
	db.Model(&domain.Course{}).Count(&courseCount)
	if courseCount == 0 {
		courses := []domain.Course{
			{KodeMk: "IF101", NamaMk: "Algoritma dan Pemrograman", Sks: 3, Semester: 1, Kuota: 40},
			{KodeMk: "IF102", NamaMk: "Struktur Data", Sks: 3, Semester: 2, Kuota: 40},
			{KodeMk: "IF103", NamaMk: "Basis Data", Sks: 4, Semester: 3, Kuota: 40},
			{KodeMk: "IF104", NamaMk: "Pemrograman Web Dasar", Sks: 3, Semester: 3, Kuota: 40},
			{KodeMk: "IF105", NamaMk: "Pemrograman Backend Lanjut", Sks: 4, Semester: 5, Kuota: 30},
			{KodeMk: "IF106", NamaMk: "Kecerdasan Buatan", Sks: 3, Semester: 5, Kuota: 40},
			{KodeMk: "IF107", NamaMk: "Jaringan Komputer", Sks: 3, Semester: 4, Kuota: 40},
			{KodeMk: "IF108", NamaMk: "Sistem Operasi", Sks: 3, Semester: 4, Kuota: 40},
			{KodeMk: "IF109", NamaMk: "Keamanan Informasi", Sks: 3, Semester: 6, Kuota: 35},
			{KodeMk: "IF110", NamaMk: "Rekayasa Perangkat Lunak", Sks: 4, Semester: 6, Kuota: 35},
		}
		db.Create(&courses)
		log.Println("✅ 10 Mata Kuliah berhasil dibuat")
	}

	log.Println("🎉 Proses seeding selesai!")
}
