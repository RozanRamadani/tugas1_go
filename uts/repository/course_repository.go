package repository

import (
	"siakad-mini/domain"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CourseRepository interface {
	FindAllWithFilters(semester, search, available string, page, perPage int) ([]domain.CourseResponse, int64, error)
	FindByIDForUpdate(courseID uint) (*domain.Course, error)
	CountEnrollments(courseID uint, tahunAkademik string) (int64, error)
}

type courseRepository struct {
	db *gorm.DB
}

func NewCourseRepository(db *gorm.DB) CourseRepository {
	return &courseRepository{db}
}

func (r *courseRepository) FindAllWithFilters(semester, search, available string, page, perPage int) ([]domain.CourseResponse, int64, error) {
	var results []domain.CourseResponse
	var totalData int64

	query := r.db.Table("courses c").
		Select("c.id, c.kode_mk, c.nama_mk, c.sks, c.semester, c.kuota, COUNT(e.id) as terisi, (c.kuota - COUNT(e.id)) as sisa_kuota").
		Joins("LEFT JOIN enrollments e ON e.course_id = c.id").
		Group("c.id, c.kode_mk, c.nama_mk, c.sks, c.semester, c.kuota")

	if semester != "" {
		query = query.Where("c.semester = ?", semester)
	}

	if search != "" {
		query = query.Where("c.nama_mk ILIKE ?", "%"+search+"%")
	}

	if available == "true" {
		query = query.Having("(c.kuota - COUNT(e.id)) > 0")
	} else if available == "false" {
		query = query.Having("(c.kuota - COUNT(e.id)) <= 0")
	}

	// Count total data using a subquery to handle GROUP BY and HAVING properly
	err := r.db.Table("(?) as subquery", query).Count(&totalData).Error
	if err != nil {
		return nil, 0, err
	}

	// Pagination
	offset := (page - 1) * perPage
	query = query.Limit(perPage).Offset(offset)
	query = query.Order("c.id ASC") // Sorting default

	err = query.Scan(&results).Error
	if err != nil {
		return nil, 0, err
	}

	return results, totalData, nil
}

func (r *courseRepository) FindByIDForUpdate(courseID uint) (*domain.Course, error) {
	var course domain.Course
	// Menggunakan row locking (FOR UPDATE)
	err := r.db.Clauses(clause.Locking{Strength: "UPDATE"}).First(&course, courseID).Error
	if err != nil {
		return nil, err
	}
	return &course, nil
}

func (r *courseRepository) CountEnrollments(courseID uint, tahunAkademik string) (int64, error) {
	var count int64
	// Filter berdasarkan course_id, bukan dibatasi tahun akademik jika kuota berlaku total?
	// Tunggu, PDF UTS biasanya kuota berlaku per tahun akademik atau total?
	// Untuk keamanan KRS, biasanya kuota dihitung per mata kuliah tanpa peduli tahun akademik
	// Tapi instruksi meminta menghitung enrollment untuk course tersebut pada tahun akademik tertentu.
	err := r.db.Model(&domain.Enrollment{}).
		Where("course_id = ? AND tahun_akademik = ?", courseID, tahunAkademik).
		Count(&count).Error
	return count, err
}
