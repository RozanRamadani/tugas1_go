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
	UpdateStudent(id uint, updatedData *domain.Student) error
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

func (s *studentService) UpdateStudent(id uint, updatedData *domain.Student) error {
	// 1. Cari data mahasiswa yang lama
	student, err := s.studentRepo.FindByID(id)
	if err != nil {
		return err // Mengembalikan error jika tidak ditemukan
	}

	// 2. Timpa data lama dengan data baru
	student.NIM = updatedData.NIM
	student.Nama = updatedData.Nama
	student.Prodi = updatedData.Prodi
	student.Angkatan = updatedData.Angkatan
	student.IpkTerakhir = updatedData.IpkTerakhir

	// 3. Simpan perubahan ke database
	return s.studentRepo.Update(student)
}
