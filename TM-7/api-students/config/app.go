package config

import (
	"errors"
	"log"

	"api-students/app/handler"
	"api-students/app/service"
	"api-students/helper"
	"api-students/middleware"
	"api-students/route"

	"github.com/gofiber/fiber/v2"
)

func NewApp(
	studentService *service.StudentService,
	authService *service.AuthService,
	jwtManager *helper.JWTManager,
	perms *helper.PermissionSet,
) *fiber.App {

	app := fiber.New(fiber.Config{
		ErrorHandler: newErrorHandler(),
	})

	requestLogger := middleware.NewRequestLogger(
		func(data middleware.RequestLog) {
			LogRequest(data)
		},
	)

	app.Use(requestLogger.Handler)

	studentHandler := handler.NewStudentHandler(
		studentService,
	)

	authHandler := handler.NewAuthHandler(
		authService,
	)

	route.Register(
		app,
		studentHandler,
		authHandler,
		jwtManager,
		perms,
	)

	log.Println("route berhasil didaftarkan")

	return app
}

type ErrorResponse struct {
	Success   bool              `json:"success"`
	Code      string            `json:"code"`
	Message   string            `json:"message"`
	RequestID string            `json:"request_id,omitempty"`
	Fields    map[string]string `json:"fields,omitempty"`
}

func newErrorHandler() fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		status := fiber.StatusInternalServerError
		code := helper.CodeInternal
		message := "terjadi kesalahan pada server"
		var fields map[string]string

		var appErr *helper.AppError
		if errors.As(err, &appErr) {
			status = appErr.Status
			code = appErr.Code
			message = appErr.Message
			fields = appErr.Fields
		} else {
			var fiberErr *fiber.Error
			if errors.As(err, &fiberErr) {
				status = fiberErr.Code
				code = "ROUTER_ERROR"
				message = fiberErr.Message
			} else {
				internal := helper.Internal(err)
				status = internal.Status
				code = internal.Code
				message = internal.Message
			}
		}

		if status >= 500 {
			log.Printf("ERROR: %s: %s", code, message)
		} else {
			log.Printf("WARN: %s: %s", code, message)
		}

		var reqID string
		if v, ok := c.Locals("requestid").(string); ok {
			reqID = v
		}

		return c.Status(status).JSON(ErrorResponse{
			Success:   false,
			Code:      code,
			Message:   message,
			RequestID: reqID,
			Fields:    fields,
		})
	}
}
