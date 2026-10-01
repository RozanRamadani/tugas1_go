package repository

import (
	"siakad-mini/domain"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type StudentRepository interface {
	FindAll() ([]domain.Student, error)
	FindByID(id uint) (*domain.Student, error)
	FindByUserID(userID uint) (*domain.Student, error)
	FindByUserIDForUpdate(userID uint) (*domain.Student, error)
	Create(student *domain.Student) error
	Update(student *domain.Student) error
	Delete(student *domain.Student) error
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

func (r *studentRepository) FindByUserID(userID uint) (*domain.Student, error) {
	var student domain.Student
	err := r.db.Where("user_id = ?", userID).First(&student).Error
	if err != nil {
		return nil, err
	}
	return &student, nil
}

func (r *studentRepository) FindByUserIDForUpdate(userID uint) (*domain.Student, error) {
	var student domain.Student
	err := r.db.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ?", userID).First(&student).Error
	if err != nil {
		return nil, err
	}
	return &student, nil
}

func (r *studentRepository) Create(student *domain.Student) error {
	return r.db.Create(student).Error
}

func (r *studentRepository) Update(student *domain.Student) error {
	return r.db.Save(student).Error
}

func (r *studentRepository) Delete(student *domain.Student) error {
	return r.db.Delete(student).Error
}
