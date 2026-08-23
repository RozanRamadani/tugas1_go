package main

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func ok(c *fiber.Ctx, message string, data any) error {
	return c.Status(fiber.StatusOK).JSON(WebResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func created(c *fiber.Ctx, message string, data any, location string) error {
	c.Set("Location", location)

	return c.Status(fiber.StatusCreated).JSON(WebResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func okList(
	c *fiber.Ctx,
	message string,
	data any,
	meta *Meta,
) error {
	return c.Status(fiber.StatusOK).JSON(WebResponse{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

func fail(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(WebResponse{
		Success: false,
		Message: message,
	})
}

func failValidation(c *fiber.Ctx, errors map[string]string) error {
	return c.Status(fiber.StatusUnprocessableEntity).JSON(WebResponse{
		Success: false,
		Message: "validasi gagal",
		Errors:  errors,
	})
}

func noContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

func parseListQuery(c *fiber.Ctx) ListQuery {
	q := ListQuery{
		Page:   1,
		Limit:  10,
		Search: strings.TrimSpace(c.Query("search")),
		Sort:   c.Query("sort", "id"),
		Order:  strings.ToLower(c.Query("order", "asc")),
	}

	if page, err := strconv.Atoi(c.Query("page", "1")); err == nil && page >= 1 {
		q.Page = page
	}

	if limit, err := strconv.Atoi(c.Query("limit", "10")); err == nil && limit >= 1 {
		if limit > 100 {
			limit = 100
		}
		q.Limit = limit
	}

	if value := c.Query("is_active"); value != "" {
		active := strings.EqualFold(value, "true")

		if strings.EqualFold(value, "true") ||
			strings.EqualFold(value, "false") {
			q.IsActive = &active
		}
	}

	return q
}