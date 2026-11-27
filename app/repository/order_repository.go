
package repository

import (
	"context"
	"errors"
	"log"
	"time"

	"flash-sale-api/app/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepository struct {
	db *pgxpool.Pool
}

func NewOrderRepository(db *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{db: db}
}

// CheckoutFlashSale menangani checkout dan mengembalikan ID order
// yang baru dibuat agar bisa dijadwalkan auto-cancel.
func (r *OrderRepository) CheckoutFlashSale(ctx context.Context, userID, slotID, quantity int) (int, error) {
	// 1. Mulai transaksi database
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	// 2. Kunci baris slot menggunakan Row-Level Locking
	var productID int
	var flashPrice float64
	var currentQuota int
	var startTime, endTime time.Time

	queryLock := `
		SELECT product_id, flash_price, quota, start_time, end_time
		FROM flash_sale_slots
		WHERE id = $1
		FOR UPDATE
	`
	err = tx.QueryRow(ctx, queryLock, slotID).Scan(
		&productID,
		&flashPrice,
		&currentQuota,
		&startTime,
		&endTime,
	)
	if err != nil {
		return 0, errors.New("slot flash sale tidak ditemukan")
	}

	// 3. Validasi batas waktu flash sale
	now := time.Now()
	if now.Before(startTime) {
		return 0, errors.New("mohon maaf, sesi flash sale belum dimulai")
	}
	if now.After(endTime) {
		return 0, errors.New("mohon maaf, sesi flash sale sudah berakhir")
	}

	// 4. Validasi kuota slot
	if currentQuota < quantity {
		log.Printf(
			"User %d gagal beli di slot %d, sisa kuota %d, minta %d",
			userID, slotID, currentQuota, quantity,
		)
		return 0, errors.New("mohon maaf, kuota slot flash sale sudah habis")
	}

	// 5. Kurangi kuota pada tabel flash_sale_slots
	_, err = tx.Exec(
		ctx,
		"UPDATE flash_sale_slots SET quota = quota - $1 WHERE id = $2",
		quantity,
		slotID,
	)
	if err != nil {
		return 0, err
	}

	// 6. Catat pesanan dan ambil ID order
	totalPrice := flashPrice * float64(quantity)
	var orderID int

	err = tx.QueryRow(ctx,
		`INSERT INTO orders (user_id, product_id, flash_sale_slot_id, quantity, total_price, status)
		 VALUES ($1, $2, $3, $4, $5, 'pending')
		 RETURNING id`,
		userID, productID, slotID, quantity, totalPrice,
	).Scan(&orderID)
	if err != nil {
		return 0, err
	}

	// 7. Commit transaksi
	if err = tx.Commit(ctx); err != nil {
		return 0, err
	}

	return orderID, nil
}

// GetOrderByID mengambil detail order berdasarkan ID.
func (r *OrderRepository) GetOrderByID(ctx context.Context, orderID int) (*model.Order, error) {
	var order model.Order

	err := r.db.QueryRow(ctx,
		"SELECT id, user_id, product_id, quantity, total_price, status, created_at FROM orders WHERE id = $1",
		orderID,
	).Scan(
		&order.ID,
		&order.UserID,
		&order.ProductID,
		&order.Quantity,
		&order.TotalPrice,
		&order.Status,
		&order.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &order, nil
}
