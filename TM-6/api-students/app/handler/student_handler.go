package handler

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"api-students/app/model"
	"api-students/app/repository"
	"api-students/app/service"
	"api-students/helper"
)

type StudentHandler struct {
	service *service.StudentService
}

func NewStudentHandler(service *service.StudentService) *StudentHandler {
	return &StudentHandler{
		service: service,
	}
}

// ============================================================
// GET /api/v1/students
// ============================================================

func (h *StudentHandler) List(c *fiber.Ctx) error {
	helperQuery := helper.ParseListQuery(c)

	query := model.ListQuery{
		Page:     helperQuery.Page,
		Limit:    helperQuery.Limit,
		Search:   helperQuery.Search,
		Sort:     helperQuery.Sort,
		Order:    helperQuery.Order,
		IsActive: helperQuery.IsActive,
	}

	students, total, err := h.service.List(
		c.Context(),
		query,
	)

	if err != nil {
		return fail(
			c,
			fiber.StatusInternalServerError,
			"gagal mengambil data student",
		)
	}

	totalPages := 0

	if query.Limit > 0 {
		totalPages = (total + query.Limit - 1) / query.Limit
	}

	return okList(
		c,
		"berhasil mengambil data student",
		students,
		&Meta{
			Page:       query.Page,
			Limit:      query.Limit,
			Total:      total,
			TotalPages: totalPages,
		},
	)
}

// ============================================================
// GET /api/v1/students/:id
// ============================================================

func (h *StudentHandler) Get(c *fiber.Ctx) error {
	id := c.Params("id")

	if _, err := strconv.Atoi(id); err != nil {
		return fail(
			c,
			fiber.StatusBadRequest,
			"id student tidak valid",
		)
	}

	currentUser, userExists := helper.CurrentUser(c)

	if !userExists {
		return fail(
			c,
			fiber.StatusUnauthorized,
			"unauthorized",
		)
	}

	student, err := h.service.Get(
		c.Context(),
		id,
		currentUser,
	)

	if err != nil {
		if errors.Is(err, service.ErrForbidden) {
			return fail(
				c,
				fiber.StatusForbidden,
				"forbidden",
			)
		}

		if errors.Is(err, repository.ErrNotFound) {
			return fail(
				c,
				fiber.StatusNotFound,
				"student tidak ditemukan",
			)
		}

		return fail(
			c,
			fiber.StatusInternalServerError,
			"gagal mengambil data student",
		)
	}

	return ok(
		c,
		"berhasil mengambil data student",
		student,
	)
}

// ============================================================
// POST /api/v1/students
// ============================================================

func (h *StudentHandler) Create(c *fiber.Ctx) error {
	var req model.CreateStudentRequest

	if err := c.BodyParser(&req); err != nil {
		return fail(
			c,
			fiber.StatusBadRequest,
			"request body tidak valid",
		)
	}

	currentUser, userExists := helper.CurrentUser(c)

	if !userExists {
		return fail(
			c,
			fiber.StatusUnauthorized,
			"unauthorized",
		)
	}

	student, err := h.service.Create(
		c.Context(),
		req,
		currentUser,
	)

	if err != nil {
		var validationErr service.ValidationError

		if errors.As(err, &validationErr) {
			return failValidation(
				c,
				validationErr.Fields,
			)
		}

		if errors.Is(err, repository.ErrDuplicate) {
			return fail(
				c,
				fiber.StatusConflict,
				"student sudah ada",
			)
		}

		return fail(
			c,
			fiber.StatusInternalServerError,
			"gagal membuat student",
		)
	}

	return created(
		c,
		"student berhasil dibuat",
		student,
		"/api/v1/students/"+student.ID,
	)
}

// ============================================================
// PUT /api/v1/students/:id
// ============================================================

