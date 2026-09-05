package middleware

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
)

func Register(app *fiber.App, logger *slog.Logger) {

	// Logger middleware
	app.Use(func(c *fiber.Ctx) error {

		start := time.Now()

		err := c.Next()

		logger.Info(
			"request",
			"method", c.Method(),
			"path", c.Path(),
			"status", c.Response().StatusCode(),
			"duration", time.Since(start),
		)

		return err
	})
}
