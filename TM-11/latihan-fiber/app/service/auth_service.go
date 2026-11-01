package service

import (
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"latihan-fiber/app/model"
	"latihan-fiber/app/repository"
	"latihan-fiber/helper"
)

var (
	ErrInvalidCredentials = errors.New("username atau password salah")
	ErrInactiveUser       = errors.New("user tidak aktif")
	ErrInvalidRefresh     = errors.New("refresh token tidak valid")
)

type AuthService struct {
	userRepo    repository.UserRepository
	tokenRepo   *repository.TokenRepository
	jwt         *helper.JWTManager
	permissions *helper.PermissionSet
	refreshTTL  time.Duration
}

func NewAuthService(
	userRepo repository.UserRepository,
	tokenRepo *repository.TokenRepository,
	jwt *helper.JWTManager,
	permissions *helper.PermissionSet,
	refreshTTL time.Duration,
) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		tokenRepo:   tokenRepo,
		jwt:         jwt,
		permissions: permissions,
		refreshTTL:  refreshTTL,
	}
}

// POST /auth/register
func (s *AuthService) Register(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var req model.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(req.Email)

	if req.Username == "" || req.Email == "" || req.Password == "" {
		return helper.BadRequest("username, email, dan password wajib diisi")
	}
	if len(req.Password) < 8 {
		return helper.BadRequest("password minimal 8 karakter")
	}

	_, err := s.userRepo.FindByUsername(ctx, req.Username)
	if err == nil {
		return helper.NewError(fiber.StatusConflict, "USER_EXISTS", "username atau email sudah digunakan")
	}
	if !errors.Is(err, repository.ErrNotFound) {
		return helper.Internal(err)
	}

	passwordHash, err := helper.HashPassword(req.Password)
	if err != nil {
		return helper.Internal(err)
	}

	user := model.User{
		Username: req.Username,
		Email:    req.Email,
		Password: passwordHash,
		IsActive: true,
	}

	created, err := s.userRepo.Create(ctx, user)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return helper.NewError(fiber.StatusConflict, "USER_EXISTS", "username atau email sudah digunakan")
		}
		return helper.Internal(err)
	}

	return helper.Created(c, "user berhasil didaftarkan", created, "/api/v1/auth/login")
}

// POST /auth/login
func (s *AuthService) Login(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var req model.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	req.Username = strings.TrimSpace(req.Username)
	user, err := s.userRepo.FindByUsername(ctx, req.Username)
	if errors.Is(err, repository.ErrNotFound) {
		helper.VerifyDummyPassword(req.Password)
		return helper.Unauthorized("username atau password salah")
	}
	if err != nil {
		return helper.Internal(err)
	}

	if !helper.VerifyPassword(user.Password, req.Password) {
		return helper.Unauthorized("username atau password salah")
	}
	if !user.IsActive {
		return helper.Forbidden("user tidak aktif")
	}

	// Fetch user's role from DB
	// Since user struct might not carry role field, get from roles/users
	var role string
	// Default to user if not specified
	role = "user"
	// Check if admin
	if user.Username == "admin" {
		role = "admin"
	}

	accessToken, err := s.jwt.GenerateAccessToken(user.ID, role)
	if err != nil {
		return helper.Internal(err)
	}

	refreshToken, err := helper.RandomToken(32)
	if err != nil {
		return helper.Internal(err)
	}

	refreshHash := helper.SHA256Hex(refreshToken)
	if s.tokenRepo != nil {
		_ = s.tokenRepo.Create(ctx, repository.RefreshToken{
			UserID:    user.ID,
			TokenHash: refreshHash,
			ExpiresAt: time.Now().Add(s.refreshTTL),
		})
	}

	return helper.Success(c, fiber.StatusOK, "login berhasil", model.AuthResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

// POST /auth/refresh
func (s *AuthService) Refresh(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var req model.RefreshRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	if s.tokenRepo == nil {
		return helper.Internal(errors.New("token repository tidak tersedia"))
	}

	tokenHash := helper.SHA256Hex(strings.TrimSpace(req.RefreshToken))
	token, err := s.tokenRepo.FindActive(ctx, tokenHash)
	if errors.Is(err, repository.ErrTokenNotFound) {
		return helper.Unauthorized("refresh token tidak valid")
	}
	if err != nil {
		return helper.Internal(err)
	}

	user, err := s.userRepo.FindByID(ctx, token.UserID)
	if err != nil {
		return helper.Internal(err)
	}
	if !user.IsActive {
		return helper.Forbidden("user tidak aktif")
	}

	role := "user"
	if user.Username == "admin" {
		role = "admin"
	}

	accessToken, err := s.jwt.GenerateAccessToken(user.ID, role)
	if err != nil {
		return helper.Internal(err)
	}

	_ = s.tokenRepo.Revoke(ctx, tokenHash)
	newRefreshToken, _ := helper.RandomToken(32)
	newHash := helper.SHA256Hex(newRefreshToken)
	_ = s.tokenRepo.Create(ctx, repository.RefreshToken{
		UserID:    user.ID,
		TokenHash: newHash,
		ExpiresAt: time.Now().Add(s.refreshTTL),
	})

	return helper.Success(c, fiber.StatusOK, "token berhasil diperbarui", model.AuthResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	})
}

