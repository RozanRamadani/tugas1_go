package config

import (
	"log"

	"api-students/app/handler"
	"api-students/app/service"
	"api-students/middleware"
	"api-students/route"

	"github.com/gofiber/fiber/v2"
)

func NewApp(studentService *service.StudentService) *fiber.App {

	app := fiber.New()

	requestLogger := middleware.NewRequestLogger(
		func(data middleware.RequestLog) {
			LogRequest(data)
		},
	)

	app.Use(requestLogger.Handler)

	studentHandler := handler.NewStudentHandler(studentService)

	route.Register(
		app,
		studentHandler,
	)

	log.Println("route berhasil didaftarkan")

	return app
}
