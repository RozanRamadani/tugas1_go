package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"api-students/helper"
)

const (
	UserIDKey = "user_id"
	RoleKey   = "role"
)

func RequireAuth(jwtManager *helper.JWTManager) fiber.Handler {

	return func(c *fiber.Ctx) error {

		authHeader := strings.TrimSpace(
			c.Get("Authorization"),
		)

		if authHeader == "" {
			return c.Status(
				fiber.StatusUnauthorized,
			).JSON(fiber.Map{
				"success": false,
				"message": "authorization header wajib diisi",
			})
		}

		parts := strings.Fields(authHeader)

		if len(parts) != 2 ||
			!strings.EqualFold(parts[0], "Bearer") {

			return c.Status(
				fiber.StatusUnauthorized,
			).JSON(fiber.Map{
				"success": false,
				"message": "format authorization harus Bearer <token>",
			})
		}

		token := parts[1]

		claims, err := jwtManager.ParseAccessToken(token)

		if err != nil {
			return c.Status(
				fiber.StatusUnauthorized,
			).JSON(fiber.Map{
				"success": false,
				"message": "access token tidak valid",
			})
		}

		c.Locals(UserIDKey, claims.UserID)
		c.Locals(RoleKey, claims.Role)

		return c.Next()
	}
}

func RequireRole(roles ...string) fiber.Handler {

	return func(c *fiber.Ctx) error {

		currentRole, ok := c.Locals(RoleKey).(string)

		if !ok || currentRole == "" {
			return c.Status(
				fiber.StatusForbidden,
			).JSON(fiber.Map{
				"success": false,
				"message": "role tidak ditemukan",
			})
		}

		for _, role := range roles {
			if currentRole == role {
				return c.Next()
			}
		}

		return c.Status(
			fiber.StatusForbidden,
		).JSON(fiber.Map{
			"success": false,
			"message": "akses ditolak",
		})
	}
}
