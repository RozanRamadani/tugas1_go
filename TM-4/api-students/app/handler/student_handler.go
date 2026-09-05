package handler

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"api-students/app/model"
	"api-students/app/repository"
)

type StudentHandler struct {
	repo repository.StudentRepository
}

func NewStudentHandler(
	repo repository.StudentRepository,
) *StudentHandler {
	return &StudentHandler{
		repo: repo,
	}
}

// ============================================================
// GET /students
// ============================================================

func (h *StudentHandler) FindAll(c *fiber.Ctx) error {

	q := parseListQuery(c)

	students, total, err := h.repo.FindAll(
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

func (h *StudentHandler) FindByID(c *fiber.Ctx) error {

	id := c.Params("id")

	if strings.TrimSpace(id) == "" {
		return fail(
			c,
			fiber.StatusBadRequest,
			"id wajib diisi",
		)
	}

	student, err := h.repo.FindByID(
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

	errs := map[string]string{}

	req.ID = strings.TrimSpace(req.ID)
	req.NIM = strings.TrimSpace(req.NIM)
	req.Name = strings.TrimSpace(req.Name)

	if req.ID == "" {
		errs["id"] = "wajib diisi"
	}

	if req.NIM == "" {
		errs["nim"] = "wajib diisi"
	}

	if req.Name == "" {
		errs["name"] = "wajib diisi"
	}

	if req.Grade < 0 || req.Grade > 100 {
		errs["grade"] = "harus berada di antara 0 dan 100"
	}

	if len(errs) > 0 {
		return failValidation(c, errs)
	}

	student := model.Student{
		ID:       req.ID,
		NIM:      req.NIM,
		Name:     req.Name,
		Grade:    req.Grade,
		IsActive: req.IsActive,
	}

	result, err := h.repo.Create(
		c.Context(),
		student,
	)

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

	id := c.Params("id")

	var req model.ReplaceStudentRequest

	if err := c.BodyParser(&req); err != nil {
		return fail(
			c,
			fiber.StatusBadRequest,
			"body harus berupa JSON yang valid",
		)
	}

	errs := map[string]string{}

	req.NIM = strings.TrimSpace(req.NIM)
	req.Name = strings.TrimSpace(req.Name)

	if req.NIM == "" {
		errs["nim"] = "wajib diisi pada PUT"
	}

	if req.Name == "" {
		errs["name"] = "wajib diisi pada PUT"
	}

	if req.Grade < 0 || req.Grade > 100 {
		errs["grade"] = "harus berada di antara 0 dan 100"
	}

	if len(errs) > 0 {
		return failValidation(c, errs)
	}

	student := model.Student{
		ID:       id,
		NIM:      req.NIM,
		Name:     req.Name,
		Grade:    req.Grade,
		IsActive: req.IsActive,
	}

	result, err := h.repo.Update(
		c.Context(),
		id,
		student,
	)

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

	id := c.Params("id")

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

	// Ambil data lama terlebih dahulu.
	current, err := h.repo.FindByID(
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

	// Hanya field yang dikirim yang diubah.

	if req.NIM != nil {
		nim := strings.TrimSpace(*req.NIM)

		if nim == "" {
			return failValidation(
				c,
				map[string]string{
					"nim": "tidak boleh kosong",
				},
			)
		}

		current.NIM = nim
	}

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)

		if name == "" {
			return failValidation(
				c,
				map[string]string{
					"name": "tidak boleh kosong",
				},
			)
		}

		current.Name = name
	}

	if req.Grade != nil {

		if *req.Grade < 0 || *req.Grade > 100 {
			return failValidation(
				c,
				map[string]string{
					"grade": "harus berada di antara 0 dan 100",
				},
			)
		}

		current.Grade = *req.Grade
	}

	if req.IsActive != nil {
		current.IsActive = *req.IsActive
	}

	result, err := h.repo.Update(
		c.Context(),
		id,
		current,
	)

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
		"student berhasil diperbarui sebagian",
		result,
	)
}

// ============================================================
// DELETE /students/:id
// ============================================================

func (h *StudentHandler) Delete(c *fiber.Ctx) error {

	id := c.Params("id")

	err := h.repo.Delete(
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

		value, err := strconv.ParseBool(raw)

		if err == nil {
			q.IsActive = &value
		}
	}

	return q
}
