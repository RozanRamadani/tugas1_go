package main

import (
	"context"
	"log"

	"api-students/app/repository"
	"api-students/app/service"
	"api-students/config"
	"api-students/database"
)

func main() {

	// Membaca file .env
	config.LoadEnv()

	// Membuat connection pool PostgreSQL
	pool, err := database.NewPool(context.Background())
	if err != nil {
		log.Fatal("gagal terhubung ke database:", err)
	}

	defer pool.Close()

	// Repository
	studentRepo := repository.NewStudentRepository(pool)

	// Service
	studentService := service.NewStudentService(studentRepo)

	// Membuat aplikasi Fiber dan mendaftarkan route
	app := config.NewApp(studentService)

	log.Println("Server berjalan di http://localhost:3000")

	log.Fatal(app.Listen(":3000"))
}
