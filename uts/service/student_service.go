package service

import (
	"siakad-mini/domain"
	"siakad-mini/repository"

	"golang.org/x/crypto/bcrypt"
)

type StudentService interface {
	GetAllStudents() ([]domain.Student, error)
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

func (s *studentService) CreateStudent(student *domain.Student, rawPassword string) error {
	// 1. Hash password yang diberikan
	bytes, err := bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// 2. Set field User yang akan di-insert berbarengan dengan Student
	student.User.Password = string(bytes)
	student.User.Role = domain.RoleMahasiswa // Pastikan selalu menjadi role mahasiswa

	// 3. Panggil repository untuk menyimpan
	return s.studentRepo.Create(student)
}
