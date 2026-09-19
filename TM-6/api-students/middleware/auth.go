package middleware

import (
	"errors"
	"strings"
	"time"

	"api-students/app/model"
	"api-students/helper"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

const (
	UserIDKey = "user_id"
	RoleKey   = "role"
	UserKey   = "user"
)

func RequireAuth(jwtManager *helper.JWTManager) fiber.Handler {

	return func(c *fiber.Ctx) error {

		authHeader := strings.TrimSpace(
			c.Get("Authorization"),
		)

		if authHeader == "" {
			c.Set(
				"WWW-Authenticate",
				`Bearer realm="api"`,
			)

			return c.Status(
				fiber.StatusUnauthorized,
			).JSON(fiber.Map{
				"success": false,
				"message": "header Authorization tidak ada atau salah bentuk",
			})
		}

		parts := strings.Fields(authHeader)

		if len(parts) != 2 ||
			!strings.EqualFold(parts[0], "Bearer") {

			c.Set(
				"WWW-Authenticate",
				`Bearer realm="api"`,
			)

			return c.Status(
				fiber.StatusUnauthorized,
			).JSON(fiber.Map{
				"success": false,
				"message": "header Authorization tidak ada atau salah bentuk",
			})
		}

		token := parts[1]

		claims, err := jwtManager.ParseAccessToken(token)

		if err != nil {

			c.Set(
				"WWW-Authenticate",
				`Bearer realm="api"`,
			)

			if errors.Is(
				err,
				helper.ErrExpiredToken,
			) {
				return c.Status(
					fiber.StatusUnauthorized,
				).JSON(fiber.Map{
					"success": false,
					"message": "access token kedaluwarsa",
				})
			}

			return c.Status(
				fiber.StatusUnauthorized,
			).JSON(fiber.Map{
				"success": false,
				"message": "access token tidak valid",
			})
		}

		// Simpan AuthUser lengkap.
		user := model.AuthUser{
			UserID: claims.UserID,
			Role:   claims.Role,
		}

		c.Locals(UserKey, user)

		// Tetap simpan nilai individual agar kompatibel
		// dengan kode lama.
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

func LoginRateLimiter() fiber.Handler {

	return limiter.New(limiter.Config{
		Max:        5,
		Expiration: 1 * time.Minute,

		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},

		LimitReached: func(c *fiber.Ctx) error {

			c.Set(
				"Retry-After",
				"60",
			)

			return c.Status(
				fiber.StatusTooManyRequests,
			).JSON(fiber.Map{
				"success": false,
				"message": "terlalu banyak percobaan login, coba lagi dalam satu menit",
			})
		},
	})
}
