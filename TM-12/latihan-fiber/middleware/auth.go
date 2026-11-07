package middleware

import (
	"errors"
	"fmt"
	"strings"

	"latihan-fiber/app/model"
	"latihan-fiber/helper"

	"github.com/gofiber/fiber/v2"
)

const (
	UserIDKey = "user_id"
	RoleKey   = "role"
	UserKey   = "user"
)

func RequireAuth(jwtManager *helper.JWTManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := strings.TrimSpace(c.Get("Authorization"))

		if authHeader == "" {
			c.Set("WWW-Authenticate", `Bearer realm="api"`)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "header Authorization tidak ada atau salah bentuk",
			})
		}

		parts := strings.Fields(authHeader)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.Set("WWW-Authenticate", `Bearer realm="api"`)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "header Authorization tidak ada atau salah bentuk",
			})
		}

		token := parts[1]
		claims, err := jwtManager.ParseAccessToken(token)
		if err != nil {
			c.Set("WWW-Authenticate", `Bearer realm="api"`)
			if errors.Is(err, helper.ErrExpiredToken) {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"success": false,
					"message": "access token kedaluwarsa",
				})
			}
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "access token tidak valid",
			})
		}

		username := claims.Username
		if username == "" {
			username = fmt.Sprintf("mhs%03d", claims.UserID)
		}

		user := model.AuthUser{
			UserID:   claims.UserID,
			Username: username,
			Role:     claims.Role,
		}

		c.Locals(UserKey, user)
		c.Locals(UserIDKey, claims.UserID)
		c.Locals(RoleKey, claims.Role)

		return c.Next()
	}
}
