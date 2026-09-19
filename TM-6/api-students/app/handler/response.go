package handler

import "github.com/gofiber/fiber/v2"

type WebResponse struct {
	Success bool              `json:"success"`
	Message string            `json:"message"`
	Data    interface{}       `json:"data,omitempty"`
	Meta    *Meta             `json:"meta,omitempty"`
	Errors  map[string]string `json:"errors,omitempty"`
}

type Meta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

func ok(
	c *fiber.Ctx,
	message string,
	data interface{},
) error {
	return c.Status(fiber.StatusOK).JSON(
		WebResponse{
			Success: true,
			Message: message,
			Data:    data,
		},
	)
}

func okList(
	c *fiber.Ctx,
	message string,
	data interface{},
	meta *Meta,
) error {
	return c.Status(fiber.StatusOK).JSON(
		WebResponse{
			Success: true,
			Message: message,
			Data:    data,
			Meta:    meta,
		},
	)
}

func created(
	c *fiber.Ctx,
	message string,
	data interface{},
	location string,
) error {

	c.Set("Location", location)

	return c.Status(fiber.StatusCreated).JSON(
		WebResponse{
			Success: true,
			Message: message,
			Data:    data,
		},
	)
}

func noContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

func fail(
	c *fiber.Ctx,
	status int,
	message string,
) error {
	return c.Status(status).JSON(
		WebResponse{
			Success: false,
			Message: message,
		},
	)
}

func failValidation(
	c *fiber.Ctx,
	errs map[string]string,
) error {
	return c.Status(fiber.StatusUnprocessableEntity).JSON(
		WebResponse{
			Success: false,
			Message: "validasi gagal",
			Errors:  errs,
		},
	)
}
