package route

import (
	"api-students/app/handler"
	"api-students/middleware"

	"github.com/gofiber/fiber/v2"
)

func Register(
	app *fiber.App,
	studentHandler *handler.StudentHandler,
) {
	api := app.Group("/api/v1")

	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"success": true,
			"message": "server berjalan",
		})
	})

	students := api.Group(
		"/students",
		middleware.RequireJSON,
	)

	students.Get("/", studentHandler.List)
	students.Get("/:id", studentHandler.Get)
	students.Post("/", studentHandler.Create)
	students.Put("/:id", studentHandler.Replace)
	students.Patch("/:id", studentHandler.Patch)
	students.Delete("/:id", studentHandler.Delete)
}
