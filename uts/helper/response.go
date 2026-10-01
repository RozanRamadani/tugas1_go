package helper

import (
	"github.com/gofiber/fiber/v2"
)

// SuccessResponse digunakan untuk mengembalikan respons sukses (2xx) yang terstruktur.
// Parameter 'data' bersifat opsional. Jika tidak ada data yang ingin dikembalikan, oper nil.
func SuccessResponse(c *fiber.Ctx, statusCode int, message string, data interface{}) error {
	response := fiber.Map{
		"success": true,
		"message": message,
	}

	if data != nil {
		response["data"] = data
	}

	return c.Status(statusCode).JSON(response)
}

// SuccessResponseWithMeta digunakan untuk mengembalikan respons sukses dengan metadata.
func SuccessResponseWithMeta(c *fiber.Ctx, statusCode int, message string, data interface{}, meta interface{}) error {
	response := fiber.Map{
		"success": true,
		"message": message,
	}

	if data != nil {
		response["data"] = data
	}

	if meta != nil {
		response["meta"] = meta
	}

	return c.Status(statusCode).JSON(response)
}

// ErrorResponse digunakan untuk mengembalikan respons gagal/kesalahan umum (4xx, 5xx).
// Contoh: Data tidak ditemukan, Kredensial salah, Internal Server Error.
func ErrorResponse(c *fiber.Ctx, statusCode int, message string) error {
	return c.Status(statusCode).JSON(fiber.Map{
		"success": false,
		"message": message,
	})
}

// ValidationErrorResponse khusus digunakan ketika request payload (JSON input) gagal tervalidasi.
// Biasa mengembalikan status 400 Bad Request atau 422 Unprocessable Entity beserta daftar error detailnya.
func ValidationErrorResponse(c *fiber.Ctx, errors interface{}) error {
	return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
		"success": false,
		"message": "Validasi gagal",
		"errors":  errors,
	})
}
