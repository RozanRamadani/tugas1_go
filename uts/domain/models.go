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

// User model untuk menyimpan data kredensial login
type User struct {
	ID        uint      `gorm:"primaryKey"`
	Username  string    `gorm:"uniqueIndex;not null;size:100"`
	Password  string    `gorm:"not null"`
	Role      Role      `gorm:"type:varchar(20);not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Student model untuk menyimpan profil mahasiswa
type Student struct {
	ID        uint           `gorm:"primaryKey"`
	NIM       string         `gorm:"uniqueIndex;not null;size:20"`
	Name      string         `gorm:"not null;size:100"`
	UserID    uint           `gorm:"not null;uniqueIndex"` // Relasi 1-to-1 ke User (opsional, disesuaikan kebutuhan)
	User      User           `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"` // Fitur soft-delete
}

// Course model untuk menyimpan mata kuliah
type Course struct {
	ID        uint      `gorm:"primaryKey"`
	Code      string    `gorm:"uniqueIndex;not null;size:20"`
	Name      string    `gorm:"not null;size:100"`
	Credits   int       `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Enrollment model untuk menyimpan pendaftaran KRS (relasi m-to-n antara Student dan Course)
type Enrollment struct {
	ID            uint      `gorm:"primaryKey"`
	StudentID     uint      `gorm:"uniqueIndex:idx_student_course_year;not null"`
	CourseID      uint      `gorm:"uniqueIndex:idx_student_course_year;not null"`
	TahunAkademik string    `gorm:"uniqueIndex:idx_student_course_year;not null;size:20"`
	Grade         *string   `gorm:"size:2"` // Nullable field, karena saat daftar KRS belum ada nilai
	Student       Student   `gorm:"foreignKey:StudentID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Course        Course    `gorm:"foreignKey:CourseID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
