package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// LoadConfig memuat environment variables dari file .env
func LoadConfig() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found, will use system environment variables")
	}
}

// GetEnv mengambil nilai environment variable, atau mengembalikan fallback jika kosong
func GetEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
