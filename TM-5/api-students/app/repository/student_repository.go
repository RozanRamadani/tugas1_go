package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"api-students/app/model"
)

// Sentinel error untuk kondisi tertentu.
var (
	ErrNotFound  = errors.New("student tidak ditemukan")
	ErrDuplicate = errors.New("data student sudah ada")
)

// StudentRepository adalah kontrak repository student.
//
// Handler tidak perlu mengetahui bagaimana data disimpan.
// Handler cukup memanggil method yang tersedia di interface ini.
type StudentRepository interface {
	FindAll(ctx context.Context, q model.ListQuery) ([]model.Student, int, error)
	FindByID(ctx context.Context, id string) (model.Student, error)
	Create(ctx context.Context, student model.Student) (model.Student, error)
	Update(ctx context.Context, id string, student model.Student) (model.Student, error)
	Delete(ctx context.Context, id string) error
}

// postgresStudentRepository adalah implementasi
// StudentRepository menggunakan PostgreSQL.
type postgresStudentRepository struct {
	pool *pgxpool.Pool
}

// NewStudentRepository membuat repository baru.
func NewStudentRepository(pool *pgxpool.Pool) StudentRepository {
	return &postgresStudentRepository{
		pool: pool,
	}
}

// ============================================================
// FIND ALL
// ============================================================

func (r *postgresStudentRepository) FindAll(
	ctx context.Context,
	q model.ListQuery,
) ([]model.Student, int, error) {

	// --------------------------------------------------------
	// Daftar putih kolom untuk ORDER BY.
	// Jangan langsung memasukkan q.Sort ke SQL.
	// --------------------------------------------------------

	allowedSort := map[string]string{
		"id":         "id",
		"nim":        "nim",
		"name":       "name",
		"grade":      "grade",
		"created_at": "created_at",
	}

	sortColumn, ok := allowedSort[q.Sort]

	if !ok {
		sortColumn = "id"
	}

	order := "ASC"

	if strings.ToLower(q.Order) == "desc" {
		order = "DESC"
	}

	// --------------------------------------------------------
	// WHERE dinamis
	// --------------------------------------------------------

	where := []string{}
	args := []any{}

	argNumber := 1

	// Filtering berdasarkan is_active.
	if q.IsActive != nil {
		where = append(
			where,
			fmt.Sprintf("is_active = $%d", argNumber),
		)

		args = append(args, *q.IsActive)

		argNumber++
	}

	// Pencarian berdasarkan nama.
	if q.Search != "" {
		where = append(
			where,
			fmt.Sprintf("name ILIKE $%d", argNumber),
		)

		args = append(args, "%"+q.Search+"%")

		argNumber++
	}

	whereSQL := ""

	if len(where) > 0 {
		whereSQL = "WHERE " + strings.Join(where, " AND ")
	}

	// --------------------------------------------------------
	// COUNT
	// --------------------------------------------------------
	//
	// COUNT harus menggunakan WHERE yang sama dengan query
	// utama agar meta.total sesuai dengan data hasil filter.
	//

	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM students
		%s
	`, whereSQL)

	var total int

	if err := r.pool.QueryRow(
		ctx,
		countQuery,
		args...,
	).Scan(&total); err != nil {
		return nil, 0, err
	}

	// --------------------------------------------------------
	// SELECT DATA
	// --------------------------------------------------------

	dataQuery := fmt.Sprintf(`
		SELECT
			id,
			nim,
			name,
			grade,
			is_active,
			created_at
		FROM students
		%s
		ORDER BY %s %s
		LIMIT $%d
		OFFSET $%d
	`,
		whereSQL,
		sortColumn,
		order,
		argNumber,
		argNumber+1,
	)

	dataArgs := append(
		append([]any{}, args...),
		q.Limit,
		q.Offset(),
	)

	rows, err := r.pool.Query(
		ctx,
		dataQuery,
		dataArgs...,
	)

	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()

	students := []model.Student{}

	for rows.Next() {

		var student model.Student

		err := rows.Scan(
			&student.ID,
			&student.NIM,
			&student.Name,
			&student.Grade,
			&student.IsActive,
			&student.CreatedAt,
		)

		if err != nil {
			return nil, 0, err
		}

		students = append(students, student)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return students, total, nil
}

// ============================================================
// FIND BY ID
// ============================================================

func (r *postgresStudentRepository) FindByID(
	ctx context.Context,
	id string,
) (model.Student, error) {

	var student model.Student

	query := `
		SELECT
			id,
			nim,
			name,
			grade,
			is_active,
			created_at
		FROM students
		WHERE id = $1
	`

	err := r.pool.QueryRow(
		ctx,
		query,
		id,
	).Scan(
		&student.ID,
		&student.NIM,
		&student.Name,
		&student.Grade,
		&student.IsActive,
		&student.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return model.Student{}, ErrNotFound
	}

	if err != nil {
		return model.Student{}, err
	}

	return student, nil
}

// ============================================================
// CREATE
// ============================================================

func (r *postgresStudentRepository) Create(
	ctx context.Context,
	student model.Student,
) (model.Student, error) {

	query := `
		INSERT INTO students (
			id,
			nim,
			name,
			grade,
			is_active
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING
			id,
			nim,
			name,
			grade,
			is_active,
			created_at
	`

	var result model.Student

	err := r.pool.QueryRow(
		ctx,
		query,
		student.ID,
		student.NIM,
		student.Name,
		student.Grade,
		student.IsActive,
	).Scan(
		&result.ID,
		&result.NIM,
		&result.Name,
		&result.Grade,
		&result.IsActive,
		&result.CreatedAt,
	)

	if isDuplicateError(err) {
		return model.Student{}, ErrDuplicate
	}

	if err != nil {
		return model.Student{}, err
	}

	return result, nil
}

// ============================================================
// UPDATE
// ============================================================

func (r *postgresStudentRepository) Update(
	ctx context.Context,
	id string,
	student model.Student,
) (model.Student, error) {

	query := `
		UPDATE students
		SET
			nim = $1,
			name = $2,
			grade = $3,
			is_active = $4
		WHERE id = $5
		RETURNING
			id,
			nim,
			name,
			grade,
			is_active,
			created_at
	`

	var result model.Student

	err := r.pool.QueryRow(
		ctx,
		query,
		student.NIM,
		student.Name,
		student.Grade,
		student.IsActive,
		id,
	).Scan(
		&result.ID,
		&result.NIM,
		&result.Name,
		&result.Grade,
		&result.IsActive,
		&result.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return model.Student{}, ErrNotFound
	}

	if isDuplicateError(err) {
		return model.Student{}, ErrDuplicate
	}

	if err != nil {
		return model.Student{}, err
	}

	return result, nil
}

// ============================================================
// DELETE
// ============================================================

func (r *postgresStudentRepository) Delete(
	ctx context.Context,
	id string,
) error {

	commandTag, err := r.pool.Exec(
		ctx,
		`DELETE FROM students WHERE id = $1`,
		id,
	)

	if err != nil {
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// ============================================================
// DUPLICATE ERROR
// ============================================================

// isDuplicateError mengecek apakah PostgreSQL
// mengembalikan error unique violation.
func isDuplicateError(err error) bool {

	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}

	return false
}
