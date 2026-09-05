package service

import (
	"testing"

	"api-students/app/model"
)

func TestValidateCreate(t *testing.T) {
	req := model.CreateStudentRequest{
		ID:       "S001",
		NIM:      "23123456",
		Name:     "Rozan",
		Grade:    90,
		IsActive: true,
	}

	errs := ValidateCreate(req)

	if len(errs) != 0 {
		t.Fatalf("seharusnya tidak ada error: %v", errs)
	}
}

func TestValidateReplaceInvalidGrade(t *testing.T) {
	req := model.ReplaceStudentRequest{
		NIM:      "23123456",
		Name:     "Rozan",
		Grade:    120,
		IsActive: true,
	}

	errs := ValidateReplace(req)

	if errs["grade"] == "" {
		t.Fatal("grade 120 seharusnya menghasilkan error")
	}
}

func TestApplyPatch(t *testing.T) {
	current := model.Student{
		ID:       "S001",
		NIM:      "23123456",
		Name:     "Rozan",
		Grade:    90,
		IsActive: true,
	}

	newName := "Rozan Baru"
	newGrade := 95.5

	req := model.PatchStudentRequest{
		Name:  &newName,
		Grade: &newGrade,
	}

	result, errs := ApplyPatch(current, req)

	if len(errs) != 0 {
		t.Fatalf("seharusnya tidak ada error: %v", errs)
	}

	if result.Name != "Rozan Baru" {
		t.Fatalf("name tidak berubah: %s", result.Name)
	}

	if result.Grade != 95.5 {
		t.Fatalf("grade tidak berubah: %.2f", result.Grade)
	}

	if result.NIM != "23123456" {
		t.Fatalf("NIM tidak boleh berubah: %s", result.NIM)
	}
}