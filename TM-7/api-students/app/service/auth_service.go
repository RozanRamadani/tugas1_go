package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"api-students/app/model"
	"api-students/app/repository"
	"api-students/helper"
)

var (
	ErrInvalidCredentials = errors.New("username atau password salah")
	ErrInactiveUser       = errors.New("user tidak aktif")
	ErrInvalidRefresh     = errors.New("refresh token tidak valid")
)

type AuthService struct {
	userRepo  *repository.UserRepository
	tokenRepo *repository.TokenRepository
	jwt       *helper.JWTManager
}

func NewAuthService(
	userRepo *repository.UserRepository,
	tokenRepo *repository.TokenRepository,
	jwt *helper.JWTManager,
) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		jwt:       jwt,
	}
}

// ============================================================
// REGISTER
// ============================================================

func (s *AuthService) Register(
	ctx context.Context,
	req model.RegisterRequest,
) (model.User, error) {

	username := strings.TrimSpace(req.Username)
	email := strings.TrimSpace(req.Email)

	if username == "" {
		return model.User{}, errors.New("username wajib diisi")
	}

	if email == "" {
		return model.User{}, errors.New("email wajib diisi")
	}

	if req.Password == "" {
		return model.User{}, errors.New("password wajib diisi")
	}

	if len(req.Password) < 8 {
		return model.User{}, errors.New(
			"password minimal 8 karakter",
		)
	}

	// Cek username apakah sudah digunakan.
	_, err := s.userRepo.FindByUsername(ctx, username)

	if err == nil {
		return model.User{}, repository.ErrUserExists
	}

	if !errors.Is(err, repository.ErrUserNotFound) {
		return model.User{}, err
	}

	// Hash password menggunakan bcrypt.
	passwordHash, err := helper.HashPassword(req.Password)

	if err != nil {
		return model.User{}, err
	}

	user := model.User{
		Username: username,
		Email:    email,
		Password: passwordHash,
		Role:     "user",
		IsActive: true,
	}

	return s.userRepo.Create(ctx, user)
}

// ============================================================
// LOGIN
// ============================================================

func (s *AuthService) Login(
	ctx context.Context,
	req model.LoginRequest,
) (model.AuthResponse, error) {

	username := strings.TrimSpace(req.Username)

	user, err := s.userRepo.FindByUsername(
		ctx,
		username,
	)

	if errors.Is(err, repository.ErrUserNotFound) {
		// Tetap melakukan bcrypt compare agar perbedaan waktu
		// antara username valid dan tidak valid tidak terlalu besar.
		helper.VerifyDummyPassword(req.Password)

		return model.AuthResponse{}, ErrInvalidCredentials
	}

	if err != nil {
		return model.AuthResponse{}, err
	}

	if !helper.VerifyPassword(
		user.Password,
		req.Password,
	) {
		return model.AuthResponse{}, ErrInvalidCredentials
	}

	if !user.IsActive {
		return model.AuthResponse{}, ErrInactiveUser
	}

	accessToken, err := s.jwt.GenerateAccessToken(
		user.ID,
		user.Role,
	)

	if err != nil {
		return model.AuthResponse{}, err
	}

	refreshToken, err := helper.RandomToken(32)

	if err != nil {
		return model.AuthResponse{}, err
	}

	refreshHash := helper.SHA256Hex(refreshToken)

	err = s.tokenRepo.Create(
		ctx,
		repository.RefreshToken{
			UserID:    user.ID,
			TokenHash: refreshHash,
			ExpiresAt: time.Now().Add(s.jwt.RefreshTTL),
		},
	)

	if err != nil {
		return model.AuthResponse{}, err
	}

	return model.AuthResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// ============================================================
// REFRESH
// ============================================================

func (s *AuthService) Refresh(
	ctx context.Context,
	refreshToken string,
) (model.AuthResponse, error) {

	refreshToken = strings.TrimSpace(refreshToken)

	if refreshToken == "" {
		return model.AuthResponse{}, ErrInvalidRefresh
	}

	tokenHash := helper.SHA256Hex(refreshToken)

	token, err := s.tokenRepo.FindActive(
		ctx,
		tokenHash,
	)

	if errors.Is(err, repository.ErrTokenNotFound) {
		return model.AuthResponse{}, ErrInvalidRefresh
	}

	if err != nil {
		return model.AuthResponse{}, err
	}

	user, err := s.userRepo.FindByID(
		ctx,
		token.UserID,
	)

	if err != nil {
		return model.AuthResponse{}, err
	}

	if !user.IsActive {
		return model.AuthResponse{}, ErrInactiveUser
	}

	accessToken, err := s.jwt.GenerateAccessToken(
		user.ID,
		user.Role,
	)

	if err != nil {
		return model.AuthResponse{}, err
	}

	// Refresh token rotation.
	err = s.tokenRepo.Revoke(
		ctx,
		tokenHash,
	)

	if err != nil {
		return model.AuthResponse{}, err
	}

	newRefreshToken, err := helper.RandomToken(32)

	if err != nil {
		return model.AuthResponse{}, err
	}

	newRefreshHash := helper.SHA256Hex(newRefreshToken)

	err = s.tokenRepo.Create(
		ctx,
		repository.RefreshToken{
			UserID:    user.ID,
			TokenHash: newRefreshHash,
			ExpiresAt: time.Now().Add(s.jwt.RefreshTTL),
		},
	)

	if err != nil {
		return model.AuthResponse{}, err
	}

	return model.AuthResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

// ============================================================
// LOGOUT
// ============================================================

func (s *AuthService) Logout(
	ctx context.Context,
	refreshToken string,
) error {

	refreshToken = strings.TrimSpace(refreshToken)

	if refreshToken == "" {
		return ErrInvalidRefresh
	}

	tokenHash := helper.SHA256Hex(refreshToken)

	err := s.tokenRepo.Revoke(
		ctx,
		tokenHash,
	)

	if errors.Is(err, repository.ErrTokenNotFound) {
		return ErrInvalidRefresh
	}

	return err
}

// ============================================================
// ME (PROFILE)
// ============================================================

func (s *AuthService) Me(
	ctx context.Context,
	userID int,
) (model.User, error) {
	return s.userRepo.FindByID(ctx, userID)
}
