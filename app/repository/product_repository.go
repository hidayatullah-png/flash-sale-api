package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"flash-sale-api/app/model"
)

var (
	ErrProductNotFound = errors.New("produk tidak ditemukan")
	ErrProductInUse    = errors.New("produk masih dipakai oleh order, tidak bisa dihapus")
)

// ProductRepository mengakses tabel products di PostgreSQL.
type ProductRepository struct {
	db *pgxpool.Pool
}

func NewProductRepository(db *pgxpool.Pool) *ProductRepository {
	return &ProductRepository{db: db}
}

const productColumns = "id, name, description, price, stock, version, created_at"

func scanProduct(row pgx.Row) (*model.Product, error) {
	var p model.Product
	err := row.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Stock, &p.Version, &p.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}
	return &p, nil
}

// Create menambah produk baru.
func (r *ProductRepository) Create(ctx context.Context, req model.ProductRequest) (*model.Product, error) {
	row := r.db.QueryRow(ctx,
		`INSERT INTO products (name, description, price, stock)
		 VALUES ($1, $2, $3, $4)
		 RETURNING `+productColumns,
		req.Name, req.Description, req.Price, req.Stock)
	return scanProduct(row)
}

// FindAll mengambil daftar produk dengan paginasi.
func (r *ProductRepository) FindAll(ctx context.Context, limit, offset int) ([]model.Product, error) {
	rows, err := r.db.Query(ctx,
		`SELECT `+productColumns+` FROM products ORDER BY id LIMIT $1 OFFSET $2`,
		limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := make([]model.Product, 0)
	for rows.Next() {
		var p model.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Stock, &p.Version, &p.CreatedAt); err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, rows.Err()
}

// FindByID mengambil satu produk.
func (r *ProductRepository) FindByID(ctx context.Context, id int) (*model.Product, error) {
	row := r.db.QueryRow(ctx, `SELECT `+productColumns+` FROM products WHERE id = $1`, id)
	return scanProduct(row)
}

// Update mengubah data produk dan menaikkan version (+1)
// supaya perubahan terdeteksi oleh logika anti race-condition di modul order.
func (r *ProductRepository) Update(ctx context.Context, id int, req model.ProductRequest) (*model.Product, error) {
	row := r.db.QueryRow(ctx,
		`UPDATE products
		 SET name = $1, description = $2, price = $3, stock = $4, version = version + 1
		 WHERE id = $5
		 RETURNING `+productColumns,
		req.Name, req.Description, req.Price, req.Stock, id)
	return scanProduct(row)
}

// Delete menghapus produk.
func (r *ProductRepository) Delete(ctx context.Context, id int) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM products WHERE id = $1`, id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" { // foreign_key_violation
			return ErrProductInUse
		}
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrProductNotFound
	}
	return nil
}
