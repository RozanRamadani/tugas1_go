package handler

import "github.com/gofiber/fiber/v2"

// ============================================================
// RESPONSE
// ============================================================

type WebResponse struct {
	Success bool              `json:"success"`
	Message string            `json:"message"`
	Data    any               `json:"data,omitempty"`
	Meta    *Meta             `json:"meta,omitempty"`
	Errors  map[string]string `json:"errors,omitempty"`
}

// Meta digunakan untuk informasi pagination.
type Meta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// ============================================================
// RESPONSE BERHASIL
// ============================================================

// ok mengembalikan response HTTP 200.
func ok(c *fiber.Ctx, message string, data any) error {
	return c.Status(fiber.StatusOK).JSON(
		WebResponse{
			Success: true,
			Message: message,
			Data:    data,
		},
	)
}

// okList mengembalikan response HTTP 200
// khusus untuk data yang menggunakan pagination.
func okList(
	c *fiber.Ctx,
	message string,
	data any,
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

// created mengembalikan response HTTP 201
// ketika data berhasil dibuat.
func created(
	c *fiber.Ctx,
	message string,
	data any,
	location string,
) error {

	// Memberitahu client lokasi resource yang baru dibuat.
	c.Set("Location", location)

	return c.Status(fiber.StatusCreated).JSON(
		WebResponse{
			Success: true,
			Message: message,
			Data:    data,
		},
	)
}

// noContent mengembalikan HTTP 204.
// Digunakan ketika DELETE berhasil dan tidak ada
// data yang perlu dikirim kembali.
func noContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

// ============================================================
// RESPONSE ERROR
// ============================================================

// fail digunakan untuk error biasa.
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

// failValidation digunakan ketika validasi input gagal.
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

// ============================================================
// WRAPPER UNTUK PACKAGE MAIN
// ============================================================

// main.go berada di package main.
// Karena fungsi ok() dan fail() menggunakan huruf kecil,
// fungsi tersebut tidak bisa dipanggil langsung dari package main.
//
// Oleh karena itu kita menyediakan wrapper dengan huruf besar.

func Ok(
	c *fiber.Ctx,
	message string,
	data any,
) error {
	return ok(c, message, data)
}

func Fail(
	c *fiber.Ctx,
	status int,
	message string,
) error {
	return fail(c, status, message)
}
