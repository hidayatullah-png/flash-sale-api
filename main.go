package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"

	// Import konfigurasi dan struktur Clean Architecture-mu
	"flash-sale-api/app/repository"
	"flash-sale-api/app/service"
	"flash-sale-api/config"
	"flash-sale-api/route"
)

func main() {
	// 1. Muat file .env
	err := godotenv.Load()
	if err != nil {
		log.Println("Peringatan: File .env tidak ditemukan, menggunakan environment system")
	}

	// 2. Inisiasi Koneksi Database
	dbPool := config.ConnectPostgres()
	defer dbPool.Close()

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

	// ==========================================
	// BAGIAN DEPENDENCY INJECTION (PENGGANTI TODO)
	// ==========================================

	// A. Suntikkan database ke dalam Repository
	orderRepo := repository.NewOrderRepository(dbPool)

	// B. Suntikkan Repository ke dalam Service
	orderService := service.NewOrderService(orderRepo)

	// C. Daftarkan Service ke dalam Router Fiber
	route.SetupRoutes(app, orderService)

	// ==========================================

	// 4. Jalankan Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("Server berjalan di http://localhost:%s", port)
	log.Fatal(app.Listen(":" + port))
}
