package worker

import (
	"context"
	"log"
	"time"

	"flash-sale-api/app/repository"

	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RunOutboxPublisher menjalankan background worker berkala untuk membaca outbox dan mem-publish event
func RunOutboxPublisher(ctx context.Context, dbPool *pgxpool.Pool, asynqClient *asynq.Client) {
	outboxRepo := repository.NewOutboxRepository(dbPool)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	log.Println("[Outbox] Publisher worker aktif...")

	for {
		select {
		case <-ctx.Done():
			log.Println("[Outbox] Publisher worker dimatikan")
			return
		case <-ticker.C:
			events, err := outboxRepo.FetchPendingEvents(ctx, 20)
			if err != nil {
				log.Printf("[Outbox] Gagal fetch event pending: %v", err)
				continue
			}

			for _, ev := range events {
				task := asynq.NewTask(ev.EventType, ev.Payload)
				if _, err := asynqClient.Enqueue(task); err != nil {
					log.Printf("[Outbox] Gagal enqueue task ID %d (%s): %v", ev.ID, ev.EventType, err)
					continue
				}

				if err := outboxRepo.MarkProcessed(ctx, ev.ID); err != nil {
					log.Printf("[Outbox] Gagal mark processed ID %d: %v", ev.ID, err)
				} else {
					log.Printf("[Outbox] Event ID %d (%s) berhasil dipublish ke Asynq", ev.ID, ev.EventType)
				}
			}
		}
	}
}
