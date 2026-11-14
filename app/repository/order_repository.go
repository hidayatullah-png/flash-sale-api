package repository

import (
	"context"
	"errors"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepository struct {
	db *pgxpool.Pool
}

func NewOrderRepository(db *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{db: db}
}

// CheckoutFlashSale adalah fungsi inti untuk menangani rebutan barang
func (r *OrderRepository) CheckoutFlashSale(ctx context.Context, userID, productID, quantity int) error {
	// 1. Mulai Transaksi Database (ACID)
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
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
		return errors.New("produk tidak ditemukan")
	}

	// 3. Validasi Stok
	if currentStock < quantity {
		log.Printf("User %d gagal beli, stok sisa %d, minta %d", userID, currentStock, quantity)
		return errors.New("mohon maaf, stok barang flash sale sudah habis")
	}

	// 4. Potong Stok Produk
	_, err = tx.Exec(ctx, "UPDATE products SET stock = stock - $1 WHERE id = $2", quantity, productID)
	if err != nil {
		return err
	}

	// 5. Catat Pesanan ke Tabel Orders
	totalPrice := price * float64(quantity)
	_, err = tx.Exec(ctx,
		"INSERT INTO orders (user_id, product_id, quantity, total_price, status) VALUES ($1, $2, $3, $4, 'pending')",
		userID, productID, quantity, totalPrice,
	)
	if err != nil {
		return err
	}

	// 6. Simpan Semua Perubahan (Commit)
	return tx.Commit(ctx)
}
