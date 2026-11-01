package service

import (
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/bson"

	"latihan-fiber/app/model"
	"latihan-fiber/app/repository"
	"latihan-fiber/helper"
)

type PortfolioService struct {
	portfolios repository.PortfolioRepository
	students   repository.StudentRepository
}

func NewPortfolioService(
	portfolios repository.PortfolioRepository,
	students repository.StudentRepository,
) *PortfolioService {
	return &PortfolioService{portfolios: portfolios, students: students}
}

// ---------- GET /students/:id/portfolio ----------
func (s *PortfolioService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	studentID := strings.TrimSpace(c.Params("id"))
	if studentID == "" {
		return helper.BadRequest("id mahasiswa tidak boleh kosong")
	}

	// Validasi lintas database: mahasiswa harus ada di PostgreSQL
	exists, err := s.students.Exists(ctx, studentID)
	if err != nil {
		return helper.Internal(err)
	}
	if !exists {
		return helper.NotFound("mahasiswa tidak ditemukan")
	}

	p, err := s.portfolios.FindByStudentID(ctx, studentID)
	if err != nil {
		if errors.Is(err, repository.ErrPortfolioNotFound) {
			return helper.NotFound("portofolio mahasiswa belum dibuat")
		}
		return helper.Internal(err)
	}

	return helper.Success(c, fiber.StatusOK, "portofolio berhasil diambil", p)
}

// ---------- PUT /students/:id/portfolio ----------
func (s *PortfolioService) Upsert(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	studentID := strings.TrimSpace(c.Params("id"))
	if studentID == "" {
		return helper.BadRequest("id mahasiswa tidak boleh kosong")
	}

	// Validasi lintas database: mahasiswa harus ada di PostgreSQL
	exists, err := s.students.Exists(ctx, studentID)
	if err != nil {
		return helper.Internal(err)
	}
	if !exists {
		return helper.NotFound("mahasiswa tidak ditemukan")
	}

	var req model.UpsertPortfolioRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	req.Headline = strings.TrimSpace(req.Headline)
	req.Bio = strings.TrimSpace(req.Bio)

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	projects := make([]model.Project, 0, len(req.Projects))
	for _, prj := range req.Projects {
		techs := NormalizeTags(prj.Technologies)
		projects = append(projects, model.Project{
			Title:        strings.TrimSpace(prj.Title),
			Technologies: techs,
			Link:         strings.TrimSpace(prj.Link),
			Year:         prj.Year,
		})
	}

	certs := make([]model.Certificate, 0, len(req.Certificates))
	for _, cert := range req.Certificates {
		var certDate time.Time
		if cert.Date != "" {
			parsed, err := time.Parse("2006-01-02", strings.TrimSpace(cert.Date))
			if err != nil {
				parsed, err = time.Parse(time.RFC3339, strings.TrimSpace(cert.Date))
			}
			if err == nil {
				certDate = parsed
			}
		}
		certs = append(certs, model.Certificate{
			Name:   strings.TrimSpace(cert.Name),
			Issuer: strings.TrimSpace(cert.Issuer),
			Date:   certDate,
		})
	}

	p, err := s.portfolios.Upsert(ctx, model.Portfolio{
		StudentID:    studentID,
		Headline:     req.Headline,
		Bio:          req.Bio,
		Projects:     projects,
		Certificates: certs,
	})
	if err != nil {
		return helper.Internal(err)
	}

	return helper.Success(c, fiber.StatusOK, "portofolio berhasil disimpan", p)
}

// ---------- POST /students/:id/portfolio/projects ----------
func (s *PortfolioService) AddProject(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	studentID := strings.TrimSpace(c.Params("id"))
	if studentID == "" {
		return helper.BadRequest("id mahasiswa tidak boleh kosong")
	}

	exists, err := s.students.Exists(ctx, studentID)
	if err != nil {
		return helper.Internal(err)
	}
	if !exists {
		return helper.NotFound("mahasiswa tidak ditemukan")
	}

	var req model.CreateProjectRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	req.Title = strings.TrimSpace(req.Title)
	req.Link = strings.TrimSpace(req.Link)
	req.Technologies = NormalizeTags(req.Technologies)

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	updated, err := s.portfolios.AddProject(ctx, studentID, model.Project{
		Title:        req.Title,
		Technologies: req.Technologies,
		Link:         req.Link,
		Year:         req.Year,
	})
	if err != nil {
		if errors.Is(err, repository.ErrPortfolioNotFound) {
			return helper.NotFound("portofolio mahasiswa belum dibuat")
		}
		if errors.Is(err, repository.ErrProjectLimit) {
			return helper.NewError(fiber.StatusConflict, "PROJECT_LIMIT", "jumlah proyek sudah mencapai batas maksimum 20")
		}
		return helper.Internal(err)
	}

	return helper.Success(c, fiber.StatusCreated, "proyek berhasil ditambahkan", updated)
}

// ---------- DELETE /students/:id/portfolio/projects/:projectId ----------
func (s *PortfolioService) DeleteProject(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	studentID := strings.TrimSpace(c.Params("id"))
	projectID := strings.TrimSpace(c.Params("projectId"))

	if studentID == "" || projectID == "" {
		return helper.BadRequest("id mahasiswa dan id proyek tidak boleh kosong")
	}

	projectOID, err := bson.ObjectIDFromHex(projectID)
	if err != nil {
		return helper.BadRequest("format id proyek tidak valid")
	}

	exists, err := s.students.Exists(ctx, studentID)
	if err != nil {
		return helper.Internal(err)
	}
	if !exists {
		return helper.NotFound("mahasiswa tidak ditemukan")
	}

	updated, err := s.portfolios.DeleteProject(ctx, studentID, projectOID)
	if err != nil {
		if errors.Is(err, repository.ErrPortfolioNotFound) {
			return helper.NotFound("portofolio mahasiswa belum dibuat")
		}
		if errors.Is(err, repository.ErrProjectNotFound) {
			return helper.NotFound("proyek tidak ditemukan")
		}
		return helper.Internal(err)
	}

	return helper.Success(c, fiber.StatusOK, "proyek berhasil dihapus", updated)
}

// ---------- GET /portfolios/stats/technologies ----------
func (s *PortfolioService) TopTechnologies(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	stats, err := s.portfolios.TopTechnologies(ctx, 10)
	if err != nil {
		return helper.Internal(err)
	}

	return helper.Success(c, fiber.StatusOK, "statistik teknologi berhasil diambil", stats)
}
