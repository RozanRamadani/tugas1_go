package helper

import (
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("token tidak valid")
	ErrExpiredToken = errors.New("token sudah kedaluwarsa")
)

type JWTManager struct {
	Secret     string
	Issuer     string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

type Claims struct {
	UserID int    `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func NewJWTManager(
	secret string,
	issuer string,
	accessTTL time.Duration,
	refreshTTL time.Duration,
) *JWTManager {
	return &JWTManager{
		Secret:     secret,
		Issuer:     issuer,
		AccessTTL:  accessTTL,
		RefreshTTL: refreshTTL,
	}
}

// GenerateAccessToken membuat JWT access token.
func (j *JWTManager) GenerateAccessToken(
	userID int,
	role string,
) (string, error) {

	now := time.Now()

	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.Itoa(userID),
			Issuer:    j.Issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(j.AccessTTL)),
		},
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	return token.SignedString([]byte(j.Secret))
}

// ParseAccessToken memvalidasi JWT access token.
func (j *JWTManager) ParseAccessToken(
	tokenString string,
) (*Claims, error) {

	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {

			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("algoritma JWT tidak valid")
			}

			return []byte(j.Secret), nil
		},
		jwt.WithIssuer(j.Issuer),
		jwt.WithExpirationRequired(),
	)

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)

	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
