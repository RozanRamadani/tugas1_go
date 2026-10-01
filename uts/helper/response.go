package helper

import (
	"github.com/gofiber/fiber/v2"
)

// SuccessResponse digunakan untuk mengembalikan respons sukses (2xx) yang terstruktur.
// Parameter 'data' bersifat opsional. Jika tidak ada data yang ingin dikembalikan, oper nil.
func SuccessResponse(c *fiber.Ctx, statusCode int, message string, data interface{}) error {
	response := fiber.Map{
		"status":  "success",
		"message": message,
	}
	
	if data != nil {
		response["data"] = data
	}

	return c.Status(statusCode).JSON(response)
}

// ErrorResponse digunakan untuk mengembalikan respons gagal/kesalahan umum (4xx, 5xx).
// Contoh: Data tidak ditemukan, Kredensial salah, Internal Server Error.
func ErrorResponse(c *fiber.Ctx, statusCode int, message string) error {
	return c.Status(statusCode).JSON(fiber.Map{
		"status":  "error",
		"message": message,
	})
}

// ValidationErrorResponse khusus digunakan ketika request payload (JSON input) gagal tervalidasi.
// Biasa mengembalikan status 400 Bad Request beserta daftar error detailnya.
func ValidationErrorResponse(c *fiber.Ctx, errors interface{}) error {
	return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
		"status":  "fail",
		"message": "Validation failed",
		"errors":  errors,
	})
}
