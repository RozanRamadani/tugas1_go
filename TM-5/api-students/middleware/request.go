package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

var metodeBerbody = map[string]bool{
	fiber.MethodPost:  true,
	fiber.MethodPut:   true,
	fiber.MethodPatch: true,
}

// RequireJSON memastikan request yang memiliki body
// menggunakan Content-Type application/json.
func RequireJSON(c *fiber.Ctx) error {

	if metodeBerbody[c.Method()] {

		ct := c.Get("Content-Type")

		if !strings.HasPrefix(ct, fiber.MIMEApplicationJSON) {
			return c.Status(fiber.StatusUnsupportedMediaType).JSON(
				fiber.Map{
					"success": false,
					"message": "Content-Type harus application/json",
				},
			)
		}
	}

	return c.Next()
}
