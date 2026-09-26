package handler

import (
	"github.com/gofiber/fiber/v2"

	"api-students/app/model"
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
	query, err := helper.ParseCursorQuery(c)
	if err != nil {
		return err
	}

	students, meta, err := h.service.List(
		c.Context(),
		query,
	)

	if err != nil {
		return err
	}

	return helper.OKList(
		c,
		"berhasil mengambil data student",
		students,
		meta,
	)
}

// ============================================================
// GET /api/v1/students/:id
// ============================================================

func (h *StudentHandler) Get(c *fiber.Ctx) error {
	id := c.Params("id")

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
		return err
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
		return err
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
		return helper.Unauthorized("unauthorized")
	}

	student, err := h.service.Replace(
		c.Context(),
		id,
		req,
		currentUser,
	)

	if err != nil {
		return err
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

	var req model.PatchStudentRequest

	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("request body tidak valid")
	}

	if req.NIM == nil &&
		req.Name == nil &&
		req.Grade == nil &&
		req.IsActive == nil {

		return helper.BadRequest("tidak ada field yang diperbarui")
	}

	currentUser, userExists := helper.CurrentUser(c)

	if !userExists {
		return helper.Unauthorized("unauthorized")
	}

	student, err := h.service.Patch(
		c.Context(),
		id,
		req,
		currentUser,
	)

	if err != nil {
		return err
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

	currentUser, userExists := helper.CurrentUser(c)

	if !userExists {
		return helper.Unauthorized("unauthorized")
	}

	err := h.service.Delete(
		c.Context(),
		id,
		currentUser,
	)

	if err != nil {
		return err
	}

	return noContent(c)
}
