package repository

import (
	"siakad-mini/domain"

	"gorm.io/gorm"
)

type EnrollmentRepository interface {
	Exists(studentID uint, courseID uint, tahunAkademik string) (bool, error)
	GetTotalSKS(studentID uint, tahunAkademik string) (int, error)
	Create(enrollment *domain.Enrollment) error
	FindByID(id uint) (*domain.Enrollment, error)
	Delete(id uint) error
}

type enrollmentRepository struct {
	db *gorm.DB
}

func NewEnrollmentRepository(db *gorm.DB) EnrollmentRepository {
	return &enrollmentRepository{db}
}

func (r *enrollmentRepository) Exists(studentID uint, courseID uint, tahunAkademik string) (bool, error) {
	var count int64
	err := r.db.Model(&domain.Enrollment{}).
		Where("student_id = ? AND course_id = ? AND tahun_akademik = ?", studentID, courseID, tahunAkademik).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *enrollmentRepository) GetTotalSKS(studentID uint, tahunAkademik string) (int, error) {
	var totalSKS int
	// Bergabung dengan tabel courses untuk mengambil dan menjumlahkan SKS
	err := r.db.Table("enrollments").
		Select("COALESCE(SUM(courses.sks), 0)").
		Joins("JOIN courses ON enrollments.course_id = courses.id").
		Where("enrollments.student_id = ? AND enrollments.tahun_akademik = ?", studentID, tahunAkademik).
		Scan(&totalSKS).Error
	if err != nil {
		return 0, err
	}
	return totalSKS, nil
}

func (r *enrollmentRepository) Create(enrollment *domain.Enrollment) error {
	return r.db.Create(enrollment).Error
}

func (r *enrollmentRepository) FindByID(id uint) (*domain.Enrollment, error) {
	var enrollment domain.Enrollment
	err := r.db.First(&enrollment, id).Error
	if err != nil {
		return nil, err
	}
	return &enrollment, nil
}

func (r *enrollmentRepository) Delete(id uint) error {
	return r.db.Delete(&domain.Enrollment{}, id).Error
}
