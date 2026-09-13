package main

import (
	"context"
	"log"
	"time"

	"api-students/app/repository"
	"api-students/app/service"
	"api-students/config"
	"api-students/database"
	"api-students/helper"
)

func main() {

	config.LoadEnv()

	if err := config.InitLogger(); err != nil {
		log.Fatal("gagal menginisialisasi logger:", err)
	}

	defer config.CloseLogger()

	pool, err := database.NewPool(context.Background())
	if err != nil {
		log.Fatal("gagal terhubung ke database:", err)
	}

	defer pool.Close()

	studentRepo := repository.NewStudentRepository(pool)

	studentService := service.NewStudentService(studentRepo)

	jwtSecret := config.GetEnv("JWT_SECRET", "super-secret-key-praktikum-backend-32byte")
	jwtIssuer := config.GetEnv("JWT_ISSUER", "praktikum-backend")
	jwtManager := helper.NewJWTManager(jwtSecret, jwtIssuer, 15*time.Minute, 7*24*time.Hour)

	app := config.NewApp(studentService, jwtManager)

	log.Println("Server berjalan di http://localhost:3000")

	if err := app.Listen(":3000"); err != nil {
		log.Fatal(err)
	}
}
