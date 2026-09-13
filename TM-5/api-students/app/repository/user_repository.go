package repository

import (
	"context"
	"errors"

	"api-students/app/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrUserNotFound = errors.New("user tidak ditemukan")
	ErrUserExists   = errors.New("username atau email sudah digunakan")
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		pool: pool,
	}
}

// ============================================================
// CREATE USER
// ============================================================

func (r *UserRepository) Create(
	ctx context.Context,
	user model.User,
) (model.User, error) {

	query := `
		INSERT INTO users (
			username,
			email,
			password,
			role,
			is_active
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING
			id,
			username,
			email,
			role,
			is_active,
			created_at
	`

	var result model.User

	err := r.pool.QueryRow(
		ctx,
		query,
		user.Username,
		user.Email,
		user.Password,
		user.Role,
		user.IsActive,
	).Scan(
		&result.ID,
		&result.Username,
		&result.Email,
		&result.Role,
		&result.IsActive,
		&result.CreatedAt,
	)

	if err != nil {
		return model.User{}, err
	}

	return result, nil
}

// ============================================================
// FIND BY USERNAME
// ============================================================

func (r *UserRepository) FindByUsername(
	ctx context.Context,
	username string,
) (model.User, error) {

	query := `
		SELECT
			id,
			username,
			email,
			password,
			role,
			is_active,
			created_at
		FROM users
		WHERE username = $1
	`

	var user model.User

	err := r.pool.QueryRow(
		ctx,
		query,
		username,
	).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return model.User{}, ErrUserNotFound
	}

	if err != nil {
		return model.User{}, err
	}

	return user, nil
}

// ============================================================
// FIND BY ID
// ============================================================

func (r *UserRepository) FindByID(
	ctx context.Context,
	id int,
) (model.User, error) {

	query := `
		SELECT
			id,
			username,
			email,
			password,
			role,
			is_active,
			created_at
		FROM users
		WHERE id = $1
	`

	var user model.User

	err := r.pool.QueryRow(
		ctx,
		query,
		id,
	).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return model.User{}, ErrUserNotFound
	}

	if err != nil {
		return model.User{}, err
	}

	return user, nil
}
