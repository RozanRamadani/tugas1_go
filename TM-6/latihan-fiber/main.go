package main

import (
	"context"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/requestid"

	"latihan-fiber/app/repository"
	"latihan-fiber/config"
	"latihan-fiber/database"
	"latihan-fiber/helper"
	"latihan-fiber/route"
)

func main() {
	// ============================================================
	// 1. LOAD CONFIGURATION & DATABASE
	// ============================================================
	config.LoadEnv()

	pool, err := database.NewPool(context.Background())
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	// ============================================================
	// 2. RAKIT DEPENDENCY & LOAD PERMISSION (LANGKAH 4)
	// ============================================================
	roleRepository := repository.NewRoleRepository(pool)

	rawPermissions, err := roleRepository.LoadPermissions(context.Background())
	if err != nil {
		log.Fatalf("gagal memuat permission: %v", err)
	}

	permissions := helper.NewPermissionSet(rawPermissions)

	log.Printf("permission dimuat, roles=%v", permissions.KnownRoles())

	// ============================================================
	// 3. BUAT APLIKASI FIBER & GLOBAL MIDDLEWARE
	// ============================================================
	app := fiber.New(fiber.Config{
		AppName: "Praktikum Backend Lanjut - Modul 6",
	})

	app.Use(requestid.New())
	app.Use(logger.New())
	app.Use(cors.New())

	// ============================================================
	// 4. REGISTER ROUTES (LANGKAH 6)
	// ============================================================
	route.Register(app, route.Dependencies{
		Pool:        pool,
		Permissions: permissions,
	})

	// ============================================================
	// 5. JALANKAN SERVER
	// ============================================================
	port := config.GetEnv("APP_PORT", "3000")
	fmt.Println("Server berjalan di http://localhost:" + port)

	log.Fatal(app.Listen(":" + port))
}
