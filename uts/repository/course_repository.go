package repository

import (
	"siakad-mini/domain"

	"gorm.io/gorm"
)

type CourseRepository interface {
	FindAllWithFilters(semester, search, available string) ([]domain.CourseResponse, error)
}

type courseRepository struct {
	db *gorm.DB
}

func NewCourseRepository(db *gorm.DB) CourseRepository {
	return &courseRepository{db}
}

func (r *courseRepository) FindAllWithFilters(semester, search, available string) ([]domain.CourseResponse, error) {
	var results []domain.CourseResponse

	// Membuat query dasar dengan JOIN ke tabel enrollments
	query := r.db.Table("courses c").
		Select("c.id, c.kode_mk, c.nama_mk, c.sks, c.semester, c.kuota, COUNT(e.id) as terisi, (c.kuota - COUNT(e.id)) as sisa_kuota").
		Joins("LEFT JOIN enrollments e ON e.course_id = c.id").
		Group("c.id, c.kode_mk, c.nama_mk, c.sks, c.semester, c.kuota")

	// Filter semester
	if semester != "" {
		query = query.Where("c.semester = ?", semester)
	}

	// Filter pencarian nama mata kuliah (case-insensitive menggunakan ILIKE)
	if search != "" {
		query = query.Where("c.nama_mk ILIKE ?", "%"+search+"%")
	}

	// Filter ketersediaan (available=true berarti sisa_kuota > 0)
	if available == "true" {
		query = query.Having("(c.kuota - COUNT(e.id)) > 0")
	} else if available == "false" {
		query = query.Having("(c.kuota - COUNT(e.id)) <= 0")
	}

	// Jalankan query dan scan hasilnya ke dalam slice of CourseResponse
	err := query.Scan(&results).Error
	if err != nil {
		return nil, err
	}

	return results, nil
}
