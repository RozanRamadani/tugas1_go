package handler

import (
	"siakad-mini/helper"
	"siakad-mini/service"

	"github.com/gofiber/fiber/v2"
)

type StudentHandler struct {
	studentService service.StudentService
}

func NewStudentHandler(studentService service.StudentService) *StudentHandler {
	return &StudentHandler{studentService}
}

// GetAll menangani rute GET /api/v1/students
func (h *StudentHandler) GetAll(c *fiber.Ctx) error {
	students, err := h.studentService.GetAllStudents()
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data mahasiswa")
	}

	return helper.SuccessResponse(c, fiber.StatusOK, "Berhasil mengambil data mahasiswa", students)
}
