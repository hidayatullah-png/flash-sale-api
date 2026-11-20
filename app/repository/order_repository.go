package repository

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func (r *OrderRepository) CheckoutFlashSale(ctx context.Context, userID, slotID, quantity int) error {
	// 1. Mulai Transaksi Database (ACID)
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
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
	err = tx.QueryRow(ctx, queryLock, slotID).Scan(&productID, &flashPrice, ¤tQuota, &startTime, &endTime)
	if err != nil {
		return errors.New("slot flash sale tidak ditemukan")
	}

	// 3. Validasi Batas Waktu Flash Sale
	now := time.Now()
	if now.Before(startTime) {
		return errors.New("mohon maaf, sesi flash sale belum dimulai")
	}
	if now.After(endTime) {
		return errors.New("mohon maaf, sesi flash sale sudah berakhir")
	}

	// 4. Validasi Kuota Slot
	if currentQuota < quantity {
		log.Printf("User %d gagal beli di slot %d, sisa kuota %d, minta %d", userID, slotID, currentQuota, quantity)
		return errors.New("mohon maaf, kuota slot flash sale sudah habis")
	}

	// 5. Kurangi Quota di tabel flash_sale_slots
	_, err = tx.Exec(ctx, "UPDATE flash_sale_slots SET quota = quota - $1 WHERE id = $2", quantity, slotID)
	if err != nil {
		return err
	}

	// 6. Catat Pesanan ke Tabel Orders menggunakan flash_price & flash_sale_slot_id
	totalPrice := flashPrice * float64(quantity)
	_, err = tx.Exec(ctx,
		`INSERT INTO orders (user_id, product_id, flash_sale_slot_id, quantity, total_price, status)
		 VALUES ($1, $2, $3, $4, $5, 'pending')`,
		userID, productID, slotID, quantity, totalPrice,
	)
	if err != nil {
		return err
	}

	// 7. Commit Transaksi & Lepas Kunci
	return tx.Commit(ctx)
}