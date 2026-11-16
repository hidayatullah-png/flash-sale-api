package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
)

// workerDB adalah koneksi database yang dipakai oleh handler worker.
// Di-set saat RunWorker dipanggil via SetWorkerDB.
var workerDB *pgxpool.Pool

// SetWorkerDB menyimpan koneksi database agar bisa dipakai handler.
func SetWorkerDB(db *pgxpool.Pool) {
	workerDB = db
}

// HandleOrderProcessTask memproses task checkout flash sale.
func HandleOrderProcessTask(
	ctx context.Context,
	task *asynq.Task,
) error {
	var payload OrderProcessPayload

	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return err
	}

	log.Printf(
		"[Worker] Memproses checkout: UserID=%d, ProductID=%d, Quantity=%d",
		payload.UserID,
		payload.ProductID,
		payload.Quantity,
	)

	// TODO: Integrasikan dengan OrderRepository
	// untuk validasi stok dan pembuatan order.

	return nil
}

// HandleOrderCancelTask membatalkan order yang belum dibayar setelah 15 menit.
// Task ini dijadwalkan otomatis saat order dibuat.
func HandleOrderCancelTask(
	ctx context.Context,
	task *asynq.Task,
) error {
	var payload OrderCancelPayload

	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("gagal unmarshal payload cancel: %w", err)
	}

	log.Printf("[Worker] Memeriksa order %d untuk auto-cancel...", payload.OrderID)

	if workerDB == nil {
		return fmt.Errorf("koneksi database belum diinisialisasi di worker")
	}

	// 1. Mulai transaksi
	tx, err := workerDB.Begin(ctx)
	if err != nil {
		return fmt.Errorf("gagal memulai transaksi: %w", err)
	}
	defer tx.Rollback(ctx)

	// 2. Cek status order (hanya cancel jika masih 'pending')
	var status string
	var productID, quantity int
	err = tx.QueryRow(ctx,
		"SELECT status, product_id, quantity FROM orders WHERE id = $1 FOR UPDATE",
		payload.OrderID,
	).Scan(&status, &productID, &quantity)
	if err != nil {
		return fmt.Errorf("order %d tidak ditemukan: %w", payload.OrderID, err)
	}

	// Jika sudah dibayar atau sudah dibatalkan, skip
	if status != "pending" {
		log.Printf("[Worker] Order %d sudah berstatus '%s', skip cancel", payload.OrderID, status)
		return nil
	}

	// 3. Update status order menjadi 'cancelled'
	_, err = tx.Exec(ctx,
		"UPDATE orders SET status = 'cancelled' WHERE id = $1",
		payload.OrderID,
	)
	if err != nil {
		return fmt.Errorf("gagal mengubah status order: %w", err)
	}

	// 4. Kembalikan stok produk
	_, err = tx.Exec(ctx,
		"UPDATE products SET stock = stock + $1 WHERE id = $2",
		quantity, productID,
	)
	if err != nil {
		return fmt.Errorf("gagal mengembalikan stok: %w", err)
	}

	// 5. Commit transaksi
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("gagal commit transaksi cancel: %w", err)
	}

	log.Printf("[Worker] Order %d berhasil di-cancel. Stok produk %d dikembalikan +%d",
		payload.OrderID, productID, quantity)

	return nil
}
