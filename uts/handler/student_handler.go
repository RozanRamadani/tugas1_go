package handler

import (
	"fmt"

	"siakad-mini/domain"
	"siakad-mini/helper"
	"siakad-mini/service"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type StudentHandler struct {
	studentService service.StudentService
	validator      *validator.Validate
}

func NewStudentHandler(studentService service.StudentService) *StudentHandler {
	return &StudentHandler{
		studentService: studentService,
		validator:      validator.New(),
	}
}

// GetAll menangani rute GET /api/v1/students
func (h *StudentHandler) GetAll(c *fiber.Ctx) error {
	students, err := h.studentService.GetAllStudents()
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data mahasiswa")
	}

	return helper.SuccessResponse(c, fiber.StatusOK, "Berhasil mengambil data mahasiswa", students)
}

type CreateStudentRequest struct {
	Email       string  `json:"email" validate:"required,email"`
	Password    string  `json:"password" validate:"required,min=8"`
	NIM         string  `json:"nim" validate:"required"`
	Nama        string  `json:"nama" validate:"required"`
	Prodi       string  `json:"prodi" validate:"required"`
	Angkatan    int     `json:"angkatan" validate:"required"`
	IpkTerakhir float64 `json:"ipk_terakhir"`
}

// Create menangani rute POST /api/v1/students
func (h *StudentHandler) Create(c *fiber.Ctx) error {
	var req CreateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "Format request tidak valid")
	}

	if err := h.validator.Struct(req); err != nil {
		var errorMessages []string
		for _, err := range err.(validator.ValidationErrors) {
			errorMessages = append(errorMessages, fmt.Sprintf("Field '%s' tidak valid berdasarkan aturan '%s'", err.Field(), err.Tag()))
		}
		return helper.ValidationErrorResponse(c, errorMessages)
	}

	// Mapping Request DTO ke Domain Model
	student := domain.Student{
		NIM:         req.NIM,
		Nama:        req.Nama,
		Prodi:       req.Prodi,
		Angkatan:    req.Angkatan,
		IpkTerakhir: req.IpkTerakhir,
		User: domain.User{
			Email: req.Email,
			// Password diset di Service
		},
	}

	err := h.studentService.CreateStudent(&student, req.Password)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal membuat data mahasiswa, periksa apakah email/NIM sudah terdaftar")
	}

	return helper.SuccessResponse(c, fiber.StatusCreated, "Berhasil membuat data mahasiswa", nil)
}

// GetByID menangani rute GET /api/v1/students/:id
func (h *StudentHandler) GetByID(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "ID tidak valid")
	}

	student, err := h.studentService.GetStudentByID(uint(id))
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusNotFound, "Data mahasiswa tidak ditemukan")
	}

	return helper.SuccessResponse(c, fiber.StatusOK, "Berhasil mengambil data mahasiswa", student)
}

type UpdateStudentRequest struct {
	NIM         string  `json:"nim" validate:"required"`
	Nama        string  `json:"nama" validate:"required"`
	Prodi       string  `json:"prodi" validate:"required"`
	Angkatan    int     `json:"angkatan" validate:"required"`
	IpkTerakhir float64 `json:"ipk_terakhir"`
}

// Update menangani rute PUT /api/v1/students/:id
func (h *StudentHandler) Update(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "ID tidak valid")
	}

	var req UpdateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "Format request tidak valid")
	}

	if err := h.validator.Struct(req); err != nil {
		var errorMessages []string
		for _, err := range err.(validator.ValidationErrors) {
			errorMessages = append(errorMessages, fmt.Sprintf("Field '%s' tidak valid berdasarkan aturan '%s'", err.Field(), err.Tag()))
		}
		return helper.ValidationErrorResponse(c, errorMessages)
	}

	updatedData := domain.Student{
		NIM:         req.NIM,
		Nama:        req.Nama,
		Prodi:       req.Prodi,
		Angkatan:    req.Angkatan,
		IpkTerakhir: req.IpkTerakhir,
	}

	err = h.studentService.UpdateStudent(uint(id), &updatedData)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusNotFound, "Gagal mengupdate: Data mahasiswa tidak ditemukan")
	}

	return helper.SuccessResponse(c, fiber.StatusOK, "Berhasil mengupdate data mahasiswa", nil)
}

// Delete menangani rute DELETE /api/v1/students/:id
func (h *StudentHandler) Delete(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "ID tidak valid")
	}

	err = h.studentService.DeleteStudent(uint(id))
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusNotFound, "Gagal menghapus: Data mahasiswa tidak ditemukan")
	}

	return helper.SuccessResponse(c, fiber.StatusOK, "Berhasil menghapus data mahasiswa (Soft Delete)", nil)
}
