package model

import "time"

type Product struct {
    ID          int       `json:"id"`
    Name        string    `json:"name"`
    Description string    `json:"description"`
    Price       float64   `json:"price"`
    Stock       int       `json:"stock"`
    Version     int       `json:"version"` // Penting untuk mencegah race condition
    CreatedAt   time.Time `json:"created_at"`
}