func (h *StudentHandler) Replace(c *fiber.Ctx) error {
	id := c.Params("id")

	if _, err := strconv.Atoi(id); err != nil {
		return fail(
			c,
			fiber.StatusBadRequest,
			"id student tidak valid",
		)
	}

	var req model.ReplaceStudentRequest

	if err := c.BodyParser(&req); err != nil {
		return fail(
			c,
			fiber.StatusBadRequest,
			"request body tidak valid",
		)
	}

	currentUser, userExists := helper.CurrentUser(c)

	if !userExists {
		return fail(
			c,
			fiber.StatusUnauthorized,
			"unauthorized",
		)
	}

	student, err := h.service.Replace(
		c.Context(),
		id,
		req,
		currentUser,
	)

	if err != nil {
		if errors.Is(err, service.ErrForbidden) {
			return fail(
				c,
				fiber.StatusForbidden,
				"forbidden",
			)
		}

		var validationErr service.ValidationError

		if errors.As(err, &validationErr) {
			return failValidation(
				c,
				validationErr.Fields,
			)
		}

		if errors.Is(err, repository.ErrNotFound) {
			return fail(
				c,
				fiber.StatusNotFound,
				"student tidak ditemukan",
			)
		}

		if errors.Is(err, repository.ErrDuplicate) {
			return fail(
				c,
				fiber.StatusConflict,
				"student sudah ada",
			)
		}

		return fail(
			c,
			fiber.StatusInternalServerError,
			"gagal memperbarui student",
		)
	}

	return ok(
		c,
		"student berhasil diperbarui",
		student,
	)
}

// ============================================================
// PATCH /api/v1/students/:id
// ============================================================

func (h *StudentHandler) Patch(c *fiber.Ctx) error {
	id := c.Params("id")

	if _, err := strconv.Atoi(id); err != nil {
		return fail(
			c,
			fiber.StatusBadRequest,
			"id student tidak valid",
		)
	}

	var req model.PatchStudentRequest

	if err := c.BodyParser(&req); err != nil {
		return fail(
			c,
			fiber.StatusBadRequest,
			"request body tidak valid",
		)
	}

	if req.NIM == nil &&
		req.Name == nil &&
		req.Grade == nil &&
		req.IsActive == nil {

		return fail(
			c,
			fiber.StatusBadRequest,
			"tidak ada field yang diperbarui",
		)
	}

	currentUser, userExists := helper.CurrentUser(c)

	if !userExists {
		return fail(
			c,
			fiber.StatusUnauthorized,
			"unauthorized",
		)
	}

	student, err := h.service.Patch(
		c.Context(),
		id,
		req,
		currentUser,
	)

	if err != nil {
		if errors.Is(err, service.ErrForbidden) {
			return fail(
				c,
				fiber.StatusForbidden,
				"forbidden",
			)
		}

		var validationErr service.ValidationError

		if errors.As(err, &validationErr) {
			return failValidation(
				c,
				validationErr.Fields,
			)
		}

		if errors.Is(err, repository.ErrNotFound) {
			return fail(
				c,
				fiber.StatusNotFound,
				"student tidak ditemukan",
			)
		}

		if errors.Is(err, repository.ErrDuplicate) {
			return fail(
				c,
				fiber.StatusConflict,
				"student sudah ada",
			)
		}

		return fail(
			c,
			fiber.StatusInternalServerError,
			"gagal memperbarui student",
		)
	}

	return ok(
		c,
		"student berhasil diperbarui",
		student,
	)
}

// ============================================================
// DELETE /api/v1/students/:id
// ============================================================

func (h *StudentHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")

	if _, err := strconv.Atoi(id); err != nil {
		return fail(
			c,
			fiber.StatusBadRequest,
			"id student tidak valid",
		)
	}

	currentUser, userExists := helper.CurrentUser(c)

	if !userExists {
		return fail(
			c,
			fiber.StatusUnauthorized,
			"unauthorized",
		)
	}

	err := h.service.Delete(
		c.Context(),
		id,
		currentUser,
	)

	if err != nil {
		if errors.Is(err, service.ErrForbidden) {
			return fail(
				c,
				fiber.StatusForbidden,
				"forbidden",
			)
		}

		if errors.Is(err, repository.ErrNotFound) {
			return fail(
				c,
				fiber.StatusNotFound,
				"student tidak ditemukan",
			)
		}

		return fail(
			c,
			fiber.StatusInternalServerError,
			"gagal menghapus student",
		)
	}

	return noContent(c)
}
