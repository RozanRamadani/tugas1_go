package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// LoadEnv membaca konfigurasi dari file .env.
func LoadEnv() {
	err := godotenv.Load()

	if err != nil {
		log.Println("file .env tidak ditemukan, menggunakan environment variable")
	}
}

// GetEnv mengambil nilai environment variable.
// Jika tidak ditemukan, gunakan nilai default.
func GetEnv(key string, defaultValue string) string {
	value := os.Getenv(key)

	if value == "" {
		return defaultValue
	}

	return value
}
