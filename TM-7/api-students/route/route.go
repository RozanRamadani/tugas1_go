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
	perms *helper.PermissionSet,
) {

	api := app.Group("/api/v1")

	// ============================================================
	// HEALTH
	// ============================================================

	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"success": true,
			"message": "server berjalan",
		})
	})

	// ============================================================
	// AUTH
	// ============================================================

	auth := api.Group(
		"/auth",
		middleware.RequireJSON,
	)

	auth.Post(
		"/register",
		authHandler.Register,
	)

	auth.Post(
		"/login",
		middleware.LoginRateLimiter(),
		authHandler.Login,
	)

	auth.Post(
		"/refresh",
		authHandler.Refresh,
	)

	auth.Post(
		"/logout",
		authHandler.Logout,
	)

	auth.Get(
		"/me",
		middleware.RequireAuth(jwtManager),
		authHandler.Me,
	)

	// ============================================================
	// STUDENTS
	// ============================================================

	students := api.Group(
		"/students",
		middleware.RequireJSON,
		middleware.RequireAuth(jwtManager),
	)

	// LIST
	students.Get(
		"/",
		middleware.RequirePermission(
			perms,
			"student:list",
		),
		studentHandler.List,
	)

	// GET BY ID
	//
	// Tidak menggunakan RequirePermission :any.
	// Ownership diperiksa di service.
	students.Get(
		"/:id",
		studentHandler.Get,
	)

	// CREATE
	students.Post(
		"/",
		middleware.RequirePermission(
			perms,
			"student:create",
		),
		studentHandler.Create,
	)

	// PUT
	//
	// Ownership / student:update:any diperiksa di service.
	students.Put(
		"/:id",
		studentHandler.Replace,
	)

	// PATCH
	//
	// Ownership / student:update:any diperiksa di service.
	students.Patch(
		"/:id",
		studentHandler.Patch,
	)

	// DELETE
	students.Delete(
		"/:id",
		middleware.RequirePermission(
			perms,
			"student:delete",
		),
		studentHandler.Delete,
	)
}
