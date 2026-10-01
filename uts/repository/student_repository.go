package repository

import (
	"siakad-mini/domain"

	"gorm.io/gorm"
)

type StudentRepository interface {
	FindAll() ([]domain.Student, error)
	FindByID(id uint) (*domain.Student, error)
	Create(student *domain.Student) error
	Update(student *domain.Student) error
}

type studentRepository struct {
	db *gorm.DB
}

func NewStudentRepository(db *gorm.DB) StudentRepository {
	return &studentRepository{db}
}

func (r *studentRepository) FindAll() ([]domain.Student, error) {
	var students []domain.Student
	err := r.db.Find(&students).Error
	if err != nil {
		return nil, err
	}
	return students, nil
}

func (r *studentRepository) FindByID(id uint) (*domain.Student, error) {
	var student domain.Student
	err := r.db.First(&student, id).Error
	if err != nil {
		return nil, err
	}
	return &student, nil
}

func (r *studentRepository) Create(student *domain.Student) error {
	return r.db.Create(student).Error
}

func (r *studentRepository) Update(student *domain.Student) error {
	// Updates akan memperbarui field-field yang berubah
	return r.db.Save(student).Error
}
