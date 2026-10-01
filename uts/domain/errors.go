package domain

import "errors"

// Kumpulan Domain/Application Errors untuk fitur Enrollment (KRS)
// Error ini murni merupakan aturan bisnis aplikasi dan tidak bergantung pada library HTTP/Fiber.

var (
	ErrStudentNotFound     = errors.New("data mahasiswa tidak ditemukan")
	ErrCourseNotFound      = errors.New("mata kuliah tidak ditemukan")
	ErrEnrollmentDuplicate = errors.New("mahasiswa sudah mengambil mata kuliah ini pada tahun akademik yang sama")
	ErrCourseFull          = errors.New("kuota mata kuliah sudah penuh")
	ErrSKSLimitExceeded    = errors.New("total SKS melebihi batas maksimal berdasarkan IPK")
)
