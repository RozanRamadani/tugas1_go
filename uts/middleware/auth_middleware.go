package middleware

import (
	"strings"

	"siakad-mini/config"
	"siakad-mini/helper"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// Protected adalah middleware untuk memvalidasi JWT token
func Protected() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 1. Membaca Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			// 7. Mengembalikan 401 jika token tidak ada
			return helper.ErrorResponse(c, fiber.StatusUnauthorized, "Akses ditolak: Token tidak ditemukan")
		}

		// 2. Mengambil Bearer token
		// Header biasanya berformat: "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return helper.ErrorResponse(c, fiber.StatusUnauthorized, "Akses ditolak: Format token tidak valid")
		}
		tokenString := parts[1]

		// 3. Memvalidasi signature & 4. Memvalidasi expiration
		secret := config.GetEnv("JWT_SECRET", "supersecretkey")
		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			// Memastikan metode enkripsi sesuai yang kita buat di service (HS256)
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.ErrUnauthorized
			}
			return []byte(secret), nil
		})

		// Jika error (misal expired) atau token dimanipulasi
		if err != nil || !token.Valid {
			return helper.ErrorResponse(c, fiber.StatusUnauthorized, "Akses ditolak: Token tidak valid atau kedaluwarsa")
		}

		// 5. Mengambil identity user
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return helper.ErrorResponse(c, fiber.StatusUnauthorized, "Akses ditolak: Gagal mengekstrak data dari token")
		}

		// 6. Menyimpan identity ke Fiber Locals
		c.Locals("user_id", claims["user_id"])
		c.Locals("role", claims["role"])

		// Meneruskan request ke Handler utama (endpoint yang dituju)
		return c.Next()
	}
}
