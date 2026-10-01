package main

import (
	"log"
	"siakad-mini/config"

	"github.com/gofiber/fiber/v2"
)

func main() {
	// 1. Inisialisasi Environment Variables
	config.LoadConfig()

	// 2. Inisialisasi Koneksi Database
	config.ConnectDB()

	app := fiber.New()

	// Endpoint /health untuk mengecek status server dan database
	app.Get("/health", func(c *fiber.Ctx) error {
		dbStatus := "down"
		
		// Ambil instance generic db untuk cek ping
		if sqlDB, err := config.DB.DB(); err == nil {
			if err := sqlDB.Ping(); err == nil {
				dbStatus = "up"
			}
		}

		return c.JSON(fiber.Map{
			"status":   "success",
			"message":  "SIAKAD Mini Server is up and running",
			"database": dbStatus,
		})
	})

	port := config.GetEnv("APP_PORT", "3000")
	log.Println("Server is running on port " + port)
	
	// Menjalankan server
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
