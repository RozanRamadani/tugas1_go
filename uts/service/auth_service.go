package service

import (
	"errors"
	"time"

	"siakad-mini/config"
	"siakad-mini/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Login(email, password string) (string, error)
}

type authService struct {
	userRepo repository.UserRepository
}

func NewAuthService(userRepo repository.UserRepository) AuthService {
	return &authService{userRepo}
}

func (s *authService) Login(email, password string) (string, error) {
	// 1. Ambil data user dari database melalui Repository
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	// 2. Bandingkan password plain-text dari request dengan hash yang ada di database
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	// 3. Jika password cocok, buat JWT Token
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // Kadaluarsa dalam 24 jam
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secret := config.GetEnv("JWT_SECRET", "supersecretkey")

	// 4. Tandatangani token dengan JWT_SECRET
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", errors.New("failed to generate token")
	}

	return tokenString, nil
}
