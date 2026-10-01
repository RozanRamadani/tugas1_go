package middleware

import (
	"siakad-mini/helper"

	"github.com/gofiber/fiber/v2"
)

// RequireRole adalah middleware otorisasi (Authorization) untuk membatasi akses endpoint.
// Middleware ini WAJIB dijalankan SETELAH middleware Protected() karena ia bergantung
// pada data 'role' yang disimpan di Fiber Locals oleh Protected().
//
// Penggunaan parameter variadic (...string) memungkinkan kita melempar lebih dari satu role,
// misal: RequireRole("admin", "mahasiswa")
func RequireRole(allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Mengambil role user yang sedang login dari Locals
		userRole, ok := c.Locals("role").(string)
		if !ok {
			// Jika tidak ada data role, berarti user belum melewati middleware Protected()
			return helper.ErrorResponse(c, fiber.StatusForbidden, "Akses ditolak: Role tidak ditemukan")
		}

		// Mengecek apakah role user ada di dalam daftar allowedRoles yang diizinkan
		isAllowed := false
		for _, allowedRole := range allowedRoles {
			if userRole == allowedRole {
				isAllowed = true
				break
			}
		}

		// Jika role user tidak termasuk yang diizinkan, tolak akses dengan status 403 Forbidden
		if !isAllowed {
			return helper.ErrorResponse(c, fiber.StatusForbidden, "Akses ditolak: Anda tidak memiliki izin (Hak Akses) untuk rute ini")
		}

		// Jika diizinkan, teruskan ke handler tujuan
		return c.Next()
	}
}
