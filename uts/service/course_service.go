package service

import (
	"siakad-mini/domain"
	"siakad-mini/repository"
)

type CourseService interface {
	GetCourses(semester, search, available string, page, perPage int) ([]domain.CourseResponse, int64, error)
}

type courseService struct {
	courseRepo repository.CourseRepository
}

func NewCourseService(courseRepo repository.CourseRepository) CourseService {
	return &courseService{courseRepo}
}

func (s *courseService) GetCourses(semester, search, available string, page, perPage int) ([]domain.CourseResponse, int64, error) {
	return s.courseRepo.FindAllWithFilters(semester, search, available, page, perPage)
}
