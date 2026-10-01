package service

import (
	"siakad-mini/domain"
	"siakad-mini/repository"

	"gorm.io/gorm"
)

type EnrollmentService interface {
	Enroll(userID uint, req domain.EnrollmentRequest) error
}

type enrollmentService struct {
	db *gorm.DB
}

func NewEnrollmentService(db *gorm.DB) EnrollmentService {
	return &enrollmentService{db: db}
}

func determineMaxSKS(ipk float64) int {
	if ipk >= 3.00 {
		return 24
	}
	if ipk >= 2.50 {
		return 21
	}
	return 18
}

func (s *enrollmentService) Enroll(userID uint, req domain.EnrollmentRequest) error {
	tx := s.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// Defer untuk memastikan rollback jika terjadi panic
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	studentRepo := repository.NewStudentRepository(tx)
	courseRepo := repository.NewCourseRepository(tx)
	enrollmentRepo := repository.NewEnrollmentRepository(tx)

	// 1. Lock Student FOR UPDATE
	student, err := studentRepo.FindByUserIDForUpdate(userID)
	if err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return domain.ErrStudentNotFound
		}
		return err
	}

	// 2. Lock Course FOR UPDATE
	course, err := courseRepo.FindByIDForUpdate(req.CourseID)
	if err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return domain.ErrCourseNotFound
		}
		return err
	}

	// 3. Check duplicate enrollment
	exists, err := enrollmentRepo.Exists(student.ID, req.CourseID, req.TahunAkademik)
	if err != nil {
		tx.Rollback()
		return err
	}
	if exists {
		tx.Rollback()
		return domain.ErrEnrollmentDuplicate
	}

	// 4. Check quota
	terisi, err := courseRepo.CountEnrollments(req.CourseID, req.TahunAkademik)
	if err != nil {
		tx.Rollback()
		return err
	}
	if int(terisi) >= course.Kuota {
		tx.Rollback()
		return domain.ErrCourseFull
	}

	// 5. Calculate current SKS
	currentSKS, err := enrollmentRepo.GetTotalSKS(student.ID, req.TahunAkademik)
	if err != nil {
		tx.Rollback()
		return err
	}

	// 6. Tentukan batas SKS berdasarkan IPK & validasi
	maxSKS := determineMaxSKS(student.IpkTerakhir)
	newTotalSKS := currentSKS + course.Sks
	if newTotalSKS > maxSKS {
		tx.Rollback()
		return domain.ErrSKSLimitExceeded
	}

	// 7. Create enrollment
	enrollment := domain.Enrollment{
		StudentID:     student.ID,
		CourseID:      course.ID,
		TahunAkademik: req.TahunAkademik,
	}

	if err := enrollmentRepo.Create(&enrollment); err != nil {
		tx.Rollback()
		return err
	}

	// 8. Commit
	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}
