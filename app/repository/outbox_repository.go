package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OutboxEvent struct {
	ID          int64           `json:"id"`
	AggregateID string          `json:"aggregate_id"`
	EventType   string          `json:"event_type"`
	Payload     json.RawMessage `json:"payload"`
	Status      string          `json:"status"`
	CreatedAt   time.Time       `json:"created_at"`
	ProcessedAt *time.Time      `json:"processed_at,omitempty"`
}

type OutboxRepository struct {
	db *pgxpool.Pool
}

func NewOutboxRepository(db *pgxpool.Pool) *OutboxRepository {
	return &OutboxRepository{db: db}
}

// InsertEventTx menyimpan event outbox dalam transaksi yang sama dengan operasi bisnis
func (r *OutboxRepository) InsertEventTx(ctx context.Context, tx pgx.Tx, aggregateID string, eventType string, payload any) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO outbox_events (aggregate_id, event_type, payload, status, created_at)
		VALUES ($1, $2, $3, 'PENDING', NOW())
	`
	_, err = tx.Exec(ctx, query, aggregateID, eventType, payloadBytes)
	return err
}

// FetchPendingEvents mengambil batch event outbox yang berstatus PENDING
func (r *OutboxRepository) FetchPendingEvents(ctx context.Context, limit int) ([]OutboxEvent, error) {
	query := `
		SELECT id, aggregate_id, event_type, payload, status, created_at, processed_at
		FROM outbox_events
		WHERE status = 'PENDING'
		ORDER BY created_at ASC
		LIMIT $1
	`
	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []OutboxEvent
	for rows.Next() {
		var ev OutboxEvent
		if err := rows.Scan(&ev.ID, &ev.AggregateID, &ev.EventType, &ev.Payload, &ev.Status, &ev.CreatedAt, &ev.ProcessedAt); err != nil {
			return nil, err
		}
		events = append(events, ev)
	}

	return events, rows.Err()
}

// MarkProcessed menandai event telah berhasil dikirim / diproses
func (r *OutboxRepository) MarkProcessed(ctx context.Context, id int64) error {
	query := `
		UPDATE outbox_events
		SET status = 'PROCESSED', processed_at = NOW()
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query, id)
	return err
}
