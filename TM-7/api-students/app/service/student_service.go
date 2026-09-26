package service

import (
	"context"
	"errors"
	"strings"

	"api-students/app/model"
	"api-students/app/repository"
	"api-students/helper"
)

type StudentService struct {
	repo  repository.StudentRepository
	perms *helper.PermissionSet
}

func NewStudentService(
	repo repository.StudentRepository,
	perms *helper.PermissionSet,
) *StudentService {
	return &StudentService{
		repo:  repo,
		perms: perms,
	}
}

// translateError memetakan error repository ke AppError terpusat.
func translateError(err error) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return helper.NotFound("student tidak ditemukan")
	case errors.Is(err, repository.ErrDuplicate):
		return helper.Conflict("student sudah ada")
	default:
		return helper.Internal(err)
	}
}

// ============================================================
// LIST
// ============================================================

func (s *StudentService) List(
	ctx context.Context,
	q model.CursorQuery,
) ([]model.Student, *model.CursorMeta, error) {

	rows, err := s.repo.FindAfterCursor(ctx, q)
	if err != nil {
		return nil, nil, translateError(err)
	}

	hasMore := len(rows) > q.Limit
	if hasMore {
		rows = rows[:q.Limit]
	}

	meta := &model.CursorMeta{Limit: q.Limit, HasMore: hasMore}
	if hasMore && len(rows) > 0 {
		last := rows[len(rows)-1]
		meta.NextCursor = helper.EncodeCursor(last.CreatedAt, last.ID)
	}

	return rows, meta, nil
}

// ============================================================
// GET
// ============================================================

func (s *StudentService) Get(
	ctx context.Context,
	id string,
	current model.AuthUser,
) (model.Student, error) {

	if strings.TrimSpace(id) == "" {
		return model.Student{}, errors.New("id tidak boleh kosong")
	}

	student, err := s.repo.FindByID(ctx, id)

	if err != nil {
		return model.Student{}, translateError(err)
	}

	// Owner boleh membaca data sendiri.
	// Non-owner membutuhkan student:read:any.
	if !CanAccessStudent(
		current,
		student.OwnerID,
		s.perms,
		"student:read:any",
	) {
		return model.Student{}, helper.Forbidden("akses ditolak")
	}

	return student, nil
}

// ============================================================
// CREATE / POST
// ============================================================

func (s *StudentService) Create(
	ctx context.Context,
	req model.CreateStudentRequest,
	current model.AuthUser,
) (model.Student, error) {

	req.ID = strings.TrimSpace(req.ID)
	req.NIM = strings.TrimSpace(req.NIM)
	req.Name = strings.TrimSpace(req.Name)

	errs := helper.ValidateStruct(req)

	if len(errs) > 0 {
		return model.Student{}, helper.Validation(errs)
	}

	student := model.Student{
		ID:       req.ID,
		NIM:      req.NIM,
		Name:     req.Name,
		Grade:    req.Grade,
		IsActive: req.IsActive,

		// owner_id SELALU berasal dari user yang login.
		// Tidak berasal dari request body.
		OwnerID: current.UserID,
	}

	student, err := s.repo.Create(ctx, student)
	if err != nil {
		return model.Student{}, translateError(err)
	}
	return student, nil
}

// ============================================================
// REPLACE / PUT
// ============================================================

func (s *StudentService) Replace(
	ctx context.Context,
	id string,
	req model.ReplaceStudentRequest,
	current model.AuthUser,
) (model.Student, error) {

	if strings.TrimSpace(id) == "" {
		return model.Student{}, errors.New("id tidak boleh kosong")
	}

	// Ambil data lama untuk mengetahui owner.
	existing, err := s.repo.FindByID(ctx, id)

	if err != nil {
		return model.Student{}, translateError(err)
	}

	// Owner boleh update data sendiri.
	// Non-owner membutuhkan student:update:any.
	if !CanAccessStudent(
		current,
		existing.OwnerID,
		s.perms,
		"student:update:any",
	) {
		return model.Student{}, helper.Forbidden("akses ditolak")
	}

	req.NIM = strings.TrimSpace(req.NIM)
	req.Name = strings.TrimSpace(req.Name)

	errs := helper.ValidateStruct(req)

	if len(errs) > 0 {
		return model.Student{}, helper.Validation(errs)
	}

	student := model.Student{
		ID:       existing.ID,
		NIM:      req.NIM,
		Name:     req.Name,
		Grade:    req.Grade,
		IsActive: req.IsActive,

		// Owner tidak boleh berubah melalui PUT.
		OwnerID: existing.OwnerID,
	}

	student, err = s.repo.Update(ctx, id, student)
	if err != nil {
		return model.Student{}, translateError(err)
	}
	return student, nil
}

// ============================================================
// PATCH
// ============================================================

func (s *StudentService) Patch(
	ctx context.Context,
	id string,
	req model.PatchStudentRequest,
	current model.AuthUser,
) (model.Student, error) {

	if strings.TrimSpace(id) == "" {
		return model.Student{}, errors.New("id tidak boleh kosong")
	}

	// Ambil data lama terlebih dahulu.
	student, err := s.repo.FindByID(ctx, id)

	if err != nil {
		return model.Student{}, translateError(err)
	}

	// Owner boleh update data sendiri.
	// Non-owner membutuhkan student:update:any.
	if !CanAccessStudent(
		current,
		student.OwnerID,
		s.perms,
		"student:update:any",
	) {
		return model.Student{}, helper.Forbidden("akses ditolak")
	}

	// Business rule PATCH.
	student, errs := ApplyPatch(student, req)

	if len(errs) > 0 {
		return model.Student{}, helper.Validation(errs)
	}

	// ApplyPatch tidak mengubah owner_id.
	student, err = s.repo.Update(ctx, id, student)
	if err != nil {
		return model.Student{}, translateError(err)
	}
	return student, nil
}

// ============================================================
// DELETE
// ============================================================

func (s *StudentService) Delete(
	ctx context.Context,
	id string,
	current model.AuthUser,
) error {

	if strings.TrimSpace(id) == "" {
		return errors.New("id tidak boleh kosong")
	}

	// Endpoint DELETE sudah dilindungi
	// middleware RequirePermission(student:delete).
	//
	// Pengecekan ulang di service digunakan sebagai
	// defense in depth.
	if !s.perms.Can(current.Role, "student:delete") {
		return helper.Forbidden("akses ditolak")
	}

	err := s.repo.Delete(ctx, id)
	if err != nil {
		return translateError(err)
	}
	return nil
}
