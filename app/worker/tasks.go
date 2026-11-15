package worker

import (
	"encoding/json"

	"github.com/hibiken/asynq"
)

const (
	TypeOrderProcess = "order:process"
)

type OrderProcessPayload struct {
	UserID    int `json:"user_id"`
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

func NewOrderProcessTask(userID, productID, quantity int) (*asynq.Task, error) {
	payload, err := json.Marshal(OrderProcessPayload{
		UserID:    userID,
		ProductID: productID,
		Quantity:  quantity,
	})
	if err != nil {
		return nil, err
	}

	return asynq.NewTask(TypeOrderProcess, payload), nil
}
