package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"
	"github.com/joho/godotenv"

	// Import konfigurasi dan struktur Clean Architecture-mu
	"flash-sale-api/app/repository"
	"flash-sale-api/app/service"
	"flash-sale-api/config"
	_ "flash-sale-api/docs" // dihasilkan oleh perintah: swag init
	"flash-sale-api/route"
)

// @title           Flash Sale API
// @version         1.0
// @description     API sistem flash sale (PostgreSQL + MongoDB).
// @host            localhost:3000
// @BasePath        /api
func main() {
	// 1. Muat file .env
	err := godotenv.Load()
	if err != nil {
		log.Println("Peringatan: File .env tidak ditemukan, menggunakan environment system")
	}

	// 2. Inisiasi Koneksi Database
	dbPool := config.ConnectPostgres()
	defer dbPool.Close()

	// [Cia] Koneksi MongoDB untuk modul Review
	mongoClient, mongoDB := config.ConnectMongo()
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = mongoClient.Disconnect(ctx)
	}()

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

	// [Cia] Swagger UI -> http://localhost:3000/swagger/
	app.Get("/swagger/*", swagger.HandlerDefault)

	// ==========================================
	// BAGIAN DEPENDENCY INJECTION (PENGGANTI TODO)
	// ==========================================

	// A. Suntikkan database ke dalam Repository
	orderRepo := repository.NewOrderRepository(dbPool)
	productRepo := repository.NewProductRepository(dbPool)    // [Cia]
	reviewRepo := repository.NewReviewRepository(mongoDB)     // [Cia]
	slotRepo := repository.NewFlashSaleSlotRepository(dbPool) // [Cia]

	// [Cia] Buat index MongoDB untuk koleksi reviews
	idxCtx, idxCancel := context.WithTimeout(context.Background(), 10*time.Second)
	if err := reviewRepo.EnsureIndexes(idxCtx); err != nil {
		log.Printf("Peringatan: gagal membuat index MongoDB: %v", err)
	}
	idxCancel()

	// B. Suntikkan Repository ke dalam Service
	orderService := service.NewOrderService(orderRepo)
	productService := service.NewProductService(productRepo)           // [Cia]
	reviewService := service.NewReviewService(reviewRepo, productRepo) // [Cia]
	slotService := service.NewFlashSaleSlotService(slotRepo)           // [Cia]

	// C. Daftarkan Service ke dalam Router Fiber
	route.SetupRoutes(app, orderService)

	// [Cia] Route Product, Review & Flash Sale Slot
	api := app.Group("/api")
	route.RegisterProductRoutes(api, productService)
	route.RegisterReviewRoutes(api, reviewService)
	route.RegisterFlashSaleSlotRoutes(api, slotService)

	// ==========================================

	// 4. Jalankan Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("Server berjalan di http://localhost:%s", port)
	log.Fatal(app.Listen(":" + port))
}