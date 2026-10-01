package config

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DB adalah instance global dari koneksi database
var DB *gorm.DB

// ConnectDB menginisialisasi koneksi ke PostgreSQL
func ConnectDB() {
	host := GetEnv("DB_HOST", "localhost")
	port := GetEnv("DB_PORT", "5432")
	user := GetEnv("DB_USER", "postgres")
	password := GetEnv("DB_PASSWORD", "secret")
	dbname := GetEnv("DB_NAME", "siakad_mini")

	// DSN (Data Source Name)
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta",
		host, user, password, dbname, port)

	// Buka koneksi menggunakan GORM
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Ambil instance sql.DB untuk mengatur connection pool
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get generic database object: %v", err)
	}

	// Setup Connection Pool
	sqlDB.SetMaxIdleConns(10)           // Jumlah maksimum koneksi yang idle (nganggur)
	sqlDB.SetMaxOpenConns(100)          // Jumlah maksimum koneksi yang terbuka secara bersamaan
	sqlDB.SetConnMaxLifetime(time.Hour) // Durasi maksimal sebuah koneksi bisa dipakai ulang

	DB = db
	log.Println("Database connection successfully established")
}
