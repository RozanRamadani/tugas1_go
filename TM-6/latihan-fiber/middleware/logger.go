package middleware

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"

	"latihan-fiber/helper"
)

// RequestLogger mencatat setiap HTTP request beserta identitas user (jika terautentikasi).
func RequestLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()

		requestID, _ := c.Locals("requestid").(string)

		attrs := []any{
			slog.String("request_id", requestID),
			slog.String("method", c.Method()),
			slog.String("path", c.Path()),
			slog.Int("status", c.Response().StatusCode()),
			slog.Duration("duration", time.Since(start)),
			slog.String("ip", c.IP()),
		}

		// Identitas (user_id & role) ikut dicatat bila request sudah melewati RequireAuth.
		// Tanpa ini, log sebuah 403 tidak berguna: kita tahu ada yang ditolak, tetapi tidak tahu siapa dan mengapa.
		if user, ok := helper.CurrentUser(c); ok {
			attrs = append(attrs,
				slog.Int("user_id", user.UserID),
				slog.String("role", user.Role),
			)
		}

		slog.Info("http_request", attrs...)
		return err
	}
}
