package worker

import (
	"encoding/json"
	"time"

	"github.com/hibiken/asynq"
)

const (
	TypeOrderProcess = "order:process"
	TypeOrderCancel  = "order:cancel"
)

type OrderProcessPayload struct {
	UserID    int `json:"user_id"`
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

type OrderCancelPayload struct {
	OrderID int `json:"order_id"`
}

// NewOrderProcessTask membuat task untuk memproses order flash sale.
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

// NewOrderCancelTask membuat task yang dijadwalkan untuk membatalkan order
// secara otomatis setelah delay tertentu (misal 15 menit).
// Jika user belum bayar dalam waktu tersebut, order dibatalkan dan stok dikembalikan.
func NewOrderCancelTask(orderID int, delay time.Duration) (*asynq.Task, error) {
	payload, err := json.Marshal(OrderCancelPayload{OrderID: orderID})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeOrderCancel, payload, asynq.ProcessIn(delay)), nil
}
