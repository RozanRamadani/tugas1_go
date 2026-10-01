package handler

import (
	"errors"
	"fmt"
	"regexp"

	"siakad-mini/domain"
	"siakad-mini/helper"
	"siakad-mini/service"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type EnrollmentHandler struct {
	enrollmentService service.EnrollmentService
	validator         *validator.Validate
}

func NewEnrollmentHandler(enrollmentService service.EnrollmentService) *EnrollmentHandler {
	return &EnrollmentHandler{
		enrollmentService: enrollmentService,
		validator:         validator.New(),
	}
}

// Create menangani rute POST /api/v1/enrollments
func (h *EnrollmentHandler) Create(c *fiber.Ctx) error {
	var req domain.EnrollmentRequest
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

	// Validasi Format Tahun Akademik: YYYY/YYYY-Ganjil atau YYYY/YYYY-Genap
	matched, _ := regexp.MatchString(`^\d{4}/\d{4}-(Ganjil|Genap)$`, req.TahunAkademik)
	if !matched {
		return helper.ValidationErrorResponse(c, []string{"Field 'tahun_akademik' harus berformat YYYY/YYYY-Ganjil atau YYYY/YYYY-Genap"})
	}

	// Ambil user_id dari JWT Locals (secara default jwt parse angka menjadi float64)
	userIDFloat, ok := c.Locals("user_id").(float64)
	if !ok {
		return helper.ErrorResponse(c, fiber.StatusUnauthorized, "Token JWT tidak valid atau user_id tidak ditemukan")
	}
	userID := uint(userIDFloat)

	// Panggil Service
	err := h.enrollmentService.Enroll(userID, req)
	if err != nil {
		// Error Mapping berdasarkan Domain Errors
		if errors.Is(err, domain.ErrStudentNotFound) || errors.Is(err, domain.ErrCourseNotFound) {
			return helper.ErrorResponse(c, fiber.StatusNotFound, err.Error())
		}
		if errors.Is(err, domain.ErrEnrollmentDuplicate) {
			return helper.ErrorResponse(c, fiber.StatusConflict, err.Error())
		}
		if errors.Is(err, domain.ErrCourseFull) || errors.Is(err, domain.ErrSKSLimitExceeded) {
			var sksErr *domain.SKSLimitError
			if errors.As(err, &sksErr) {
				return helper.ErrorResponse(c, fiber.StatusUnprocessableEntity, sksErr.Error())
			}
			return helper.ErrorResponse(c, fiber.StatusUnprocessableEntity, err.Error())
		}

		// Fallback untuk error tidak terduga / error database aseli
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Terjadi kesalahan internal pada server")
	}

	return helper.SuccessResponse(c, fiber.StatusCreated, "Berhasil membuat KRS", nil)
}

// Delete menangani rute DELETE /api/v1/enrollments/:id
func (h *EnrollmentHandler) Delete(c *fiber.Ctx) error {
	// Ambil user_id dari JWT Locals
	userIDFloat, ok := c.Locals("user_id").(float64)
	if !ok {
		return helper.ErrorResponse(c, fiber.StatusUnauthorized, "Token JWT tidak valid atau user_id tidak ditemukan")
	}
	userID := uint(userIDFloat)

	// Parsing parameter ID
	enrollmentID, err := c.ParamsInt("id")
	if err != nil {
		return helper.ValidationErrorResponse(c, []string{"ID KRS tidak valid"})
	}

	// Panggil Service
	err = h.enrollmentService.CancelEnrollment(userID, uint(enrollmentID))
	if err != nil {
		if errors.Is(err, domain.ErrStudentNotFound) {
			return helper.ErrorResponse(c, fiber.StatusNotFound, err.Error())
		}
		if errors.Is(err, domain.ErrEnrollmentNotFound) {
			return helper.ErrorResponse(c, fiber.StatusNotFound, err.Error())
		}
		if errors.Is(err, domain.ErrEnrollmentForbidden) {
			return helper.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		}
		// Fallback untuk error tidak terduga
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Terjadi kesalahan internal pada server")
	}

	return c.SendStatus(fiber.StatusNoContent)
}
