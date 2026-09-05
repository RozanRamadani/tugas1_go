package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/requestid"

	"latihan-fiber/app/repository"
	"latihan-fiber/config"
	"latihan-fiber/database"
)

var metodeBerbody = map[string]bool{
	fiber.MethodPost:  true,
	fiber.MethodPut:   true,
	fiber.MethodPatch: true,
}

// requireJSON menolak request berisi body yang Content-Type-nya bukan JSON.
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

	// ============================================================
	// 1. LOAD CONFIGURATION
	// ============================================================

	config.LoadEnv()

	// ============================================================
	// 2. CONNECT TO DATABASE
	// ============================================================

	pool, err := database.NewPool(context.Background())

	if err != nil {
		log.Fatalf("database: %v", err)
	}

	// Pool ditutup ketika aplikasi berhenti.
	defer pool.Close()

	// ============================================================
	// 3. RAKIT DEPENDENCY
	// ============================================================
	//
	// pool
	//   ↓
	// repository
	//   ↓
	// handler

	userRepository := repository.NewUserRepository(pool)

	userHandler := NewUserHandler(userRepository)

	// ============================================================
	// 4. BUAT APLIKASI FIBER
	// ============================================================

	app := fiber.New(fiber.Config{

		AppName: "Praktikum Backend Lanjut - Pertemuan 3",

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

	// ============================================================
	// 5. GLOBAL MIDDLEWARE
	// ============================================================

	app.Use(requestid.New())

	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${locals:requestid} ${method} ${path} ${status} ${latency}\n",
	}))

	app.Use(cors.New())

	// ============================================================
	// 6. ROOT ENDPOINT
	// ============================================================

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	// ============================================================
	// 7. API GROUP
	// ============================================================

	api := app.Group("/api/v1")

	// ============================================================
	// 8. HEALTH CHECK
	// ============================================================
	//
	// Sekarang health tidak hanya mengecek server,
	// tetapi juga koneksi PostgreSQL.

	api.Get("/health", func(c *fiber.Ctx) error {

		ctx, cancel := context.WithTimeout(
			c.UserContext(),
			2*time.Second,
		)

		defer cancel()

		if err := pool.Ping(ctx); err != nil {

			return fail(
				c,
				fiber.StatusServiceUnavailable,
				"database tidak dapat dihubungi",
			)
		}

		return ok(
			c,
			"server dan database berjalan",
			nil,
		)
	})

	// ============================================================
	// 9. USER ROUTES
	// ============================================================

	// requireJSON hanya berlaku untuk endpoint users.
	u := api.Group("/users", requireJSON)

	u.Get("/", userHandler.List)

	u.Get("/:id", userHandler.Get)

	u.Post("/", userHandler.Create)

	u.Put("/:id", userHandler.Replace)

	u.Patch("/:id", userHandler.Patch)

	u.Delete("/:id", userHandler.Delete)

	// ============================================================
	// 10. ENDPOINT TIDAK DIKENAL
	// ============================================================

	app.Use(func(c *fiber.Ctx) error {

		return fail(
			c,
			fiber.StatusNotFound,
			"endpoint tidak ditemukan",
		)
	})

	// ============================================================
	// 11. JALANKAN SERVER
	// ============================================================

	port := config.GetEnv(
		"APP_PORT",
		"3000",
	)

	fmt.Println(
		"Server berjalan di http://localhost:" + port,
	)

	log.Fatal(
		app.Listen(":" + port),
	)
}
