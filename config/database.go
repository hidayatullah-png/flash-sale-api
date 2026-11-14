package config

import (
    "context"
    "log"
    "os"

    "github.com/jackc/pgx/v5/pgxpool"
)

func ConnectPostgres() *pgxpool.Pool {
    dbURL := os.Getenv("DB_URL")
    if dbURL == "" {
        log.Fatal("DB_URL belum disetel di file .env")
    }

    pool, err := pgxpool.New(context.Background(), dbURL)
    if err != nil {
        log.Fatalf("Gagal menyambung ke database: %v", err)
    }

    log.Println("Berhasil terhubung ke Supabase PostgreSQL!")
    return pool
}