package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"flash-sale-api/app/model"
)

// ErrProductNotFound sudah dideklarasikan di product_repository.go (package yang sama),
// jadi di sini hanya error khusus slot.
var (
	ErrSlotNotFound = errors.New("slot flash sale tidak ditemukan")
	ErrSlotInUse    = errors.New("slot flash sale tidak dapat dihapus karena sudah memiliki pesanan")
)

type FlashSaleSlotRepository struct {
	db *pgxpool.Pool
}

func NewFlashSaleSlotRepository(db *pgxpool.Pool) *FlashSaleSlotRepository {
	return &FlashSaleSlotRepository{db: db}
}

func (r *FlashSaleSlotRepository) Create(ctx context.Context, req model.CreateFlashSaleSlotRequest) (*model.FlashSaleSlot, error) {
	// 1. Pastikan produk ada di tabel products dan ambil nama + harga normalnya
	var productName string
	var normalPrice float64
	err := r.db.QueryRow(ctx,
		"SELECT name, price FROM products WHERE id = $1",
		req.ProductID,
	).Scan(&productName, &normalPrice)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}

	// 2. Simpan slot flash sale baru
	slot := &model.FlashSaleSlot{
		ProductID:   req.ProductID,
		ProductName: productName,
		NormalPrice: normalPrice,
		FlashPrice:  req.FlashPrice,
		Quota:       req.Quota,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
	}

	query := `
		INSERT INTO flash_sale_slots (product_id, flash_price, quota, start_time, end_time)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`
	err = r.db.QueryRow(ctx, query,
		req.ProductID, req.FlashPrice, req.Quota, req.StartTime, req.EndTime,
	).Scan(&slot.ID, &slot.CreatedAt)
	if err != nil {
		return nil, err
	}

	slot.Status = slot.ComputeStatus(time.Now())
	return slot, nil
}

func (r *FlashSaleSlotRepository) FindAll(ctx context.Context, activeOnly bool) ([]model.FlashSaleSlot, error) {
	query := `
		SELECT s.id, s.product_id, p.name, p.price, s.flash_price, s.quota, s.start_time, s.end_time, s.created_at
		FROM flash_sale_slots s
		JOIN products p ON p.id = s.product_id
	`
	if activeOnly {
		query += ` WHERE NOW() BETWEEN s.start_time AND s.end_time AND s.quota > 0`
	}
	query += ` ORDER BY s.start_time ASC, s.id DESC`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	slots := make([]model.FlashSaleSlot, 0)
	now := time.Now()
	for rows.Next() {
		var s model.FlashSaleSlot
		if err := rows.Scan(
			&s.ID, &s.ProductID, &s.ProductName, &s.NormalPrice,
			&s.FlashPrice, &s.Quota, &s.StartTime, &s.EndTime, &s.CreatedAt,
		); err != nil {
			return nil, err
		}
		s.Status = s.ComputeStatus(now)
		slots = append(slots, s)
	}

	return slots, rows.Err()
}

func (r *FlashSaleSlotRepository) FindByID(ctx context.Context, id int) (*model.FlashSaleSlot, error) {
	query := `
		SELECT s.id, s.product_id, p.name, p.price, s.flash_price, s.quota, s.start_time, s.end_time, s.created_at
		FROM flash_sale_slots s
		JOIN products p ON p.id = s.product_id
		WHERE s.id = $1
	`
	var s model.FlashSaleSlot
	err := r.db.QueryRow(ctx, query, id).Scan(
		&s.ID, &s.ProductID, &s.ProductName, &s.NormalPrice,
		&s.FlashPrice, &s.Quota, &s.StartTime, &s.EndTime, &s.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSlotNotFound
		}
		return nil, err
	}

	s.Status = s.ComputeStatus(time.Now())
	return &s, nil
}

func (r *FlashSaleSlotRepository) Update(ctx context.Context, id int, req model.UpdateFlashSaleSlotRequest) (*model.FlashSaleSlot, error) {
	query := `
		UPDATE flash_sale_slots
		SET flash_price = $1, quota = $2, start_time = $3, end_time = $4
		WHERE id = $5
	`
	cmdTag, err := r.db.Exec(ctx, query, req.FlashPrice, req.Quota, req.StartTime, req.EndTime, id)
	if err != nil {
		return nil, err
	}
	if cmdTag.RowsAffected() == 0 {
		return nil, ErrSlotNotFound
	}

	return r.FindByID(ctx, id)
}

func (r *FlashSaleSlotRepository) Delete(ctx context.Context, id int) error {
	cmdTag, err := r.db.Exec(ctx, "DELETE FROM flash_sale_slots WHERE id = $1", id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" { // foreign_key_violation
			return ErrSlotInUse
		}
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return ErrSlotNotFound
	}
	return nil
}
