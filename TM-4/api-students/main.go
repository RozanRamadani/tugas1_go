package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"

	"api-students/app/handler"
	"api-students/app/repository"
	"api-students/config"
	"api-students/database"
)

func main() {

	// ============================================================
	// 1. Load environment variable dari .env
	// ============================================================

	config.LoadEnv()

	// ============================================================
	// 2. Buat connection pool PostgreSQL
	// ============================================================

	ctx := context.Background()

	pool, err := database.NewPool(ctx)

	if err != nil {
		log.Fatalf("gagal terhubung ke database: %v", err)
	}

	defer pool.Close()

	log.Println("database berhasil terhubung")

	// ============================================================
	// 3. Buat repository
	// ============================================================

	studentRepo := repository.NewStudentRepository(pool)

	// ============================================================
	// 4. Buat handler
	// ============================================================

	studentHandler := handler.NewStudentHandler(studentRepo)

	// ============================================================
	// 5. Buat Fiber
	// ============================================================

	app := fiber.New(fiber.Config{
		AppName: "Praktikum Backend Lanjut - Students",

		ErrorHandler: func(c *fiber.Ctx, err error) error {

			status := fiber.StatusInternalServerError
			message := "terjadi kesalahan pada server"

			if fiberErr, ok := err.(*fiber.Error); ok {
				status = fiberErr.Code
				message = fiberErr.Message
			}

			return handler.Fail(
				c,
				status,
				message,
			)
		},
	})

	// ============================================================
	// 6. Health check
	// ============================================================

	app.Get("/api/v1/health", func(c *fiber.Ctx) error {

		// Health tidak hanya mengecek server,
		// tetapi juga memastikan database masih bisa diakses.

		pingCtx, cancel := context.WithTimeout(
			context.Background(),
			2*time.Second,
		)

		defer cancel()

		if err := pool.Ping(pingCtx); err != nil {

			return handler.Fail(
				c,
				fiber.StatusServiceUnavailable,
				"database tidak tersedia",
			)
		}

		return handler.Ok(
			c,
			"server dan database berjalan",
			nil,
		)
	})

	// ============================================================
	// 7. Route students
	// ============================================================

	students := app.Group("/api/v1/students")

	students.Get("/", studentHandler.FindAll)
	students.Get("/:id", studentHandler.FindByID)

	students.Post("/", studentHandler.Create)

	students.Put("/:id", studentHandler.Replace)

	students.Patch("/:id", studentHandler.Patch)

	students.Delete("/:id", studentHandler.Delete)

	// ============================================================
	// 8. Endpoint yang tidak ditemukan
	// ============================================================

	app.Use(func(c *fiber.Ctx) error {

		return handler.Fail(
			c,
			fiber.StatusNotFound,
			"endpoint tidak ditemukan",
		)
	})

	// ============================================================
	// 9. Jalankan server
	// ============================================================

	port := os.Getenv("APP_PORT")

	if port == "" {
		port = "3000"
	}

	log.Printf(
		"server berjalan di http://localhost:%s",
		port,
	)

	log.Fatal(
		app.Listen(":" + port),
	)
}
