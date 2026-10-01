package domain

import (
	"errors"
	"fmt"
)

// Kumpulan Domain/Application Errors untuk fitur Enrollment (KRS)
// Error ini murni merupakan aturan bisnis aplikasi dan tidak bergantung pada library HTTP/Fiber.

var (
	ErrStudentNotFound     = errors.New("data mahasiswa tidak ditemukan")
	ErrCourseNotFound      = errors.New("mata kuliah tidak ditemukan")
	ErrEnrollmentDuplicate = errors.New("mahasiswa sudah mengambil mata kuliah ini pada tahun akademik yang sama")
	ErrCourseFull          = errors.New("kuota mata kuliah sudah penuh")
	ErrSKSLimitExceeded    = errors.New("total SKS melebihi batas maksimal berdasarkan IPK")
	ErrEnrollmentNotFound  = errors.New("data KRS tidak ditemukan")
	ErrEnrollmentForbidden = errors.New("anda tidak memiliki akses untuk membatalkan KRS ini")
)

type SKSLimitError struct {
	RemainingSKS int
}

func (e *SKSLimitError) Error() string {
	return fmt.Sprintf("total SKS melebihi batas maksimal, sisa SKS Anda: %d SKS", e.RemainingSKS)
}

func (e *SKSLimitError) Unwrap() error {
	return ErrSKSLimitExceeded
}
