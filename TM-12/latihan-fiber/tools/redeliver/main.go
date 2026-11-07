package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/hibiken/asynq"
	"latihan-fiber/app/model"
	"latihan-fiber/config"
	"latihan-fiber/database"
)

func main() {
	outboxID := flag.Int64("outbox", 0, "ID outbox yang akan dikirim ulang")
	copies := flag.Int("copies", 1, "Jumlah salinan task yang dikirim ke Redis")
	flag.Parse()

	if *outboxID <= 0 {
		log.Fatal("flag -outbox harus berupa ID outbox positif")
	}
	if *copies <= 0 {
		log.Fatal("flag -copies harus minimal 1")
	}

	config.LoadEnv()
	ctx := context.Background()

	pool, err := database.NewPool(ctx)
	if err != nil {
		log.Fatalf("koneksi database: %v", err)
	}
	defer pool.Close()

	var eventType string
	var payload []byte
	err = pool.QueryRow(ctx, `SELECT event_type, payload FROM outbox WHERE id = $1`, *outboxID).Scan(&eventType, &payload)
	if err != nil {
		log.Fatalf("mengambil outbox %d: %v", *outboxID, err)
	}

	client := asynq.NewClient(config.RedisOpt())
	defer client.Close()

	job := model.NotificationJob{
		OutboxID:  *outboxID,
		EventType: eventType,
		Payload:   payload,
	}
	jobPayload, err := json.Marshal(job)
	if err != nil {
		log.Fatalf("marshal job: %v", err)
	}

	for i := 1; i <= *copies; i++ {
		// Task dikirim tanpa TaskID tunggal (atau dengan ID salinan unik) agar Redis menerima seluruh salinan
		task := asynq.NewTask(
			model.TaskSendNotification,
			jobPayload,
			asynq.Queue(model.QueueNotifications),
			asynq.MaxRetry(5),
			asynq.Timeout(30*time.Second),
			asynq.Retention(24*time.Hour),
		)
		info, err := client.EnqueueContext(ctx, task)
		if err != nil {
			log.Fatalf("gagal enqueue salinan %d: %v", i, err)
		}
		fmt.Printf("Salinan %d terkirim ke Redis: task_id=%s queue=%s\n", i, info.ID, info.Queue)
	}

	fmt.Printf("Berhasil mengirim %d salinan task untuk outbox ID %d ke Redis.\n", *copies, *outboxID)
}

