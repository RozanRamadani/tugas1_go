package service

import (
	"context"
	"errors"
	"strings"

	"api-students/app/model"
	"api-students/app/repository"
)

type StudentService struct {
	repo repository.StudentRepository
}

func NewStudentService(
	repo repository.StudentRepository,
) *StudentService {
	return &StudentService{
		repo: repo,
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
) (model.Student, error) {

	if strings.TrimSpace(id) == "" {
		return model.Student{}, errors.New("id tidak boleh kosong")
	}

	return s.repo.FindByID(ctx, id)
}

// ============================================================
// CREATE / POST
// ============================================================

func (s *StudentService) Create(
	ctx context.Context,
	req model.CreateStudentRequest,
) (model.Student, error) {

	errs := ValidateCreate(req)

	if len(errs) > 0 {
		return model.Student{}, ValidationError{
			Errors: errs,
		}
	}

	student := model.Student{
		ID:       strings.TrimSpace(req.ID),
		NIM:      strings.TrimSpace(req.NIM),
		Name:     strings.TrimSpace(req.Name),
		Grade:    req.Grade,
		IsActive: req.IsActive,
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
) (model.Student, error) {

	if strings.TrimSpace(id) == "" {
		return model.Student{}, errors.New("id tidak boleh kosong")
	}

	errs := ValidateReplace(req)

	if len(errs) > 0 {
		return model.Student{}, ValidationError{
			Errors: errs,
		}
	}

	student := model.Student{
		ID:       id,
		NIM:      strings.TrimSpace(req.NIM),
		Name:     strings.TrimSpace(req.Name),
		Grade:    req.Grade,
		IsActive: req.IsActive,
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
) (model.Student, error) {

	if strings.TrimSpace(id) == "" {
		return model.Student{}, errors.New("id tidak boleh kosong")
	}

	student, err := s.repo.FindByID(ctx, id)

	if err != nil {
		return model.Student{}, err
	}

	student, errs := ApplyPatch(student, req)

	if len(errs) > 0 {
		return model.Student{}, ValidationError{
			Errors: errs,
		}
	}

	return s.repo.Update(ctx, id, student)
}

// ============================================================
// DELETE
// ============================================================

func (s *StudentService) Delete(
	ctx context.Context,
	id string,
) error {

	if strings.TrimSpace(id) == "" {
		return errors.New("id tidak boleh kosong")
	}

	return s.repo.Delete(ctx, id)
}