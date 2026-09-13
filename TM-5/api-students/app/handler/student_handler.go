package handler

import (
	"errors"
	"strconv"
	"strings"

	"api-students/app/model"
	"api-students/app/repository"
	"api-students/app/service"

	"github.com/gofiber/fiber/v2"
)

type StudentHandler struct {
	service *service.StudentService
}

func NewStudentHandler(
	studentService *service.StudentService,
) *StudentHandler {
	return &StudentHandler{
		service: studentService,
	}
}

// ============================================================
// GET /students
// ============================================================

func (h *StudentHandler) List(c *fiber.Ctx) error {

	q := parseListQuery(c)

	students, total, err := h.service.List(
		c.Context(),
		q,
	)

	if err != nil {
		return fail(
			c,
			fiber.StatusInternalServerError,
			"gagal mengambil data student",
		)
	}

	totalPages := 0

	if q.Limit > 0 {
		totalPages = (total + q.Limit - 1) / q.Limit
	}

	return okList(
		c,
		"daftar student berhasil diambil",
		students,
		&Meta{
			Page:       q.Page,
			Limit:      q.Limit,
			Total:      total,
			TotalPages: totalPages,
		},
	)
}

// ============================================================
// GET /students/:id
// ============================================================

func (h *StudentHandler) Get(c *fiber.Ctx) error {

	id := strings.TrimSpace(c.Params("id"))

	if id == "" {
		return fail(
			c,
			fiber.StatusBadRequest,
			"id wajib diisi",
		)
	}

	student, err := h.service.Get(
		c.Context(),
		id,
	)

	if errors.Is(err, repository.ErrNotFound) {
		return fail(
			c,
			fiber.StatusNotFound,
			"student tidak ditemukan",
		)
	}

	if err != nil {
		return fail(
			c,
			fiber.StatusInternalServerError,
			"gagal mengambil student",
		)
	}

	return ok(
		c,
		"student ditemukan",
		student,
	)
}

// ============================================================
// POST /students
// ============================================================

func (h *StudentHandler) Create(c *fiber.Ctx) error {

	var req model.CreateStudentRequest

	if err := c.BodyParser(&req); err != nil {
		return fail(
			c,
			fiber.StatusBadRequest,
			"body harus berupa JSON yang valid",
		)
	}

	result, err := h.service.Create(
		c.Context(),
		req,
	)

	var validationErr *service.ValidationError

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
			"NIM atau ID sudah digunakan",
		)
	}

	if err != nil {
		return fail(
			c,
			fiber.StatusInternalServerError,
			"gagal membuat student",
		)
	}

	return created(
		c,
		"student berhasil dibuat",
		result,
		"/api/v1/students/"+result.ID,
	)
}

// ============================================================
// PUT /students/:id
// ============================================================

func (h *StudentHandler) Replace(c *fiber.Ctx) error {

	id := strings.TrimSpace(c.Params("id"))

	if id == "" {
		return fail(
			c,
			fiber.StatusBadRequest,
			"id wajib diisi",
		)
	}

	var req model.ReplaceStudentRequest

	if err := c.BodyParser(&req); err != nil {
		return fail(
			c,
			fiber.StatusBadRequest,
			"body harus berupa JSON yang valid",
		)
	}

	result, err := h.service.Replace(
		c.Context(),
		id,
		req,
	)

	var validationErr *service.ValidationError

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
			"NIM sudah digunakan",
		)
	}

	if err != nil {
		return fail(
			c,
			fiber.StatusInternalServerError,
			"gagal memperbarui student",
		)
	}

	return ok(
		c,
		"student berhasil diganti seluruhnya",
		result,
	)
}

// ============================================================
// PATCH /students/:id
// ============================================================

func (h *StudentHandler) Patch(c *fiber.Ctx) error {

	id := strings.TrimSpace(c.Params("id"))

	if id == "" {
		return fail(
			c,
			fiber.StatusBadRequest,
			"id wajib diisi",
		)
	}

	var req model.PatchStudentRequest

	if err := c.BodyParser(&req); err != nil {
		return fail(
			c,
			fiber.StatusBadRequest,
			"body harus berupa JSON yang valid",
		)
	}

	if req.NIM == nil &&
		req.Name == nil &&
		req.Grade == nil &&
		req.IsActive == nil {

		return fail(
			c,
			fiber.StatusBadRequest,
			"tidak ada field yang diubah",
		)
	}

	result, err := h.service.Patch(
		c.Context(),
		id,
		req,
	)

	if errors.Is(err, repository.ErrNotFound) {
		return fail(
			c,
			fiber.StatusNotFound,
			"student tidak ditemukan",
		)
	}

	var validationErr *service.ValidationError

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
			"NIM sudah digunakan",
		)
	}

	if err != nil {
		return fail(
			c,
			fiber.StatusInternalServerError,
			"gagal memperbarui student",
		)
	}

	return ok(
		c,
		"student berhasil diperbarui sebagian",
		result,
	)
}

// ============================================================
// DELETE /students/:id
// ============================================================

func (h *StudentHandler) Delete(c *fiber.Ctx) error {

	id := strings.TrimSpace(c.Params("id"))

	if id == "" {
		return fail(
			c,
			fiber.StatusBadRequest,
			"id wajib diisi",
		)
	}

	err := h.service.Delete(
		c.Context(),
		id,
	)

	if errors.Is(err, repository.ErrNotFound) {
		return fail(
			c,
			fiber.StatusNotFound,
			"student tidak ditemukan",
		)
	}

	if err != nil {
		return fail(
			c,
			fiber.StatusInternalServerError,
			"gagal menghapus student",
		)
	}

	return noContent(c)
}

// ============================================================
// QUERY
// ============================================================

func parseListQuery(c *fiber.Ctx) model.ListQuery {

	q := model.ListQuery{
		Page:   c.QueryInt("page", 1),
		Limit:  c.QueryInt("limit", 10),
		Search: strings.TrimSpace(c.Query("search")),
		Sort:   c.Query("sort", "id"),
		Order:  strings.ToLower(c.Query("order", "asc")),
	}

	if q.Page < 1 {
		q.Page = 1
	}

	if q.Limit < 1 {
		q.Limit = 10
	}

	if q.Limit > 100 {
		q.Limit = 100
	}

	allowedSort := map[string]bool{
		"id":         true,
		"nim":        true,
		"name":       true,
		"grade":      true,
		"created_at": true,
	}

	if !allowedSort[q.Sort] {
		q.Sort = "id"
	}

	if q.Order != "desc" {
		q.Order = "asc"
	}

	if raw := c.Query("is_active"); raw != "" {
		if value, err := strconv.ParseBool(raw); err == nil {
			q.IsActive = &value
		}
	}

	return q
}
