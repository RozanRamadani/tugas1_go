package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

// Method yang wajib menggunakan JSON body.
var metodeBerbody = map[string]bool{
	fiber.MethodPost:  true,
	fiber.MethodPut:   true,
	fiber.MethodPatch: true,
}

// Tolak POST/PUT/PATCH jika Content-Type bukan JSON.
func requireJSON(c *fiber.Ctx) error {
	if metodeBerbody[c.Method()] {
		ct := c.Get("Content-Type")

		if !strings.HasPrefix(ct, fiber.MIMEApplicationJSON) {
			return fail(
				c,
				fiber.StatusUnsupportedMediaType,
				"Content-Type harus application/json",
			)
		}
	}

	return c.Next()
}

func main() {

	// Membuat aplikasi Fiber.
	app := fiber.New(fiber.Config{
		AppName: "API Students - Praktikum Backend",

		// Error yang tidak ditangani endpoint masuk ke sini.
		ErrorHandler: func(c *fiber.Ctx, err error) error {

			status := fiber.StatusInternalServerError
			pesan := "terjadi kesalahan pada server"

			if e, ok := err.(*fiber.Error); ok {
				status = e.Code
				pesan = e.Message
			}

			return fail(c, status, pesan)
		},
	})

	// Middleware global.
	app.Use(requestid.New())

	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${locals:requestid} ${method} ${path} ${status} ${latency}\n",
	}))

	app.Use(cors.New())

	// Endpoint utama.
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("API Students berjalan")
	})

	// Semua endpoint API memakai /api/v1.
	api := app.Group("/api/v1")

	// Health check.
	api.Get("/health", func(c *fiber.Ctx) error {
		return ok(
			c,
			"server berjalan",
			fiber.Map{
				"timestamp": time.Now(),
			},
		)
	})

	// requireJSON hanya berlaku untuk /students.
	s := api.Group("/students", requireJSON)

	// CRUD Student.
	s.Get("/", listStudents)
	s.Get("/:id", getStudent)
	s.Post("/", createStudent)
	s.Put("/:id", replaceStudent)
	s.Patch("/:id", patchStudent)
	s.Delete("/:id", deleteStudent)

	// Endpoint yang tidak terdaftar → 404.
	app.Use(func(c *fiber.Ctx) error {
		return fail(
			c,
			fiber.StatusNotFound,
			"endpoint tidak ditemukan",
		)
	})

	fmt.Println("Server berjalan di http://localhost:3000")

	// Jalankan semua file dalam package main.
	log.Fatal(app.Listen(":3000"))
}
