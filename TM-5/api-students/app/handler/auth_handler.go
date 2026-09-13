package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"api-students/app/model"
	"api-students/app/repository"
	"api-students/app/service"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(
	authService *service.AuthService,
) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// ============================================================
// POST /auth/register
// ============================================================

func (h *AuthHandler) Register(c *fiber.Ctx) error {

	var req model.RegisterRequest

	if err := c.BodyParser(&req); err != nil {
		return fail(
			c,
			fiber.StatusBadRequest,
			"body harus berupa JSON yang valid",
		)
	}

	user, err := h.authService.Register(
		c.Context(),
		req,
	)

	if errors.Is(err, repository.ErrUserExists) {
		return fail(
			c,
			fiber.StatusConflict,
			"username atau email sudah digunakan",
		)
	}

	if err != nil {
		return fail(
			c,
			fiber.StatusBadRequest,
			err.Error(),
		)
	}

	return created(
		c,
		"user berhasil didaftarkan",
		user,
		"/api/v1/auth/login",
	)
}

// ============================================================
// POST /auth/login
// ============================================================

func (h *AuthHandler) Login(c *fiber.Ctx) error {

	var req model.LoginRequest

	if err := c.BodyParser(&req); err != nil {
		return fail(
			c,
			fiber.StatusBadRequest,
			"body harus berupa JSON yang valid",
		)
	}

	result, err := h.authService.Login(
		c.Context(),
		req,
	)

	if errors.Is(err, service.ErrInvalidCredentials) {
		return fail(
			c,
			fiber.StatusUnauthorized,
			"username atau password salah",
		)
	}

	if errors.Is(err, service.ErrInactiveUser) {
		return fail(
			c,
			fiber.StatusForbidden,
			"user tidak aktif",
		)
	}

	if err != nil {
		return fail(
			c,
			fiber.StatusInternalServerError,
			"gagal melakukan login",
		)
	}

	return ok(
		c,
		"login berhasil",
		result,
	)
}

// ============================================================
// POST /auth/refresh
// ============================================================

func (h *AuthHandler) Refresh(c *fiber.Ctx) error {

	var req model.RefreshRequest

	if err := c.BodyParser(&req); err != nil {
		return fail(
			c,
			fiber.StatusBadRequest,
			"body harus berupa JSON yang valid",
		)
	}

	result, err := h.authService.Refresh(
		c.Context(),
		req.RefreshToken,
	)

	if errors.Is(err, service.ErrInvalidRefresh) {
		return fail(
			c,
			fiber.StatusUnauthorized,
			"refresh token tidak valid",
		)
	}

	if errors.Is(err, service.ErrInactiveUser) {
		return fail(
			c,
			fiber.StatusForbidden,
			"user tidak aktif",
		)
	}

	if err != nil {
		return fail(
			c,
			fiber.StatusInternalServerError,
			"gagal melakukan refresh token",
		)
	}

	return ok(
		c,
		"token berhasil diperbarui",
		result,
	)
}

// ============================================================
// POST /auth/logout
// ============================================================

func (h *AuthHandler) Logout(c *fiber.Ctx) error {

	var req model.RefreshRequest

	if err := c.BodyParser(&req); err != nil {
		return fail(
			c,
			fiber.StatusBadRequest,
			"body harus berupa JSON yang valid",
		)
	}

	err := h.authService.Logout(
		c.Context(),
		req.RefreshToken,
	)

	if errors.Is(err, service.ErrInvalidRefresh) {
		return fail(
			c,
			fiber.StatusUnauthorized,
			"refresh token tidak valid",
		)
	}

	if err != nil {
		return fail(
			c,
			fiber.StatusInternalServerError,
			"gagal melakukan logout",
		)
	}

	return ok(
		c,
		"logout berhasil",
		nil,
	)
}
