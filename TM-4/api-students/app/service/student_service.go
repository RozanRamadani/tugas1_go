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
// CREATE
// ============================================================

func (s *StudentService) Create(
	ctx context.Context,
	req model.CreateStudentRequest,
) (model.Student, error) {

	if err := ValidateCreate(req); err != nil {
		return model.Student{}, err
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

	if err := ValidateReplace(req); err != nil {
		return model.Student{}, err
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

	if req.NIM != nil {
		if strings.TrimSpace(*req.NIM) == "" {
			return model.Student{}, errors.New("nim tidak boleh kosong")
		}

		student.NIM = strings.TrimSpace(*req.NIM)
	}

	if req.Name != nil {
		if strings.TrimSpace(*req.Name) == "" {
			return model.Student{}, errors.New("name tidak boleh kosong")
		}

		student.Name = strings.TrimSpace(*req.Name)
	}

	if req.Grade != nil {
		if *req.Grade < 0 || *req.Grade > 100 {
			return model.Student{}, errors.New(
				"grade harus berada di antara 0 dan 100",
			)
		}

		student.Grade = *req.Grade
	}

	if req.IsActive != nil {
		student.IsActive = *req.IsActive
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