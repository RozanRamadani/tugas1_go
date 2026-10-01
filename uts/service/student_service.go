package service

import (
	"siakad-mini/domain"
	"siakad-mini/repository"

	"golang.org/x/crypto/bcrypt"
)

type StudentService interface {
	GetAllStudents() ([]domain.Student, error)
	GetStudentByID(id uint) (*domain.Student, error)
	CreateStudent(student *domain.Student, rawPassword string) error
}

type studentService struct {
	studentRepo repository.StudentRepository
}

func NewStudentService(studentRepo repository.StudentRepository) StudentService {
	return &studentService{studentRepo}
}

func (s *studentService) GetAllStudents() ([]domain.Student, error) {
	return s.studentRepo.FindAll()
}

func (s *studentService) GetStudentByID(id uint) (*domain.Student, error) {
	return s.studentRepo.FindByID(id)
}

func (s *studentService) CreateStudent(student *domain.Student, rawPassword string) error {
	bytes, err := bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	student.User.Password = string(bytes)
	student.User.Role = domain.RoleMahasiswa

	return s.studentRepo.Create(student)
}
