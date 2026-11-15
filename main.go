package main

import (
	"context"
	"log"
	"os"
	"time"

	_ "flash-sale-api/docs"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"

	"flash-sale-api/app/repository"
	"flash-sale-api/app/service"
	"flash-sale-api/config"
	"flash-sale-api/helper"
	"flash-sale-api/route"
)

func main() {
	// 1. Muat file .env
	err := godotenv.Load()
	if err != nil {
		log.Println("Peringatan: File .env tidak ditemukan, menggunakan environment system")
	}

	// Inisialisasi Redis
	redisClient, err := config.NewRedisClient(context.Background())
	if err != nil {
		log.Fatalf("Gagal menginisialisasi Redis: %v", err)
	}
	defer redisClient.Close()

	log.Println("Berhasil terhubung ke Redis!")

	// 2. Inisiasi Koneksi Database (GABUNGAN SUKMA & CIA)
	dbPool := config.ConnectPostgres()
	defer dbPool.Close()

	// Sesuaikan dengan fungsi buatan Cia yang mengembalikan 2 nilai
	mongoClient, mongoDB := config.ConnectMongo()
	defer mongoClient.Disconnect(context.Background())

	// 3. Inisiasi Fiber
	app := fiber.New(fiber.Config{
		AppName: "Flash Sale API v1.0",
	})
	app.Use(logger.New())

	// 4. Setup JWT Manager (PUNYA FAZA)
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "rahasia-flash-sale-super-aman-123"
	}
	jwtManager := helper.NewJWTManager(jwtSecret, "flash-sale-api", 24*time.Hour)

	// 5. Masukkan Database ke Repository (GABUNGAN BERTIGA)
	orderRepo := repository.NewOrderRepository(dbPool)     // Sukma
	userRepo := repository.NewUserRepository(dbPool)       // Faza
	productRepo := repository.NewProductRepository(dbPool) // Cia
	reviewRepo := repository.NewReviewRepository(mongoDB)  // Cia

	// 6. Masukkan Repository ke Service (GABUNGAN BERTIGA)
	orderService := service.NewOrderService(orderRepo)
	authService := service.NewAuthService(userRepo, jwtManager)
	productService := service.NewProductService(productRepo)

	// Tambahkan productRepo di sini
	reviewService := service.NewReviewService(reviewRepo, productRepo)
	// 7. Daftarkan Service ke Router (DISESUAIKAN)
	// Pastikan parameter fungsi SetupRoutes di route.go juga diubah untuk menerima semuanya
	route.SetupRoutes(app, orderService, authService, productService, reviewService, jwtManager)

	// 8. Jalankan Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("Server berjalan di http://localhost:%s", port)
	log.Fatal(app.Listen(":" + port))
}
