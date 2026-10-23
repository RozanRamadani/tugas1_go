package service

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"latihan-fiber/app/model"
	"latihan-fiber/app/repository"
	"latihan-fiber/helper"
)

type SupervisorService struct {
	supervisors repository.SupervisorRepository
	tx          *repository.TxManager
}

func NewSupervisorService(supervisors repository.SupervisorRepository, tx *repository.TxManager) *SupervisorService {
	return &SupervisorService{supervisors: supervisors, tx: tx}
}

// ---------- GET /supervisors ----------
func (s *SupervisorService) List(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	data, err := s.supervisors.FindAll(ctx)
	if err != nil {
		return helper.Internal(err)
	}
	return helper.Success(c, fiber.StatusOK, "daftar pembimbing berhasil diambil", data)
}

// ---------- GET /supervisors/:id ----------
func (s *SupervisorService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	data, err := s.supervisors.FindByID(ctx, id)
	if err != nil {
		return translateSupervisorError(err)
	}
	return helper.Success(c, fiber.StatusOK, "pembimbing ditemukan", data)
}

// ---------- POST /supervisors/:id/assign ----------
// D.2 Butir 2: Penetapan pembimbing. Wajib UPDATE bersyarat untuk kuota dan kunci row students.
func (s *SupervisorService) Assign(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	supervisorID, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	var req model.AssignSupervisorRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	var updatedSupervisor model.Supervisor

	err := s.tx.WithTx(ctx, func(tx repository.DBTX) error {
		repo := s.supervisors.WithTx(tx)

		// 1. Kunci baris student untuk mencegah mahasiswa ditetapkan ke dua pembimbing secara serentak
		if err := repo.LockStudent(ctx, req.StudentID); err != nil {
			return err
		}

		// 2. Ambil kuota pembimbing secara atomik dengan UPDATE bersyarat
		sup, err := repo.IncrementSupervised(ctx, supervisorID)
		if err != nil {
			return err
		}
		updatedSupervisor = sup

		// 3. Catat penetapan pembimbing (dijaga UNIQUE constraint di supervisions)
		return repo.InsertSupervision(ctx, req.StudentID, supervisorID)
	})

	if err != nil {
		return translateSupervisorError(err)
	}

	return helper.Created(c, "berhasil menetapkan dosen pembimbing", updatedSupervisor,
		"/api/v1/supervisors/"+strconv.Itoa(supervisorID))
}

// ---------- POST /supervisions/switch ----------
// D.2 Butir 3: Mengganti pembimbing. Wajib atomik dan bebas deadlock (urutan ID menaik).
func (s *SupervisorService) Switch(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var req model.SwitchSupervisorRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	var newSupervisor model.Supervisor

	err := s.tx.WithTx(ctx, func(tx repository.DBTX) error {
		repo := s.supervisors.WithTx(tx)

		// 1. Kunci data mahasiswa
		if err := repo.LockStudent(ctx, req.StudentID); err != nil {
			return err
		}

		// 2. Periksa pembimbing mahasiswa saat ini
		currentSupervision, err := repo.FindSupervisionByStudent(ctx, req.StudentID)
		if err != nil {
			return err
		}
		if currentSupervision == nil {
			return repository.ErrNotSupervised
		}

		if currentSupervision.SupervisorID == req.ToSupervisorID {
			return helper.Validation(map[string]string{
				"to_supervisor_id": "tidak boleh sama dengan pembimbing saat ini",
			})
		}

		// 3. Kunci KEDUA pembimbing dengan urutan ID menaik untuk mencegah deadlock
		if err := repo.LockSupervisors(ctx, currentSupervision.SupervisorID, req.ToSupervisorID); err != nil {
			return err
		}

		// 4. Kurangi bimbingan pembimbing lama
		if _, err := repo.DecrementSupervised(ctx, currentSupervision.SupervisorID); err != nil {
			return err
		}

		// 5. Tambah bimbingan pembimbing baru secara atomik
		sup, err := repo.IncrementSupervised(ctx, req.ToSupervisorID)
		if err != nil {
			return err
		}
		newSupervisor = sup

		// 6. Update data supervisi ke pembimbing baru
		return repo.UpdateSupervision(ctx, req.StudentID, req.ToSupervisorID)
	})

	if err != nil {
		return translateSupervisorError(err)
	}

	return helper.Success(c, fiber.StatusOK, "berhasil mengganti dosen pembimbing", newSupervisor)
}

// ---------- PATCH /supervisors/:id ----------
// D.2 Butir 4: Ubah kuota dengan optimistic locking.
func (s *SupervisorService) UpdateQuota(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	var req model.UpdateSupervisorQuotaRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	updated, err := s.supervisors.UpdateQuota(ctx, id, req.Quota, req.Version)
	if err != nil {
		return translateSupervisorError(err)
	}

	return helper.Success(c, fiber.StatusOK, "kuota pembimbing berhasil diubah", updated)
}

func translateSupervisorError(err error) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return helper.NotFound("data tidak ditemukan")
	case errors.Is(err, repository.ErrSupervisorFull):
		return helper.NewError(fiber.StatusConflict, "SUPERVISOR_FULL", "kuota dosen pembimbing sudah penuh")
	case errors.Is(err, repository.ErrAlreadySupervised):
		return helper.NewError(fiber.StatusConflict, "ALREADY_SUPERVISED", "mahasiswa sudah memiliki dosen pembimbing")
	case errors.Is(err, repository.ErrNotSupervised):
		return helper.NewError(fiber.StatusConflict, "NOT_SUPERVISED", "mahasiswa belum memiliki dosen pembimbing")
	case errors.Is(err, repository.ErrVersionConflict):
		return helper.NewError(fiber.StatusConflict, "VERSION_CONFLICT", "data pembimbing sudah diubah pihak lain, muat ulang lalu coba lagi")
	case errors.Is(err, repository.ErrQuotaBelowSupervised):
		return helper.NewError(fiber.StatusConflict, "QUOTA_BELOW_SUPERVISED", "kuota tidak boleh lebih kecil dari jumlah mahasiswa bimbingan saat ini")
	default:
		return helper.Internal(err)
	}
}
