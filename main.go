package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"

	// Sesuaikan nama modul ini dengan nama di file go.mod milikmu
	"flash-sale-api/config" 
)

func main() {
	// 1. Muat file .env
	err := godotenv.Load()
	if err != nil {
		log.Println("Peringatan: File .env tidak ditemukan, menggunakan environment system")
	}

	// 2. Inisiasi Koneksi Database (BARU DITAMBAHKAN)
	dbPool := config.ConnectPostgres()
	defer dbPool.Close() // Pastikan koneksi ditutup saat server mati

	// 3. Inisiasi Fiber
	app := fiber.New(fiber.Config{
		AppName: "Flash Sale API v1.0",
	})

	// Middleware bawaan Fiber
	app.Use(logger.New())
	app.Use(recover.New())

	// Route Test Sederhana
	app.Get("/ping", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Boilerplate Flash Sale API Berhasil Berjalan!",
		})
	})

	// TODO: Nanti Route dan Dependency Injection di-setup di sini
	// Dan kita akan melempar dbPool ke dalam service
	// route.SetupRoutes(app, dbPool)

	// 4. Jalankan Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("Server berjalan di http://localhost:%s", port)
	log.Fatal(app.Listen(":" + port))
}