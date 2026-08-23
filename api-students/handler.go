package main

import (
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Data sementara disimpan di memory.
// Restart server = data kembali kosong.
var students []Student
var nextID = 1

// Cari posisi student berdasarkan ID.
func findStudentIndex(id int) int {
	for i := range students {
		if students[i].ID == id {
			return i
		}
	}

	return -1
}

// Cari student berdasarkan NIM.
// NIM harus unik.
func findStudentByNIM(nim string) int {
	for i := range students {
		if strings.EqualFold(students[i].NIM, nim) {
			return i
		}
	}

	return -1
}

// Ambil ID dari URL /students/:id.
func paramID(c *fiber.Ctx) (int, bool) {
	id, err := strconv.Atoi(c.Params("id"))

	if err != nil || id < 1 {
		return 0, false
	}

	return id, true
}

// GET /api/v1/students
// Daftar + search + filter + sort + pagination.
func listStudents(c *fiber.Ctx) error {
	q := parseListQuery(c)

	hasil := []Student{}

	// 1. Filter.
	for _, s := range students {

		if q.IsActive != nil && s.IsActive != *q.IsActive {
			continue
		}

		// Search berdasarkan nama, case-insensitive.
		if q.Search != "" &&
			!strings.Contains(
				strings.ToLower(s.Name),
				strings.ToLower(q.Search),
			) {
			continue
		}

		hasil = append(hasil, s)
	}

	// 2. Sort dengan whitelist field.
	sort.SliceStable(hasil, func(i, j int) bool {
		var kecil bool

		switch q.Sort {
		case "nim":
			kecil = hasil[i].NIM < hasil[j].NIM

		case "name":
			kecil = hasil[i].Name < hasil[j].Name

		case "grade":
			kecil = hasil[i].Grade < hasil[j].Grade

		default:
			// Default = sort berdasarkan ID.
			kecil = hasil[i].ID < hasil[j].ID
		}

		if q.Order == "desc" {
			return !kecil
		}

		return kecil
	})

	// 3. Pagination.
	total := len(hasil)
	totalPages := (total + q.Limit - 1) / q.Limit

	mulai := (q.Page - 1) * q.Limit

	if mulai > total {
		mulai = total
	}

	akhir := mulai + q.Limit

	if akhir > total {
		akhir = total
	}

	return okList(
		c,
		"daftar student berhasil diambil",
		hasil[mulai:akhir],
		&Meta{
			Page:       q.Page,
			Limit:      q.Limit,
			Total:      total,
			TotalPages: totalPages,
		},
	)
}

// GET /api/v1/students/:id
func getStudent(c *fiber.Ctx) error {
	id, valid := paramID(c)

	if !valid {
		return fail(
			c,
			fiber.StatusBadRequest,
			"id harus berupa angka positif",
		)
	}

	i := findStudentIndex(id)

	if i == -1 {
		return fail(
			c,
			fiber.StatusNotFound,
			"student tidak ditemukan",
		)
	}

	return ok(c, "student ditemukan", students[i])
}

// POST /api/v1/students
func createStudent(c *fiber.Ctx) error {
	var req CreateStudentRequest

	// JSON harus valid.
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

	// Validasi field.
	if req.NIM == "" {
		errs["nim"] = "wajib diisi"
	}

	if req.Name == "" {
		errs["name"] = "wajib diisi"
	}

	if req.Grade < 0 || req.Grade > 100 {
		errs["grade"] = "harus berada di antara 0 dan 100"
	}

	// Kalau NIM sudah digunakan, kembalikan 409 Conflict.
	if findStudentByNIM(req.NIM) != -1 {
		return fail(
			c,
			fiber.StatusConflict,
			"NIM sudah digunakan",
		)
	}

	if len(errs) > 0 {
		return failValidation(c, errs)
	}

	// Buat student baru.
	baru := Student{
		ID:        nextID,
		NIM:       req.NIM,
		Name:      req.Name,
		Grade:     req.Grade,
		IsActive:  req.IsActive,
		CreatedAt: time.Now(),
	}

	students = append(students, baru)
	nextID++

	return created(
		c,
		"student berhasil dibuat",
		baru,
		"/api/v1/students/"+strconv.Itoa(baru.ID),
	)
}

// PUT = mengganti seluruh isi student.
func replaceStudent(c *fiber.Ctx) error {
	id, valid := paramID(c)

	if !valid {
		return fail(
			c,
			fiber.StatusBadRequest,
			"id harus berupa angka positif",
		)
	}

	i := findStudentIndex(id)

	if i == -1 {
		return fail(
			c,
			fiber.StatusNotFound,
			"student tidak ditemukan",
		)
	}

	var req ReplaceStudentRequest

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

	// Cek NIM hanya jika berubah.
	duplicate := findStudentByNIM(req.NIM)

	if duplicate != -1 && duplicate != i {
		errs["nim"] = "sudah digunakan"
	}

	if len(errs) > 0 {
		return failValidation(c, errs)
	}

	// Semua field diganti.
	students[i].NIM = req.NIM
	students[i].Name = req.Name
	students[i].Grade = req.Grade
	students[i].IsActive = req.IsActive

	return ok(
		c,
		"student berhasil diganti seluruhnya",
		students[i],
	)
}

// PATCH = hanya field yang dikirim yang diubah.
func patchStudent(c *fiber.Ctx) error {
	id, valid := paramID(c)

	if !valid {
		return fail(
			c,
			fiber.StatusBadRequest,
			"id harus berupa angka positif",
		)
	}

	i := findStudentIndex(id)

	if i == -1 {
		return fail(
			c,
			fiber.StatusNotFound,
			"student tidak ditemukan",
		)
	}

	var req PatchStudentRequest

	if err := c.BodyParser(&req); err != nil {
		return fail(
			c,
			fiber.StatusBadRequest,
			"body harus berupa JSON yang valid",
		)
	}

	// Tidak boleh PATCH dengan body kosong.
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

	// Update NIM jika dikirim.
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

		duplicate := findStudentByNIM(nim)

		if duplicate != -1 && duplicate != i {
			return fail(
				c,
				fiber.StatusConflict,
				"NIM sudah digunakan",
			)
		}

		students[i].NIM = nim
	}

	// Update nama jika dikirim.
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

		students[i].Name = name
	}

	// Update grade jika dikirim.
	if req.Grade != nil {
		if *req.Grade < 0 || *req.Grade > 100 {
			return failValidation(
				c,
				map[string]string{
					"grade": "harus berada di antara 0 dan 100",
				},
			)
		}

		students[i].Grade = *req.Grade
	}

	// Update status jika dikirim.
	if req.IsActive != nil {
		students[i].IsActive = *req.IsActive
	}

	return ok(
		c,
		"student berhasil diperbarui sebagian",
		students[i],
	)
}

// DELETE /api/v1/students/:id
func deleteStudent(c *fiber.Ctx) error {
	id, valid := paramID(c)

	if !valid {
		return fail(
			c,
			fiber.StatusBadRequest,
			"id harus berupa angka positif",
		)
	}

	i := findStudentIndex(id)

	if i == -1 {
		return fail(
			c,
			fiber.StatusNotFound,
			"student tidak ditemukan",
		)
	}

	// Hapus data dari slice.
	students = append(students[:i], students[i+1:]...)

	return noContent(c)
}
