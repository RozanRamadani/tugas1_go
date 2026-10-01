package domain

import (
	"time"

	"gorm.io/gorm"
)

type Role string

const (
	RoleAdmin     Role = "admin"
	RoleMahasiswa Role = "mahasiswa"
)

// User model sesuai dengan PDF (id, email, password, role)
type User struct {
	ID        uint   `gorm:"primaryKey"`
	Email     string `gorm:"uniqueIndex;not null;size:100"` // Menggunakan email
	Password  string `gorm:"not null"`
	Role      Role   `gorm:"type:varchar(20);not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Student model sesuai PDF (id, user_id, nim, nama, prodi, angkatan, ipk_terakhir, deleted_at)
type Student struct {
	ID          uint   `gorm:"primaryKey"`
	UserID      uint   `gorm:"not null;uniqueIndex"` // Relasi 1-1 ke User
	NIM         string `gorm:"uniqueIndex;not null;size:20"`
	Nama        string `gorm:"not null;size:100"`
	Prodi       string `gorm:"size:100"`
	Angkatan    int
	IpkTerakhir float64
	User        User `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"` // Fitur soft-delete
}

// Course model sesuai PDF (id, kode_mk, nama_mk, sks, semester, kuota)
type Course struct {
	ID        uint   `gorm:"primaryKey"`
	KodeMk    string `gorm:"uniqueIndex;not null;size:20"`
	NamaMk    string `gorm:"not null;size:100"`
	Sks       int    `gorm:"not null"`
	Semester  int    `gorm:"not null"`
	Kuota     int    `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// CourseResponse adalah struktur khusus (DTO) untuk mengembalikan response JSON
// yang memuat kalkulasi sisa_kuota dan terisi dari tabel enrollments.
type CourseResponse struct {
	ID        uint   `json:"id"`
	KodeMk    string `json:"kode_mk"`
	NamaMk    string `json:"nama_mk"`
	Sks       int    `json:"sks"`
	Semester  int    `json:"semester"`
	Kuota     int    `json:"kuota"`
	Terisi    int    `json:"terisi"`
	SisaKuota int    `json:"sisa_kuota"`
}

// Enrollment model sesuai PDF (id, student_id, course_id, tahun_akademik, created_at)
type Enrollment struct {
	ID            uint    `gorm:"primaryKey"`
	StudentID     uint    `gorm:"uniqueIndex:idx_student_course_year;not null"`
	CourseID      uint    `gorm:"uniqueIndex:idx_student_course_year;not null"`
	TahunAkademik string  `gorm:"uniqueIndex:idx_student_course_year;not null;size:20"`
	Student       Student `gorm:"foreignKey:StudentID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Course        Course  `gorm:"foreignKey:CourseID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// EnrollmentRequest adalah struktur (DTO) khusus untuk menangkap request body pembuatan KRS
// Perhatikan: StudentID sengaja tidak disertakan di sini karena harus diambil dari JWT.
type EnrollmentRequest struct {
	CourseID      uint   `json:"course_id" validate:"required"`
	TahunAkademik string `json:"tahun_akademik" validate:"required"`
}

type CourseEnrolled struct {
	CourseID      uint   `json:"course_id"`
	KodeMk        string `json:"kode_mk"`
	NamaMk        string `json:"nama_mk"`
	Sks           int    `json:"sks"`
	Semester      int    `json:"semester"`
	TahunAkademik string `json:"tahun_akademik"`
}

type StudentDetailResponse struct {
	ID               uint             `json:"id"`
	NIM              string           `json:"nim"`
	Nama             string           `json:"nama"`
	Prodi            string           `json:"prodi"`
	Angkatan         int              `json:"angkatan"`
	IpkTerakhir      float64          `json:"ipk_terakhir"`
	DaftarMataKuliah []CourseEnrolled `json:"daftar_mata_kuliah"`
	TotalSks         int              `json:"total_sks"`
	BatasSks         int              `json:"batas_sks"`
}

// PaginationMeta adalah metadata umum untuk response dengan pagination.
type PaginationMeta struct {
	CurrentPage int   `json:"current_page"`
	PerPage     int   `json:"per_page"`
	TotalData   int64 `json:"total_data"`
	TotalPage   int   `json:"total_page"`
}
