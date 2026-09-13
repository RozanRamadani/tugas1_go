package route

import (
	"api-students/app/handler"
	"api-students/helper"
	"api-students/middleware"

	"github.com/gofiber/fiber/v2"
)

func Register(
	app *fiber.App,
	studentHandler *handler.StudentHandler,
	authHandler *handler.AuthHandler,
	jwtManager *helper.JWTManager,
) {
	api := app.Group("/api/v1")

	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"success": true,
			"message": "server berjalan",
		})
	})

	auth := api.Group("/auth", middleware.RequireJSON)
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", middleware.LoginRateLimiter(), authHandler.Login)
	auth.Post("/refresh", authHandler.Refresh)
	auth.Post("/logout", authHandler.Logout)
	auth.Get("/me", middleware.RequireAuth(jwtManager), authHandler.Me)

	students := api.Group(
		"/students",
		middleware.RequireJSON,
		middleware.RequireAuth(jwtManager),
	)

	students.Get("/", studentHandler.List)
	students.Get("/:id", studentHandler.Get)
	students.Post("/", studentHandler.Create)
	students.Put("/:id", studentHandler.Replace)
	students.Patch("/:id", studentHandler.Patch)
	students.Delete("/:id", studentHandler.Delete)
}
