package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrTokenNotFound = errors.New("refresh token tidak ditemukan")

type RefreshToken struct {
	ID        int
	UserID    int
	TokenHash string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

type TokenRepository struct {
	pool *pgxpool.Pool
}

func NewTokenRepository(pool *pgxpool.Pool) *TokenRepository {
	return &TokenRepository{
		pool: pool,
	}
}

// ============================================================
// CREATE REFRESH TOKEN
// ============================================================

func (r *TokenRepository) Create(
	ctx context.Context,
	token RefreshToken,
) error {

	query := `
		INSERT INTO refresh_tokens (
			user_id,
			token_hash,
			expires_at
		)
		VALUES ($1, $2, $3)
	`

	_, err := r.pool.Exec(
		ctx,
		query,
		token.UserID,
		token.TokenHash,
		token.ExpiresAt,
	)

	return err
}

// ============================================================
// FIND ACTIVE TOKEN
// ============================================================

func (r *TokenRepository) FindActive(
	ctx context.Context,
	tokenHash string,
) (RefreshToken, error) {

	query := `
		SELECT
			id,
			user_id,
			token_hash,
			expires_at,
			revoked_at,
			created_at
		FROM refresh_tokens
		WHERE token_hash = $1
		  AND revoked_at IS NULL
		  AND expires_at > NOW()
	`

	var token RefreshToken

	err := r.pool.QueryRow(
		ctx,
		query,
		tokenHash,
	).Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&token.ExpiresAt,
		&token.RevokedAt,
		&token.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return RefreshToken{}, ErrTokenNotFound
	}

	if err != nil {
		return RefreshToken{}, err
	}

	return token, nil
}

// ============================================================
// REVOKE TOKEN
// ============================================================

func (r *TokenRepository) Revoke(
	ctx context.Context,
	tokenHash string,
) error {

	query := `
		UPDATE refresh_tokens
		SET revoked_at = NOW()
		WHERE token_hash = $1
		  AND revoked_at IS NULL
	`

	result, err := r.pool.Exec(
		ctx,
		query,
		tokenHash,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrTokenNotFound
	}

	return nil
}

// ============================================================
// REVOKE ALL USER TOKENS
// ============================================================

func (r *TokenRepository) RevokeAll(
	ctx context.Context,
	userID int,
) error {

	query := `
		UPDATE refresh_tokens
		SET revoked_at = NOW()
		WHERE user_id = $1
		  AND revoked_at IS NULL
	`

	_, err := r.pool.Exec(
		ctx,
		query,
		userID,
	)

	return err
}
