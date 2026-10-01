package main

import (
	"log"
	"siakad-mini/config"
	"siakad-mini/handler"
	"siakad-mini/repository"
	"siakad-mini/service"

	"github.com/gofiber/fiber/v2"
)

func main() {
	// 1. Inisialisasi Environment Variables & Koneksi Database
	config.LoadConfig()
	config.ConnectDB()

	// 2. Setup Dependency Injection (DI)
	userRepo := repository.NewUserRepository(config.DB)
	authService := service.NewAuthService(userRepo)
	authHandler := handler.NewAuthHandler(authService)

	// 3. Setup Framework Fiber
	app := fiber.New()

	// Endpoint /health untuk mengecek status server dan database
	app.Get("/health", func(c *fiber.Ctx) error {
		dbStatus := "down"
		
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

	// 4. Setup Routing API
	api := app.Group("/api/v1")
	authGroup := api.Group("/auth")
	authGroup.Post("/login", authHandler.Login)

	// 5. Jalankan server
	port := config.GetEnv("APP_PORT", "3000")
	log.Println("Server is running on port " + port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
