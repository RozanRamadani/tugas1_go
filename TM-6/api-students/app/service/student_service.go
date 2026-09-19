package service

import (
	"context"
	"errors"
	"strings"

	"api-students/app/model"
	"api-students/app/repository"
	"api-students/helper"
)

var ErrForbidden = errors.New("akses ditolak")

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

// ============================================================
// LIST
// ============================================================

func (s *StudentService) List(
	ctx context.Context,
	q model.ListQuery,
) ([]model.Student, int, error) {
	return s.repo.FindAll(ctx, q)
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
		return model.Student{}, err
	}

	// Owner boleh membaca data sendiri.
	// Non-owner membutuhkan student:read:any.
	if !CanAccessStudent(
		current,
		student.OwnerID,
		s.perms,
		"student:read:any",
	) {
		return model.Student{}, ErrForbidden
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

	errs := ValidateCreate(req)

	if len(errs) > 0 {
		return model.Student{}, ValidationError{
			Fields: errs,
		}
	}

	student := model.Student{
		ID:       strings.TrimSpace(req.ID),
		NIM:      strings.TrimSpace(req.NIM),
		Name:     strings.TrimSpace(req.Name),
		Grade:    req.Grade,
		IsActive: req.IsActive,

		// owner_id SELALU berasal dari user yang login.
		// Tidak berasal dari request body.
		OwnerID: current.UserID,
	}

	return s.repo.Create(ctx, student)
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
		return model.Student{}, err
	}

	// Owner boleh update data sendiri.
	// Non-owner membutuhkan student:update:any.
	if !CanAccessStudent(
		current,
		existing.OwnerID,
		s.perms,
		"student:update:any",
	) {
		return model.Student{}, ErrForbidden
	}

	errs := ValidateReplace(req)

	if len(errs) > 0 {
		return model.Student{}, ValidationError{
			Fields: errs,
		}
	}

	student := model.Student{
		ID:       existing.ID,
		NIM:      strings.TrimSpace(req.NIM),
		Name:     strings.TrimSpace(req.Name),
		Grade:    req.Grade,
		IsActive: req.IsActive,

		// Owner tidak boleh berubah melalui PUT.
		OwnerID: existing.OwnerID,
	}

	return s.repo.Update(ctx, id, student)
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
		return model.Student{}, err
	}

	// Owner boleh update data sendiri.
	// Non-owner membutuhkan student:update:any.
	if !CanAccessStudent(
		current,
		student.OwnerID,
		s.perms,
		"student:update:any",
	) {
		return model.Student{}, ErrForbidden
	}

	// Business rule PATCH.
	student, errs := ApplyPatch(student, req)

	if len(errs) > 0 {
		return model.Student{}, ValidationError{
			Fields: errs,
		}
	}

	// ApplyPatch tidak mengubah owner_id.
	return s.repo.Update(ctx, id, student)
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
		return ErrForbidden
	}

	return s.repo.Delete(ctx, id)
}
