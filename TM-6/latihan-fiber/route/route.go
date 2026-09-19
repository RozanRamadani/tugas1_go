package route

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"latihan-fiber/helper"
	"latihan-fiber/middleware"
)

// Dependencies menampung seluruh dependency yang dibutuhkan oleh route handler.
// Dengan struct ini, menambahkan service/helper baru cukup menambah field di sini.
type Dependencies struct {
	Pool        *pgxpool.Pool
	Permissions *helper.PermissionSet
	// UserService & AuthService akan dimasukkan saat Langkah 7
}

// Register mendaftarkan seluruh endpoint aplikasi beserta penjaganya (middleware).
func Register(app *fiber.App, deps Dependencies) {
	api := app.Group("/api/v1")

	// ------------------------------------------------------------
	// 1. ENDPOINT PUBLIK
	// ------------------------------------------------------------
	api.Get("/health", healthCheck(deps.Pool))

	// ------------------------------------------------------------
	// 2. ENDPOINT USERS (Wajib Login & Diperiksa Hak Aksesnya)
	// ------------------------------------------------------------
	// Group /users ini langsung dipasangi RequireJSON & RequireAuth.
	// Tanpa token sah, request akan ditolak dengan HTTP 401 Unauthorized.
	users := api.Group("/users")

	perms := deps.Permissions

	// ------------------------------------------------------------
	// KATEGORI A: Hak Akses Tanpa Melihat Isi Data -> Diperiksa di Middleware
	// ------------------------------------------------------------

	// GET /api/v1/users -> Melihat daftar seluruh user
	users.Get("/",
		middleware.RequirePermission(perms, "user:list"),
		func(c *fiber.Ctx) error {
			return c.SendString("List user")
		},
	)

	// POST /api/v1/users -> Menambah user baru
	users.Post("/",
		middleware.RequirePermission(perms, "user:update:any"),
		func(c *fiber.Ctx) error {
			return c.SendString("Create user")
		},
	)

	// DELETE /api/v1/users/:id -> Menghapus user
	users.Delete("/:id",
		middleware.RequirePermission(perms, "user:delete"),
		func(c *fiber.Ctx) error {
			return c.SendString("Delete user")
		},
	)

	// PATCH /api/v1/users/:id/role -> Mengubah role milik user lain
	users.Patch("/:id/role",
		middleware.RequirePermission(perms, "role:assign"),
		func(c *fiber.Ctx) error {
			return c.SendString("Assign role user")
		},
	)

	// ------------------------------------------------------------
	// KATEGORI B: Hak Akses Bergantung Kepemilikan Data -> Diperiksa di Service
	// ------------------------------------------------------------
	// Ketiga endpoint ini terlihat tidak dipasangi RequirePermission di sini.
	// Mengapa? Karena middleware belum tahu apakah :id yang diminta adalah
	// milik pemanggil sendiri atau milik orang lain.
	// Pengecekannya diserahkan ke layer Service (Langkah 7).

	users.Get("/:id", func(c *fiber.Ctx) error {
		return c.SendString("Get user by ID (diperiksa di service)")
	})

	users.Put("/:id", func(c *fiber.Ctx) error {
		return c.SendString("Replace user (diperiksa di service)")
	})

	users.Patch("/:id", func(c *fiber.Ctx) error {
		return c.SendString("Patch user (diperiksa di service)")
	})
}

func healthCheck(pool *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"message": "server berjalan",
		})
	}
}
