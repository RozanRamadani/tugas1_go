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

	app := config.NewApp(studentService)

	log.Println("Server berjalan di http://localhost:3000")

	if err := app.Listen(":3000"); err != nil {
		log.Fatal(err)
	}
}
