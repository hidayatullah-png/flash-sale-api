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

// CheckoutFlashSale adalah fungsi inti untuk menangani rebutan barang flash sale.
// Mengembalikan orderID yang baru dibuat agar bisa dijadwalkan auto-cancel.
func (r *OrderRepository) CheckoutFlashSale(ctx context.Context, userID, slotID, quantity int) (int, error) {
	// 1. Mulai Transaksi Database (ACID)
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	// 2. KUNCI BARIS SLOT (Row-Level Locking pada flash_sale_slots)
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
	err = tx.QueryRow(ctx, queryLock, slotID).Scan(&productID, &flashPrice, &currentQuota, &startTime, &endTime)
	if err != nil {
		return 0, errors.New("slot flash sale tidak ditemukan")
	}

	// 3. Validasi Batas Waktu Flash Sale
	now := time.Now()
	if now.Before(startTime) {
		return 0, errors.New("mohon maaf, sesi flash sale belum dimulai")
	}
	if now.After(endTime) {
		return 0, errors.New("mohon maaf, sesi flash sale sudah berakhir")
	}

	// 4. Validasi Kuota Slot
	if currentQuota < quantity {
		log.Printf("User %d gagal beli di slot %d, sisa kuota %d, minta %d", userID, slotID, currentQuota, quantity)
		return 0, errors.New("mohon maaf, kuota slot flash sale sudah habis")
	}

	// 5. Kurangi Quota di tabel flash_sale_slots
	_, err = tx.Exec(ctx, "UPDATE flash_sale_slots SET quota = quota - $1 WHERE id = $2", quantity, slotID)
	if err != nil {
		return 0, err
	}

	// 6. Catat Pesanan ke Tabel Orders menggunakan flash_price & flash_sale_slot_id
	// Gunakan RETURNING id untuk mendapatkan ID order yang baru dibuat
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

	// 7. Commit Transaksi & Lepas Kunci
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
	).Scan(&order.ID, &order.UserID, &order.ProductID, &order.Quantity, &order.TotalPrice, &order.Status, &order.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &order, nil
}
