package model

import "time"

// Product adalah barang flash sale yang disimpan di PostgreSQL.
type Product struct {
	ID          int       `json:"id" example:"1"`
	Name        string    `json:"name" example:"Sepatu Lari Limited"`
	Description string    `json:"description" example:"Edisi terbatas, hanya 100 pasang"`
	Price       float64   `json:"price" example:"499000"`
	Stock       int       `json:"stock" example:"100"`
	Version     int       `json:"version" example:"1"` // Penting untuk mencegah race condition
	CreatedAt   time.Time `json:"created_at"`
}

// ProductRequest adalah body untuk membuat / mengedit produk.
type ProductRequest struct {
	Name        string  `json:"name" example:"Sepatu Lari Limited"`
	Description string  `json:"description" example:"Edisi terbatas, hanya 100 pasang"`
	Price       float64 `json:"price" example:"499000"`
	Stock       int     `json:"stock" example:"100"`
}
