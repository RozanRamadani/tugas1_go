package main

import (
	"log"
	"time"

	"siakad-mini/config"
	"siakad-mini/domain"
	"siakad-mini/handler"
	"siakad-mini/helper"
	"siakad-mini/middleware"
	"siakad-mini/repository"
	"siakad-mini/service"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

func main() {
	// 1. Inisialisasi Environment Variables & Koneksi Database
	config.LoadConfig()
	config.ConnectDB()

	// 2. Setup Dependency Injection (DI)
	// -- Student --
	studentRepo := repository.NewStudentRepository(config.DB)
	studentService := service.NewStudentService(studentRepo)
	studentHandler := handler.NewStudentHandler(studentService)

	// -- Auth --
	userRepo := repository.NewUserRepository(config.DB)
	authService := service.NewAuthService(userRepo, studentRepo)
	authHandler := handler.NewAuthHandler(authService)

	// -- Course --
	courseRepo := repository.NewCourseRepository(config.DB)
	courseService := service.NewCourseService(courseRepo)
	courseHandler := handler.NewCourseHandler(courseService)

	// -- Enrollment --
	enrollmentService := service.NewEnrollmentService(config.DB)
	enrollmentHandler := handler.NewEnrollmentHandler(enrollmentService)

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

	// Rute Publik (Auth)
	authGroup := api.Group("/auth")

	// Middleware Rate Limiter untuk Login (Maks 5 kegagalan / menit)
	loginLimiter := limiter.New(limiter.Config{
		Max:                    5,
		Expiration:             1 * time.Minute,
		SkipSuccessfulRequests: true, // Hanya menghitung status >= 400 sebagai kegagalan
		LimitReached: func(c *fiber.Ctx) error {
			return helper.ErrorResponse(c, fiber.StatusTooManyRequests, "Terlalu banyak percobaan login. Silakan coba lagi nanti.")
		},
	})

	authGroup.Post("/login", loginLimiter, authHandler.Login)
	authGroup.Get("/me", middleware.Protected(), authHandler.Me)

	// Rute Student
	studentGroup := api.Group("/students")
	studentGroup.Use(middleware.Protected())

	// 10A - GET /api/v1/students
	studentGroup.Get("/", middleware.RequireRole(string(domain.RoleAdmin)), studentHandler.GetAll)
	// 10B - POST /api/v1/students
	studentGroup.Post("/", middleware.RequireRole(string(domain.RoleAdmin)), studentHandler.Create)
	// 10C - GET /api/v1/students/:id (Admin & Mahasiswa)
	studentGroup.Get("/:id", middleware.RequireRole(string(domain.RoleAdmin), string(domain.RoleMahasiswa)), studentHandler.GetByID)
	// 10D - PUT /api/v1/students/:id
	studentGroup.Put("/:id", middleware.RequireRole(string(domain.RoleAdmin)), studentHandler.Update)
	// 10E - DELETE /api/v1/students/:id
	studentGroup.Delete("/:id", middleware.RequireRole(string(domain.RoleAdmin)), studentHandler.Delete)

	// Rute Course (Bisa diakses Admin maupun Mahasiswa)
	courseGroup := api.Group("/courses")
	courseGroup.Use(middleware.Protected())
	courseGroup.Use(middleware.RequireRole(string(domain.RoleAdmin), string(domain.RoleMahasiswa)))

	// 11 - GET /api/v1/courses
	courseGroup.Get("/", courseHandler.GetAll)

	// Rute Enrollment (Hanya Mahasiswa)
	enrollmentGroup := api.Group("/enrollments")
	enrollmentGroup.Use(middleware.Protected())
	enrollmentGroup.Use(middleware.RequireRole(string(domain.RoleMahasiswa)))

	// 12 - POST /api/v1/enrollments
	enrollmentGroup.Post("/", enrollmentHandler.Create)

	// 13 - DELETE /api/v1/enrollments/:id
	enrollmentGroup.Delete("/:id", enrollmentHandler.Delete)

	// 5. Jalankan server
	port := config.GetEnv("APP_PORT", "3000")
	log.Println("Server is running on port " + port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
