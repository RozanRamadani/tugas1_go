package service

import (
	"errors"
	"strings"

	"api-students/app/model"
)

var (
	ErrValidation = errors.New("validasi gagal")
)

type ValidationError struct {
	Fields map[string]string
}

func (e *ValidationError) Error() string {
	return "validasi gagal"
}

// ============================================================
// VALIDASI CREATE
// ============================================================

func ValidateCreate(
	req model.CreateStudentRequest,
) error {

	errs := map[string]string{}

	if strings.TrimSpace(req.ID) == "" {
		errs["id"] = "wajib diisi"
	}

	if strings.TrimSpace(req.NIM) == "" {
		errs["nim"] = "wajib diisi"
	}

	if strings.TrimSpace(req.Name) == "" {
		errs["name"] = "wajib diisi"
	}

	if req.Grade < 0 || req.Grade > 100 {
		errs["grade"] = "harus berada di antara 0 dan 100"
	}

	if len(errs) > 0 {
		return &ValidationError{
			Fields: errs,
		}
	}

	return nil
}

// ============================================================
// VALIDASI PUT
// ============================================================

func ValidateReplace(
	req model.ReplaceStudentRequest,
) error {

	errs := map[string]string{}

	if strings.TrimSpace(req.NIM) == "" {
		errs["nim"] = "wajib diisi pada PUT"
	}

	if strings.TrimSpace(req.Name) == "" {
		errs["name"] = "wajib diisi pada PUT"
	}

	if req.Grade < 0 || req.Grade > 100 {
		errs["grade"] = "harus berada di antara 0 dan 100"
	}

	if len(errs) > 0 {
		return &ValidationError{
			Fields: errs,
		}
	}

	return nil
}
