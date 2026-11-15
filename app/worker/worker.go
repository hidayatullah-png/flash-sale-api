package worker

import (
	"context"
	"encoding/json"
	"log"

	"github.com/hibiken/asynq"
)

func HandleOrderProcessTask(
	ctx context.Context,
	task *asynq.Task,
) error {
	var payload OrderProcessPayload

	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return err
	}

	log.Printf(
		"Processing checkout: UserID=%d, ProductID=%d, Quantity=%d",
		payload.UserID,
		payload.ProductID,
		payload.Quantity,
	)

	// TODO: Integrasikan dengan OrderRepository
	// untuk validasi stok dan pembuatan order.

	return nil
}
