package repository

import (
	"context"
	"errors"
	"log"

	"flash-sale-api/app/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepository struct {
	db *pgxpool.Pool
}

func NewOrderRepository(db *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{db: db}
}

// CheckoutFlashSale adalah fungsi inti untuk menangani rebutan barang.
// Mengembalikan orderID yang baru dibuat agar bisa dijadwalkan auto-cancel.
func (r *OrderRepository) CheckoutFlashSale(ctx context.Context, userID, productID, quantity int) (int, error) {
	// 1. Mulai Transaksi Database (ACID)
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return 0, err
	}
	// Pastikan transaksi dibatalkan (rollback) jika gagal di tengah jalan
	defer tx.Rollback(ctx)

	// 2. KUNCI BARIS DATA (Row-Level Locking)
	// Kita tambahkan "FOR UPDATE" agar user lain yang mau beli barang ini harus antre
	// menunggu transaksi ini selesai. Ini ilmu Modul 9 untuk mencegah Race Condition!
	var currentStock int
	var price float64
	err = tx.QueryRow(ctx, "SELECT stock, price FROM products WHERE id = $1 FOR UPDATE", productID).Scan(&currentStock, &price)
	if err != nil {
		return 0, errors.New("produk tidak ditemukan")
	}

	// 3. Validasi Stok
	if currentStock < quantity {
		log.Printf("User %d gagal beli, stok sisa %d, minta %d", userID, currentStock, quantity)
		return 0, errors.New("mohon maaf, stok barang flash sale sudah habis")
	}

	// 4. Potong Stok Produk
	_, err = tx.Exec(ctx, "UPDATE products SET stock = stock - $1 WHERE id = $2", quantity, productID)
	if err != nil {
		return 0, err
	}

	// 5. Catat Pesanan ke Tabel Orders dan ambil ID-nya
	totalPrice := price * float64(quantity)
	var orderID int
	err = tx.QueryRow(ctx,
		"INSERT INTO orders (user_id, product_id, quantity, total_price, status) VALUES ($1, $2, $3, $4, 'pending') RETURNING id",
		userID, productID, quantity, totalPrice,
	).Scan(&orderID)
	if err != nil {
		return 0, err
	}

	// 6. Simpan Semua Perubahan (Commit)
	if err = tx.Commit(ctx); err != nil {
		return 0, err
	}

	return orderID, nil
}

// CancelOrder membatalkan order yang masih pending dan mengembalikan stok.
// Dipakai oleh worker auto-cancel.
func (r *OrderRepository) CancelOrder(ctx context.Context, orderID int) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Cek dan kunci order
	var status string
	var productID, quantity int
	err = tx.QueryRow(ctx,
		"SELECT status, product_id, quantity FROM orders WHERE id = $1 FOR UPDATE",
		orderID,
	).Scan(&status, &productID, &quantity)
	if err != nil {
		return errors.New("order tidak ditemukan")
	}

	if status != "pending" {
		return nil // sudah dibayar atau sudah dibatalkan, tidak perlu apa-apa
	}

	// Update status
	_, err = tx.Exec(ctx, "UPDATE orders SET status = 'cancelled' WHERE id = $1", orderID)
	if err != nil {
		return err
	}

	// Kembalikan stok
	_, err = tx.Exec(ctx, "UPDATE products SET stock = stock + $1 WHERE id = $2", quantity, productID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
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
