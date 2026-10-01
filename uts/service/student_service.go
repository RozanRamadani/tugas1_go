package service

import (
	"siakad-mini/domain"
	"siakad-mini/repository"

	"golang.org/x/crypto/bcrypt"
)

type StudentService interface {
	GetAllStudents() ([]domain.Student, error)
	GetStudentByID(id uint) (*domain.Student, error)
	GetStudentDetail(id uint, tahunAkademik string) (*domain.StudentDetailResponse, error)
	CreateStudent(student *domain.Student, rawPassword string) error
	UpdateStudent(id uint, updatedData *domain.Student) (*domain.Student, error)
	DeleteStudent(id uint) error
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

func (s *studentService) GetStudentDetail(id uint, tahunAkademik string) (*domain.StudentDetailResponse, error) {
	student, err := s.studentRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	enrollments, err := s.studentRepo.GetEnrolledCourses(id, tahunAkademik)
	if err != nil {
		return nil, err
	}

	var daftarMk []domain.CourseEnrolled
	totalSks := 0
	for _, e := range enrollments {
		daftarMk = append(daftarMk, domain.CourseEnrolled{
			CourseID:      e.CourseID,
			KodeMk:        e.Course.KodeMk,
			NamaMk:        e.Course.NamaMk,
			Sks:           e.Course.Sks,
			Semester:      e.Course.Semester,
			TahunAkademik: e.TahunAkademik,
		})
		totalSks += e.Course.Sks
	}

	// Memastikan slice tidak bernilai nil saat diserialisasi (akan menjadi "[]" daripada null)
	if daftarMk == nil {
		daftarMk = []domain.CourseEnrolled{}
	}

	batasSks := 18
	if student.IpkTerakhir >= 3.00 {
		batasSks = 24
	} else if student.IpkTerakhir >= 2.50 {
		batasSks = 21
	}

	resp := &domain.StudentDetailResponse{
		ID:               student.ID,
		NIM:              student.NIM,
		Nama:             student.Nama,
		Prodi:            student.Prodi,
		Angkatan:         student.Angkatan,
		IpkTerakhir:      student.IpkTerakhir,
		DaftarMataKuliah: daftarMk,
		TotalSks:         totalSks,
		BatasSks:         batasSks,
	}

	return resp, nil
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

func (s *studentService) UpdateStudent(id uint, updatedData *domain.Student) (*domain.Student, error) {
	student, err := s.studentRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	student.Nama = updatedData.Nama
	student.Prodi = updatedData.Prodi
	student.Angkatan = updatedData.Angkatan
	student.IpkTerakhir = updatedData.IpkTerakhir

	err = s.studentRepo.Update(student)
	return student, err
}

func (s *studentService) DeleteStudent(id uint) error {
	// 1. Cari data mahasiswa yang lama
	student, err := s.studentRepo.FindByID(id)
	if err != nil {
		return err // Mengembalikan error jika tidak ditemukan
	}

	// 2. Hapus data (Soft Delete)
	return s.studentRepo.Delete(student)
}
