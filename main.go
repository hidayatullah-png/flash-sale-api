package main

import (
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"

	// Import konfigurasi dan struktur Clean Architecture-mu
	"flash-sale-api/app/repository"
	"flash-sale-api/app/service"
	"flash-sale-api/config"
	"flash-sale-api/helper"
	"flash-sale-api/route"
)

func main() {
	//Muat file .env
	err := godotenv.Load()
	if err != nil {
		log.Println("Peringatan: File .env tidak ditemukan, menggunakan environment system")
	}

	//Inisiasi Koneksi Database
	dbPool := config.ConnectPostgres()
	defer dbPool.Close()

	//Inisiasi Fiber
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

	// Setup JWT Manager (mengambil secret dari .env atau fallback default)
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "rahasia-flash-sale-super-aman-123"
	}
	jwtManager := helper.NewJWTManager(jwtSecret, "flash-sale-api", 24*time.Hour)

	//memasukkan database ke dalam Repository
	orderRepo := repository.NewOrderRepository(dbPool)
	userRepo := repository.NewUserRepository(dbPool)

	// memasukkan Repository ke dalam Service
	orderService := service.NewOrderService(orderRepo)
	authService := service.NewAuthService(userRepo, jwtManager)

	//Daftarkan Service ke dalam Router Fiber
	route.SetupRoutes(app, orderService, authService, jwtManager)

	//Jalankan Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("Server berjalan di http://localhost:%s", port)
	log.Fatal(app.Listen(":" + port))
}