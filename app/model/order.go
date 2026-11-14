package model

import "time"

type Order struct {
    ID         int       `json:"id"`
    UserID     int       `json:"user_id"`
    ProductID  int       `json:"product_id"`
    Quantity   int       `json:"quantity"`
    TotalPrice float64   `json:"total_price"`
    Status     string    `json:"status"` // pending, paid, cancelled
    CreatedAt  time.Time `json:"created_at"`
}