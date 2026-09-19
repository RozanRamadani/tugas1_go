package route

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"latihan-fiber/helper"
	"latihan-fiber/middleware"
)

type Dependencies struct {
	Pool        *pgxpool.Pool
	Permissions *helper.PermissionSet
	// UserService & AuthService akan ditambahkan di Langkah 7
}

func Register(app *fiber.App, deps Dependencies) {
	api := app.Group("/api/v1")

	// --- publik ---
	api.Get("/health", healthCheck(deps.Pool))

	// --- users ---
	// Endpoint sementara /users untuk Langkah 6
	users := api.Group("/users")

	perms := deps.Permissions

	// Hak dapat diputuskan tanpa melihat data -> middleware
	users.Get("/",
		middleware.RequirePermission(perms, "user:list"),
		func(c *fiber.Ctx) error {
			return c.SendString("List user (placeholder)")
		},
	)

	users.Delete("/:id",
		middleware.RequirePermission(perms, "user:delete"),
		func(c *fiber.Ctx) error {
			return c.SendString("Delete user (placeholder)")
		},
	)

	users.Patch("/:id/role",
		middleware.RequirePermission(perms, "role:assign"),
		func(c *fiber.Ctx) error {
			return c.SendString("Assign role user (placeholder)")
		},
	)
}

func healthCheck(pool *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
		})
	}
}
