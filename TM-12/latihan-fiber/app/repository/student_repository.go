package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type StudentRepository interface {
	Exists(ctx context.Context, studentID string) (bool, error)
}

type studentPostgresRepository struct {
	db DBTX
}

func NewStudentRepository(db DBTX) StudentRepository {
	return &studentPostgresRepository{db: db}
}

func (r *studentPostgresRepository) Exists(ctx context.Context, studentID string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM students WHERE id = $1)", studentID).Scan(&exists)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("memeriksa keberadaan mahasiswa: %w", err)
	}
	return exists, nil
}
