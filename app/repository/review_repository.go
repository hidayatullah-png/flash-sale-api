package repository

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"flash-sale-api/app/model"
)

var (
	ErrReviewNotFound  = errors.New("ulasan tidak ditemukan")
	ErrInvalidReviewID = errors.New("id ulasan tidak valid")
)

// ReviewRepository mengakses koleksi "reviews" di MongoDB (bukan PostgreSQL).
type ReviewRepository struct {
	col *mongo.Collection
}

func NewReviewRepository(db *mongo.Database) *ReviewRepository {
	return &ReviewRepository{col: db.Collection("reviews")}
}

// EnsureIndexes membuat index agar query per produk cepat. Panggil sekali saat startup.
func (r *ReviewRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.col.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "product_id", Value: 1},
			{Key: "created_at", Value: -1},
		},
	})
	return err
}

// Create menyimpan ulasan baru.
func (r *ReviewRepository) Create(ctx context.Context, review *model.Review) error {
	review.CreatedAt = time.Now().UTC()
	res, err := r.col.InsertOne(ctx, review)
	if err != nil {
		return err
	}
	review.ID = res.InsertedID.(primitive.ObjectID)
	return nil
}

// FindByProductID mengambil ulasan satu produk, terbaru dulu.
func (r *ReviewRepository) FindByProductID(ctx context.Context, productID int, limit, skip int64) ([]model.Review, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(limit).
		SetSkip(skip)

	cur, err := r.col.Find(ctx, bson.D{{Key: "product_id", Value: productID}}, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	reviews := make([]model.Review, 0)
	if err := cur.All(ctx, &reviews); err != nil {
		return nil, err
	}
	return reviews, nil
}

// GetRatingSummary menghitung rata-rata rating & jumlah ulasan dengan aggregation pipeline.
func (r *ReviewRepository) GetRatingSummary(ctx context.Context, productID int) (*model.RatingSummary, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.D{{Key: "product_id", Value: productID}}}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: nil},
			{Key: "average", Value: bson.D{{Key: "$avg", Value: "$rating"}}},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
	}

	cur, err := r.col.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	summary := &model.RatingSummary{ProductID: productID}
	if cur.Next(ctx) {
		if err := cur.Decode(summary); err != nil {
			return nil, err
		}
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	return summary, nil
}

// Delete menghapus ulasan berdasarkan ObjectID (hex string).
func (r *ReviewRepository) Delete(ctx context.Context, id string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return ErrInvalidReviewID
	}
	res, err := r.col.DeleteOne(ctx, bson.D{{Key: "_id", Value: oid}})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return ErrReviewNotFound
	}
	return nil
}
