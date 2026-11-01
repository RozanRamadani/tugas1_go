package service

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"


	"github.com/gofiber/fiber/v2"

	"latihan-fiber/app/model"
	"latihan-fiber/app/repository"
	"latihan-fiber/helper"
)

type UserService struct {
	repo  repository.UserRepository
	perms *helper.PermissionSet
}

func NewUserService(
	repo repository.UserRepository,
	perms *helper.PermissionSet,
) *UserService {
	return &UserService{repo: repo, perms: perms}
}

// reqCtx memberi batas waktu untuk setiap operasi basis data.
func reqCtx(c *fiber.Ctx) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.UserContext(), 5*time.Second)
}

func translateError(c *fiber.Ctx, err error, defaultMsg string) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return helper.Fail(c, fiber.StatusNotFound, "user tidak ditemukan")
	case errors.Is(err, repository.ErrDuplicate):
		return helper.Fail(c, fiber.StatusConflict, "username sudah dipakai")
	default:
		return helper.Fail(c, fiber.StatusInternalServerError, defaultMsg)
	}
}

func paramID(c *fiber.Ctx) (int, bool) {
	idStr := c.Params("id")
	if idStr == "" {
		return 0, false
	}
	val, err := strconv.Atoi(idStr)
	if err != nil || val < 1 {
		return 0, false
	}
	return val, true
}


// ---------- GET /users ----------
func (s *UserService) List(c *fiber.Ctx) error {
	ctx, cancel := reqCtx(c)
	defer cancel()

	q := model.ListQuery{
		Page:  c.QueryInt("page", 1),
		Limit: c.QueryInt("limit", 10),
	}

	users, total, err := s.repo.FindAll(ctx, q)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal mengambil data user")
	}

	return helper.Success(c, fiber.StatusOK, "daftar user berhasil diambil", fiber.Map{
		"users": users,
		"total": total,
	})
}

// ---------- GET /users/:id ----------
func (s *UserService) Get(c *fiber.Ctx) error {
	ctx, cancel := reqCtx(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}

	id, valid := paramID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	// IMPORTANT: Pemeriksaan hak akses dilakukan SEBELUM data diambil dari database.
	// Jika dibalik (ambil data dulu baru periksa), penyerang dapat mengukur selisih waktu (timing attack)
	// untuk mengetahui apakah ID tersebut ada di database atau tidak.
	if !CanAccessUser(current, id, s.perms, "user:read:any") {
		return helper.Fail(c, fiber.StatusForbidden, "tidak berhak mengakses data user lain")
	}

	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateError(c, err, "gagal mengambil data user")
	}

	return helper.Success(c, fiber.StatusOK, "user ditemukan", user)
}

// ---------- POST /users ----------
func (s *UserService) Create(c *fiber.Ctx) error {
	ctx, cancel := reqCtx(c)
	defer cancel()

	var req model.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(req.Email)

	errs := map[string]string{}
	if req.Username == "" {
		errs["username"] = "wajib diisi"
	}
	if !strings.Contains(req.Email, "@") {
		errs["email"] = "format email tidak valid"
	}
	if len(req.Password) < 8 {
		errs["password"] = "minimal 8 karakter"
	}
	if len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	baru, err := s.repo.Create(ctx, model.User{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
		IsActive: true,
	})
	if err != nil {
		return translateError(c, err, "gagal menyimpan user")
	}

	return helper.Success(c, fiber.StatusCreated, "user berhasil dibuat", baru)
}

// ---------- PUT /users/:id ----------
func (s *UserService) Replace(c *fiber.Ctx) error {
	ctx, cancel := reqCtx(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}

	id, valid := paramID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	if !CanAccessUser(current, id, s.perms, "user:update:any") {
		return helper.Fail(c, fiber.StatusForbidden, "tidak berhak mengubah data user lain")
	}

	var req model.ReplaceUserRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	hasil, err := s.repo.Update(ctx, model.User{
		ID:       id,
		Username: req.Username,
		Email:    req.Email,
		IsActive: req.IsActive,
	})
	if err != nil {
		return translateError(c, err, "gagal memperbarui user")
	}

	return helper.Success(c, fiber.StatusOK, "user berhasil diganti seluruhnya", hasil)
}

// ---------- PATCH /users/:id ----------
func (s *UserService) Patch(c *fiber.Ctx) error {
	ctx, cancel := reqCtx(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}

	id, valid := paramID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	if !CanAccessUser(current, id, s.perms, "user:update:any") {
		return helper.Fail(c, fiber.StatusForbidden, "tidak berhak mengubah data user lain")
	}

	var req model.PatchUserRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	saatIni, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateError(c, err, "gagal mengambil data user")
	}

	if req.Username != nil {
		saatIni.Username = *req.Username
	}
	if req.Email != nil {
		saatIni.Email = *req.Email
	}
	if req.IsActive != nil {
		saatIni.IsActive = *req.IsActive
	}

	hasil, err := s.repo.Update(ctx, saatIni)
	if err != nil {
		return translateError(c, err, "gagal memperbarui user")
	}

	return helper.Success(c, fiber.StatusOK, "user berhasil diperbarui sebagian", hasil)
}

// ---------- PATCH /users/:id/role ----------
func (s *UserService) AssignRole(c *fiber.Ctx) error {
	ctx, cancel := reqCtx(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}

	id, valid := paramID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	var req model.AssignRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	if errs := ValidateAssignRole(current, id, req, s.perms); len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	result, err := s.repo.UpdateRole(ctx, id, strings.TrimSpace(req.Role))
	if err != nil {
		return translateError(c, err, "gagal mengubah role user")
	}

	return helper.Success(c, fiber.StatusOK, "role user berhasil diubah", result)
}

// ---------- DELETE /users/:id ----------
func (s *UserService) Delete(c *fiber.Ctx) error {
	ctx, cancel := reqCtx(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}

	id, valid := paramID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	// Aturan Keamanan: Memiliki permission menghapus tidak berarti boleh menghapus akun sendiri
	if current.UserID == id {
		return helper.Fail(c, fiber.StatusForbidden, "tidak boleh menghapus akun sendiri")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return translateError(c, err, "gagal menghapus user")
	}

	return c.SendStatus(fiber.StatusNoContent)
}
