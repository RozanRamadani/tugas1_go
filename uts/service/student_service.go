package service

import (
	"siakad-mini/domain"
	"siakad-mini/repository"
)

type StudentService interface {
	GetAllStudents() ([]domain.Student, error)
}

type studentService struct {
	studentRepo repository.StudentRepository
}

func NewStudentService(studentRepo repository.StudentRepository) StudentService {
	return &studentService{studentRepo}
}

func (s *studentService) GetAllStudents() ([]domain.Student, error) {
	// Memanggil fungsi FindAll di repository
	return s.studentRepo.FindAll()
}
