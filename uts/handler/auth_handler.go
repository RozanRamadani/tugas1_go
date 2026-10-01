package handler

import (
	"fmt"

	"siakad-mini/helper"
	"siakad-mini/service"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// AuthHandler menangani request HTTP yang berhubungan dengan otentikasi
type AuthHandler struct {
	authService service.AuthService
	validator   *validator.Validate
}

// NewAuthHandler adalah constructor
func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		validator:   validator.New(),
	}
}

// LoginRequest merepresentasikan bentuk JSON yang diharapkan dari client
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"` // Wajib diisi dan format email
	Password string `json:"password" validate:"required,min=8"` // Wajib diisi dan minimal 8 karakter
}

// Login memproses endpoint POST /api/v1/auth/login
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest

	// 1. Parsing JSON Body
	if err := c.BodyParser(&req); err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "Format request tidak valid")
	}

	// 2. Validasi Input sesuai aturan (email dan min password 8)
	if err := h.validator.Struct(req); err != nil {
		var errorMessages []string
		for _, err := range err.(validator.ValidationErrors) {
			errorMessages = append(errorMessages, fmt.Sprintf("Field '%s' tidak valid berdasarkan aturan '%s'", err.Field(), err.Tag()))
		}
		// Mengembalikan HTTP 422 untuk validation error
		return helper.ValidationErrorResponse(c, errorMessages)
	}

	// 3. Panggil Layer Service untuk proses autentikasi (Business Logic)
	token, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		// Mengembalikan HTTP 401 jika credential salah
		return helper.ErrorResponse(c, fiber.StatusUnauthorized, "Email atau password salah")
	}

	// 4. Jika berhasil, buat respons sukses berisi JWT
	data := fiber.Map{
		"access_token": token,
	}
	return helper.SuccessResponse(c, fiber.StatusOK, "Login berhasil", data)
}
