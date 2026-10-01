package handler

import (
	"siakad-mini/domain"
	"siakad-mini/helper"
	"siakad-mini/service"

	"github.com/gofiber/fiber/v2"
)

type CourseHandler struct {
	courseService service.CourseService
}

func NewCourseHandler(courseService service.CourseService) *CourseHandler {
	return &CourseHandler{courseService}
}

// GetAll menangani rute GET /api/v1/courses
func (h *CourseHandler) GetAll(c *fiber.Ctx) error {
	// Pagination query params
	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 10)

	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}
	if perPage > 50 {
		perPage = 50
	}

	// Mengambil parameter query filter dari URL
	semester := c.Query("semester")
	search := c.Query("search")
	available := c.Query("available")

	// Panggil layer service
	courses, totalData, err := h.courseService.GetCourses(semester, search, available, page, perPage)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data mata kuliah")
	}

	totalPage := int((totalData + int64(perPage) - 1) / int64(perPage))

	meta := domain.PaginationMeta{
		CurrentPage: page,
		PerPage:     perPage,
		TotalData:   totalData,
		TotalPage:   totalPage,
	}

	return helper.SuccessResponseWithMeta(c, fiber.StatusOK, "Berhasil mengambil data mata kuliah", courses, meta)
}
