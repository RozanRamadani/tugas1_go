package repository

import (
	"siakad-mini/domain"

	"gorm.io/gorm"
)

type StudentRepository interface {
	FindAll() ([]domain.Student, error)
}

type studentRepository struct {
	db *gorm.DB
}

func NewStudentRepository(db *gorm.DB) StudentRepository {
	return &studentRepository{db}
}

func (r *studentRepository) FindAll() ([]domain.Student, error) {
	var students []domain.Student
	// Mengambil semua data mahasiswa dari database (belum memuat relasi user agar password tidak bocor)
	err := r.db.Find(&students).Error
	if err != nil {
		return nil, err
	}
	return students, nil
}
