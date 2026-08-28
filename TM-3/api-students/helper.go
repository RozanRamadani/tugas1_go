package main

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// 200 OK
func ok(c *fiber.Ctx, message string, data any) error {
	return c.Status(fiber.StatusOK).JSON(WebResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// 201 Created + Location
func created(c *fiber.Ctx, message string, data any, location string) error {
	c.Set("Location", location)

	return c.Status(fiber.StatusCreated).JSON(WebResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// 200 OK + pagination
func okList(c *fiber.Ctx, message string, data any, meta *Meta) error {
	return c.Status(fiber.StatusOK).JSON(WebResponse{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

// Response error umum.
func fail(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(WebResponse{
		Success: false,
		Message: message,
	})
}

// 422 = data valid secara JSON, tetapi gagal validasi.
func failValidation(c *fiber.Ctx, errors map[string]string) error {
	return c.Status(fiber.StatusUnprocessableEntity).JSON(WebResponse{
		Success: false,
		Message: "validasi gagal",
		Errors:  errors,
	})
}

// 204 = berhasil tanpa response body.
func noContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

// Membaca query string dengan nilai default yang aman.
func parseListQuery(c *fiber.Ctx) ListQuery {
	q := ListQuery{
		Page:   1,
		Limit:  10,
		Search: strings.TrimSpace(c.Query("search")),
		Sort:   c.Query("sort", "id"),
		Order:  strings.ToLower(c.Query("order", "asc")),
	}

	// page minimal 1.
	if page, err := strconv.Atoi(c.Query("page", "1")); err == nil && page >= 1 {
		q.Page = page
	}

	// limit maksimal 100 agar request tidak terlalu besar.
	if limit, err := strconv.Atoi(c.Query("limit", "10")); err == nil && limit >= 1 {
		if limit > 100 {
			limit = 100
		}
		q.Limit = limit
	}

	// Filter is_active=true/false.
	if value := c.Query("is_active"); value != "" {
		if strings.EqualFold(value, "true") ||
			strings.EqualFold(value, "false") {

			active := strings.EqualFold(value, "true")
			q.IsActive = &active
		}
	}

	return q
}