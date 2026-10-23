package config

import (
	"errors"
	"log"

	"latihan-fiber/helper"
	"latihan-fiber/middleware"
	"latihan-fiber/route"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

type ErrorResponse struct {
	Success   bool              `json:"success"`
	Code      string            `json:"code"`
	Message   string            `json:"message"`
	RequestID string            `json:"request_id,omitempty"`
	Fields    map[string]string `json:"fields,omitempty"`
}

func NewApp(deps route.Dependencies) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      "Praktikum Backend Lanjut - Modul 9",
		ErrorHandler: newErrorHandler(),
	})

	app.Use(requestid.New())
	app.Use(middleware.RequestLogger())
	app.Use(cors.New())

	route.Register(app, deps)

	return app
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
			log.Printf("WARN: request_rejected %s: %s", code, message)
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
