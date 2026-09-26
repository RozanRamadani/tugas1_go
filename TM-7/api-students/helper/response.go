package helper

import "github.com/gofiber/fiber/v2"

type WebResponse struct {
	Success bool              `json:"success"`
	Message string            `json:"message"`
	Data    any               `json:"data,omitempty"`
	Meta    any               `json:"meta,omitempty"`
	Errors  map[string]string `json:"errors,omitempty"`
}

func Success(c *fiber.Ctx, status int, message string, data any) error {
	return c.Status(status).JSON(WebResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func OK(c *fiber.Ctx, message string, data any) error {
	return Success(c, fiber.StatusOK, message, data)
}

func OKList(c *fiber.Ctx, message string, data any, meta any) error {
	return c.Status(fiber.StatusOK).JSON(WebResponse{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

func Created(c *fiber.Ctx, message string, data any, location string) error {
	c.Set("Location", location)

	return c.Status(fiber.StatusCreated).JSON(WebResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func NoContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

func Fail(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(WebResponse{
		Success: false,
		Message: message,
	})
}

func FailValidation(c *fiber.Ctx, errs map[string]string) error {
	return c.Status(fiber.StatusUnprocessableEntity).JSON(WebResponse{
		Success: false,
		Message: "validasi gagal",
		Errors:  errs,
	})
}
