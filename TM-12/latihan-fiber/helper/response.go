package helper

import (
	"latihan-fiber/app/model"

	"github.com/gofiber/fiber/v2"
)

// CurrentUser mengambil AuthUser yang tersimpan dalam Ctx.
func CurrentUser(c *fiber.Ctx) (model.AuthUser, bool) {
	user, ok := c.Locals("user").(model.AuthUser)
	return user, ok
}

// Fail mengembalikan respon JSON error standar.
func Fail(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(model.WebResponse{
		Success: false,
		Message: message,
	})
}

// FailValidation mengembalikan respon error validasi 422.
func FailValidation(c *fiber.Ctx, errs map[string]string) error {
	return c.Status(fiber.StatusUnprocessableEntity).JSON(model.WebResponse{
		Success: false,
		Message: "validasi gagal",
		Errors:  errs,
	})
}

// Success mengembalikan respon JSON sukses standar.
func Success(c *fiber.Ctx, status int, message string, data any) error {
	return c.Status(status).JSON(model.WebResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Created mengembalikan respon 201 Created dengan header Location.
func Created(c *fiber.Ctx, message string, data any, location string) error {
	c.Set("Location", location)
	return c.Status(fiber.StatusCreated).JSON(model.WebResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// NoContent mengembalikan respon 204 No Content.
func NoContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}
