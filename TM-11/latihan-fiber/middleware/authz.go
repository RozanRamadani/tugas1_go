package middleware

import (
	"latihan-fiber/helper"

	"github.com/gofiber/fiber/v2"
)

// RequirePermission menolak request yang role-nya tidak memiliki
// permission tertentu.
//
// Middleware ini digunakan untuk keputusan akses yang dapat ditentukan
// tanpa melihat isi data.
func RequirePermission(
	perms *helper.PermissionSet,
	permission string,
) fiber.Handler {

	return func(c *fiber.Ctx) error {

		user, ok := helper.CurrentUser(c)

		if !ok {
			return helper.Fail(
				c,
				fiber.StatusUnauthorized,
				"belum terautentikasi",
			)
		}

		if !perms.Can(user.Role, permission) {
			return helper.Fail(
				c,
				fiber.StatusForbidden,
				"role "+user.Role+" tidak memiliki hak "+permission,
			)
		}

		return c.Next()
	}
}

// RequireRole memeriksa nama role secara langsung.
//
// Fungsi ini disediakan sebagai pembanding dengan RequirePermission.
func RequireRole(roles ...string) fiber.Handler {

	allowed := make(map[string]struct{}, len(roles))

	for _, role := range roles {
		allowed[role] = struct{}{}
	}

	return func(c *fiber.Ctx) error {

		user, ok := helper.CurrentUser(c)

		if !ok {
			return helper.Fail(
				c,
				fiber.StatusUnauthorized,
				"belum terautentikasi",
			)
		}

		if _, granted := allowed[user.Role]; !granted {
			return helper.Fail(
				c,
				fiber.StatusForbidden,
				"role Anda tidak berhak mengakses endpoint ini",
			)
		}

		return c.Next()
	}
}