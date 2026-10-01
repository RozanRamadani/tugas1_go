package repository

import (
	"siakad-mini/domain"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type StudentRepository interface {
	FindAllWithFilters(prodi, angkatan, search, sort string, page, perPage int) ([]domain.Student, int64, error)
	FindByID(id uint) (*domain.Student, error)
	FindByUserID(userID uint) (*domain.Student, error)
	FindByUserIDForUpdate(userID uint) (*domain.Student, error)
	Create(student *domain.Student) error
	Update(student *domain.Student) error
	Delete(student *domain.Student) error
	GetEnrolledCourses(studentID uint, tahunAkademik string) ([]domain.Enrollment, error)
}

type studentRepository struct {
	db *gorm.DB
}

func NewStudentRepository(db *gorm.DB) StudentRepository {
	return &studentRepository{db}
}

func (r *studentRepository) FindAllWithFilters(prodi, angkatan, search, sort string, page, perPage int) ([]domain.Student, int64, error) {
	var students []domain.Student
	var totalData int64

	query := r.db.Model(&domain.Student{})

	if prodi != "" {
		query = query.Where("prodi = ?", prodi)
	}

	if angkatan != "" {
		query = query.Where("angkatan = ?", angkatan)
	}

	if search != "" {
		query = query.Where("nama ILIKE ? OR nim ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// Hitung total data sebelum limit & offset
	err := query.Count(&totalData).Error
	if err != nil {
		return nil, 0, err
	}

	// Sorting
	if sort != "" {
		if sort == "-ipk_terakhir" {
			query = query.Order("ipk_terakhir DESC")
		} else if sort == "ipk_terakhir" {
			query = query.Order("ipk_terakhir ASC")
		} else if sort == "-nama" {
			query = query.Order("nama DESC")
		} else if sort == "nama" {
			query = query.Order("nama ASC")
		}
	} else {
		// Default sort by created_at desc
		query = query.Order("created_at DESC")
	}

	// Pagination
	offset := (page - 1) * perPage
	query = query.Limit(perPage).Offset(offset)

	err = query.Find(&students).Error
	if err != nil {
		return nil, 0, err
	}
	return students, totalData, nil
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

func (r *studentRepository) GetEnrolledCourses(studentID uint, tahunAkademik string) ([]domain.Enrollment, error) {
	var enrollments []domain.Enrollment
	query := r.db.Preload("Course").Where("student_id = ?", studentID)

	if tahunAkademik != "" {
		query = query.Where("tahun_akademik = ?", tahunAkademik)
	}

	err := query.Find(&enrollments).Error
	return enrollments, err
}
