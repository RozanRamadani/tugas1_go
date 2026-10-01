package handler

import (
	"fmt"
	"regexp"

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

	// Otorisasi: Mahasiswa hanya boleh melihat profil sendiri
	roleStr, okRole := c.Locals("role").(string)
	jwtUserID, okUser := c.Locals("user_id").(float64)
	if !okRole || !okUser {
		return helper.ErrorResponse(c, fiber.StatusUnauthorized, "Token tidak valid")
	}

	// Cek apakah data mahasiswa ada
	student, err := h.studentService.GetStudentByID(uint(id))
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusNotFound, "Data mahasiswa tidak ditemukan")
	}

	// Jika role mahasiswa, userID JWT harus sama dengan userID milik Student
	if roleStr == string(domain.RoleMahasiswa) && uint(jwtUserID) != student.UserID {
		return helper.ErrorResponse(c, fiber.StatusForbidden, "Akses ditolak: Anda tidak memiliki izin untuk melihat profil mahasiswa lain")
	}

	// Filter tahun akademik (opsional)
	tahunAkademik := c.Query("tahun_akademik")
	if tahunAkademik != "" {
		matched, _ := regexp.MatchString(`^\d{4}/\d{4}-(Ganjil|Genap)$`, tahunAkademik)
		if !matched {
			return helper.ValidationErrorResponse(c, []string{"Field 'tahun_akademik' harus berformat YYYY/YYYY-Ganjil atau YYYY/YYYY-Genap"})
		}
	}

	// Ambil detail (lengkap dengan mata kuliah, total sks, batas sks)
	detail, err := h.studentService.GetStudentDetail(uint(id), tahunAkademik)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal merelasikan data mahasiswa")
	}

	return helper.SuccessResponse(c, fiber.StatusOK, "Berhasil mengambil data mahasiswa", detail)
}

type UpdateStudentRequest struct {
	// NIM tidak dimasukkan agar tidak dapat diubah (meskipun dikirim oleh user)
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
		Nama:        req.Nama,
		Prodi:       req.Prodi,
		Angkatan:    req.Angkatan,
		IpkTerakhir: req.IpkTerakhir,
	}

	updatedStudent, err := h.studentService.UpdateStudent(uint(id), &updatedData)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusNotFound, "Gagal mengupdate: Data mahasiswa tidak ditemukan")
	}

	return helper.SuccessResponse(c, fiber.StatusOK, "Berhasil mengupdate data mahasiswa", updatedStudent)
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
