package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Review adalah ulasan pembeli yang disimpan di MongoDB.
type Review struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty" swaggertype:"string" example:"665f1c2e8f1b2a0012ab34cd"`
	ProductID int                `json:"product_id" bson:"product_id" example:"1"`
	UserID    int                `json:"user_id" bson:"user_id" example:"7"`
	Rating    int                `json:"rating" bson:"rating" example:"5"`
	Comment   string             `json:"comment" bson:"comment" example:"Barang bagus, pengiriman cepat!"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
}

// ReviewRequest adalah body untuk membuat ulasan.
type ReviewRequest struct {
	UserID  int    `json:"user_id" example:"7"`
	Rating  int    `json:"rating" example:"5"` // 1 - 5
	Comment string `json:"comment" example:"Barang bagus, pengiriman cepat!"`
}

// RatingSummary adalah ringkasan rating satu produk.
type RatingSummary struct {
	ProductID     int     `json:"product_id" bson:"-" example:"1"`
	AverageRating float64 `json:"average_rating" bson:"average" example:"4.5"`
	TotalReviews  int     `json:"total_reviews" bson:"count" example:"12"`
}
