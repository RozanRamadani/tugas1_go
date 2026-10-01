package handler

import (
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
	// Mengambil parameter query dari URL (misal: ?semester=3&search=Basis&available=true)
	semester := c.Query("semester")
	search := c.Query("search")
	available := c.Query("available")

	// Panggil layer service
	courses, err := h.courseService.GetCourses(semester, search, available)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data mata kuliah")
	}

	return helper.SuccessResponse(c, fiber.StatusOK, "Berhasil mengambil data mata kuliah", courses)
}
