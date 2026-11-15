package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"flash-sale-api/app/worker"
	"flash-sale-api/config"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Inisialisasi Database Pool untuk Worker (dibutuhkan untuk auto-cancel & restore stock)
	dbPool := config.ConnectPostgres()
	defer dbPool.Close()

	worker.SetWorkerDB(dbPool)
	log.Println("Worker database pool terhubung!")

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	log.Println("Starting Flash Sale Worker...")

	if err := worker.RunWorker(ctx); err != nil {
		log.Fatalf("Worker stopped with error: %v", err)
	}

	log.Println("Worker stopped successfully")
}
