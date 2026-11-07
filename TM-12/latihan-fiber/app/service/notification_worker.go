package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/hibiken/asynq"
	"latihan-fiber/app/model"
	"latihan-fiber/app/repository"
	"latihan-fiber/helper"
)

// NotificationWorker mengerjakan task "notification:send".
//
// Bagi worker, Asynq adalah framework-nya — sama seperti Fiber bagi API.
// Karena itu method di sini menerima *asynq.Task, seperti method service
// lain menerima *fiber.Ctx.
type NotificationWorker struct {
	users         repository.UserRepository
	courses       repository.CourseRepository
	notifications repository.NotificationRepository
	tx            *repository.TxManager
	mailer        helper.Mailer
	logger        *slog.Logger
}

func NewNotificationWorker(
	users repository.UserRepository,
	courses repository.CourseRepository,
	notifications repository.NotificationRepository,
	tx *repository.TxManager,
	mailer helper.Mailer,
	logger *slog.Logger,
) *NotificationWorker {
	return &NotificationWorker{users: users, courses: courses, notifications: notifications,
		tx: tx, mailer: mailer, logger: logger}
}

// HandleSend mengembalikan error untuk meminta Asynq melakukan retry nanti,
// atau error yang membungkus asynq.SkipRetry untuk menyatakan "jangan di-retry,
// task ini tidak mungkin berhasil".
//
// Asynq menjamin setiap task dijalankan PALING SEDIKIT satu kali, bukan
// TEPAT satu kali. Task yang sama bisa datang dua kali, misalnya bila
// worker mati setelah mengirim email tetapi sebelum sempat melapor ke Redis.
// Karena itu handler ini harus idempoten: menjalankannya dua kali untuk
// task yang sama harus berakibat sama dengan menjalankannya sekali.
func (w *NotificationWorker) HandleSend(ctx context.Context, t *asynq.Task) error {
	var job model.NotificationJob
	if err := json.Unmarshal(t.Payload(), &job); err != nil {
		return fmt.Errorf("%w: payload task rusak: %v", asynq.SkipRetry, err)
	}

	if job.EventType == model.EventCourseQuotaChanged {
		var event model.CourseQuotaChangedEvent
		if err := json.Unmarshal(job.Payload, &event); err != nil {
			return fmt.Errorf("%w: payload quota event %d rusak: %v", asynq.SkipRetry, job.OutboxID, err)
		}

		students, err := w.courses.FindEnrolledStudents(ctx, event.CourseID)
		if err != nil {
			return err
		}

		for _, student := range students {
			subject := fmt.Sprintf("KRS: kuota %s %s berubah menjadi %d", event.CourseCode, event.CourseName, event.NewQuota)
			body := fmt.Sprintf("Halo %s, kuota untuk kelas %s %s telah diubah menjadi %d peserta.",
				student.Username, event.CourseCode, event.CourseName, event.NewQuota)
			msg := model.Message{To: student.Email, Subject: subject, Body: body}

			alreadySent := false
			err := w.tx.WithTxOnce(ctx, func(tx repository.DBTX) error {
				claimed, err := w.notifications.WithTx(tx).Claim(ctx, model.Notification{
					OutboxID: job.OutboxID, UserID: student.ID, Channel: "email",
					Subject: msg.Subject, Body: msg.Body,
				})
				if err != nil {
					return err
				}
				if !claimed {
					alreadySent = true
					return nil
				}
				return w.mailer.Send(ctx, msg)
			})
			if errors.Is(err, helper.ErrPermanentDelivery) {
				w.logger.Warn("quota_notification_permanent_failure",
					slog.Int("user_id", student.ID),
					slog.String("error", err.Error()))
				continue
			}
			if err != nil {
				return err
			}

			retry, _ := asynq.GetRetryCount(ctx)
			w.logger.Info("notification_processed",
				slog.Int64("outbox_id", job.OutboxID),
				slog.String("event", job.EventType),
				slog.Int("user_id", student.ID),
				slog.Int("retry", retry),
				slog.Bool("already_sent", alreadySent))
		}
		return nil
	}

	var event model.EnrollmentEvent
	if err := json.Unmarshal(job.Payload, &event); err != nil {
		return fmt.Errorf("%w: payload event %d rusak: %v", asynq.SkipRetry, job.OutboxID, err)
	}

	user, err := w.users.FindByID(ctx, event.UserID)
	if errors.Is(err, repository.ErrNotFound) {
		// Akunnya sudah dihapus. Melakukan retry tidak akan memunculkannya.
		return fmt.Errorf("%w: user %d tidak ditemukan", asynq.SkipRetry, event.UserID)
	}
	if err != nil {
		return err // database sedang bermasalah: layak di-retry
	}

	msg, err := BuildMessage(job.EventType, event, user)
	if err != nil {
		return fmt.Errorf("%w: %v", asynq.SkipRetry, err)
	}

	alreadySent := false
	err = w.tx.WithTxOnce(ctx, func(tx repository.DBTX) error {
		// Claim dan pengiriman berada di dalam SATU transaction:
		// - pengiriman gagal -> claim ikut dibatalkan -> attempt
		// berikutnya masih dapat mengirim;
		// - pengiriman sukses -> claim di-commit -> attempt berikutnya
		// melihat claim ini dan tidak mengirim lagi.
		claimed, err := w.notifications.WithTx(tx).Claim(ctx, model.Notification{
			OutboxID: job.OutboxID, UserID: user.ID, Channel: "email",
			Subject: msg.Subject, Body: msg.Body,
		})
		if err != nil {
			return err
		}
		if !claimed {
			alreadySent = true
			return nil
		}
		return w.mailer.Send(ctx, msg)
	})

	if errors.Is(err, helper.ErrPermanentDelivery) {
		return fmt.Errorf("%w: %v", asynq.SkipRetry, err)
	}
	if err != nil {
		return err
	}

	retry, _ := asynq.GetRetryCount(ctx)
	w.logger.Info("notification_processed",
		slog.Int64("outbox_id", job.OutboxID),
		slog.String("event", job.EventType),
		slog.Int("retry", retry),
		slog.Bool("already_sent", alreadySent))
	return nil
}
