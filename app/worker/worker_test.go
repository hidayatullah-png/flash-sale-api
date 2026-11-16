package worker_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"flash-sale-api/app/worker"

	"github.com/hibiken/asynq"
)

// TestOrderCancelTaskPayload memverifikasi pembuatan task auto-cancel Asynq
func TestOrderCancelTaskPayload(t *testing.T) {
	orderID := 99
	delay := 15 * time.Minute

	task, err := worker.NewOrderCancelTask(orderID, delay)
	if err != nil {
		t.Fatalf("Gagal membuat cancel task: %v", err)
	}

	if task.Type() != worker.TypeOrderCancel {
		t.Errorf("Expected task type %s, got %s", worker.TypeOrderCancel, task.Type())
	}

	var payload worker.OrderCancelPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		t.Fatalf("Gagal unmarshal payload: %v", err)
	}

	if payload.OrderID != orderID {
		t.Errorf("Expected OrderID %d, got %d", orderID, payload.OrderID)
	}
}

// TestHandleOrderCancelWithoutDB memverifikasi handler auto-cancel aman saat DB pool belum disetel
func TestHandleOrderCancelWithoutDB(t *testing.T) {
	task, err := worker.NewOrderCancelTask(101, 1*time.Second)
	if err != nil {
		t.Fatalf("Gagal membuat task: %v", err)
	}

	// dbPool nil secara default jika belum di SetWorkerDB
	worker.SetWorkerDB(nil)

	err = worker.HandleOrderCancelTask(context.Background(), task)
	if err == nil {
		t.Errorf("Harusnya return error saat worker DB nil, tetapi tidak ada error")
	}
}

// TestAsynqServerConfig memverifikasi server queues dan mux handlers
func TestAsynqServerConfig(t *testing.T) {
	server := asynq.NewServer(
		asynq.RedisClientOpt{Addr: "localhost:6379"},
		asynq.Config{
			Concurrency: 5,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
		},
	)
	if server == nil {
		t.Fatal("Gagal membuat Asynq server instance")
	}

	mux := worker.RegisterHandlers(server)
	if mux == nil {
		t.Fatal("Gagal membuat register handlers mux")
	}
}
