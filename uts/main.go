package main

import (
	"log"
	"siakad-mini/config"
	"siakad-mini/domain"
	"siakad-mini/handler"
	"siakad-mini/middleware"
	"siakad-mini/repository"
	"siakad-mini/service"

	"github.com/gofiber/fiber/v2"
)

func main() {
	// 1. Inisialisasi Environment Variables & Koneksi Database
	config.LoadConfig()
	config.ConnectDB()

	// 2. Setup Dependency Injection (DI)
	// -- Auth --
	userRepo := repository.NewUserRepository(config.DB)
	authService := service.NewAuthService(userRepo)
	authHandler := handler.NewAuthHandler(authService)

	// -- Student --
	studentRepo := repository.NewStudentRepository(config.DB)
	studentService := service.NewStudentService(studentRepo)
	studentHandler := handler.NewStudentHandler(studentService)

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
	authGroup.Post("/login", authHandler.Login)

	// Rute Student (Hanya Admin)
	studentGroup := api.Group("/students")
	studentGroup.Use(middleware.Protected())
	studentGroup.Use(middleware.RequireRole(string(domain.RoleAdmin)))
	
	// 10A - GET /api/v1/students
	studentGroup.Get("/", studentHandler.GetAll)
	// 10B - POST /api/v1/students
	studentGroup.Post("/", studentHandler.Create)
	// 10C - GET /api/v1/students/:id
	studentGroup.Get("/:id", studentHandler.GetByID)
	// 10D - PUT /api/v1/students/:id
	studentGroup.Put("/:id", studentHandler.Update)
	// 10E - DELETE /api/v1/students/:id
	studentGroup.Delete("/:id", studentHandler.Delete)

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

	// 5. Jalankan server
	port := config.GetEnv("APP_PORT", "3000")
	log.Println("Server is running on port " + port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
