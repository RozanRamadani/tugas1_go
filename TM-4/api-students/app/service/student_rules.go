package service

import (
	"strings"

	"api-students/app/model"
)

// ValidateCreate memvalidasi data student untuk POST.
func ValidateCreate(req model.CreateStudentRequest) map[string]string {
	errs := map[string]string{}

	req.ID = strings.TrimSpace(req.ID)
	req.NIM = strings.TrimSpace(req.NIM)
	req.Name = strings.TrimSpace(req.Name)

	if req.ID == "" {
		errs["id"] = "wajib diisi"
	}

	if req.NIM == "" {
		errs["nim"] = "wajib diisi"
	}

	if req.Name == "" {
		errs["name"] = "wajib diisi"
	}

	if req.Grade < 0 || req.Grade > 100 {
		errs["grade"] = "harus berada di antara 0 dan 100"
	}

	return errs
}

// ValidateReplace memvalidasi data student untuk PUT.
func ValidateReplace(req model.ReplaceStudentRequest) map[string]string {
	errs := map[string]string{}

	req.NIM = strings.TrimSpace(req.NIM)
	req.Name = strings.TrimSpace(req.Name)

	if req.NIM == "" {
		errs["nim"] = "wajib diisi pada PUT"
	}

	if req.Name == "" {
		errs["name"] = "wajib diisi pada PUT"
	}

	if req.Grade < 0 || req.Grade > 100 {
		errs["grade"] = "harus berada di antara 0 dan 100"
	}

	return errs
}

// ApplyPatch menerapkan perubahan PATCH pada data student.
func ApplyPatch(
	current model.Student,
	req model.PatchStudentRequest,
) (model.Student, map[string]string) {

	errs := map[string]string{}

	if req.NIM != nil {
		nim := strings.TrimSpace(*req.NIM)

		if nim == "" {
			errs["nim"] = "tidak boleh kosong"
		} else {
			current.NIM = nim
		}
	}

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)

		if name == "" {
			errs["name"] = "tidak boleh kosong"
		} else {
			current.Name = name
		}
	}

	if req.Grade != nil {
		if *req.Grade < 0 || *req.Grade > 100 {
			errs["grade"] = "harus berada di antara 0 dan 100"
		} else {
			current.Grade = *req.Grade
		}
	}

	if req.IsActive != nil {
		current.IsActive = *req.IsActive
	}

	return current, errs
}

// ValidationError menyimpan kumpulan error validasi.
type ValidationError struct {
	Errors map[string]string
}

func (e ValidationError) Error() string {
	return "validation error"
}