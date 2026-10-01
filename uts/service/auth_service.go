package service

import (
	"errors"
	"time"

	"siakad-mini/config"
	"siakad-mini/domain"
	"siakad-mini/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Login(email, password string) (string, error)
	GetMe(userID uint, role string) (interface{}, error)
}

type authService struct {
	userRepo    repository.UserRepository
	studentRepo repository.StudentRepository
}

func NewAuthService(userRepo repository.UserRepository, studentRepo repository.StudentRepository) AuthService {
	return &authService{userRepo, studentRepo}
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

	// 2.5 Jika role mahasiswa, periksa apakah data student aktif (tidak di-soft-delete)
	if user.Role == domain.RoleMahasiswa {
		_, err := s.studentRepo.FindByUserID(user.ID)
		if err != nil {
			return "", errors.New("invalid credentials")
		}
	}

	// 3. Jika password cocok dan akun aktif, buat JWT Token
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

func (s *authService) GetMe(userID uint, role string) (interface{}, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	authMe := domain.AuthMeResponse{
		ID:    user.ID,
		Email: user.Email,
		Role:  user.Role,
	}

	if role == string(domain.RoleAdmin) {
		return authMe, nil
	}

	if role == string(domain.RoleMahasiswa) {
		student, err := s.studentRepo.FindByUserID(userID)
		if err != nil {
			return nil, errors.New("student data not found")
		}

		return domain.AuthMeStudentResponse{
			AuthMeResponse: authMe,
			NIM:            student.NIM,
			Nama:           student.Nama,
			Prodi:          student.Prodi,
			Angkatan:       student.Angkatan,
		}, nil
	}

	return nil, errors.New("invalid role")
}
