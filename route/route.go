package route

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"

	"flash-sale-api/app/service"
	"flash-sale-api/helper"
	"flash-sale-api/middleware"
)

// SetupRoutes sekarang menerima SEMUA service dari kalian bertiga
func SetupRoutes(
	app *fiber.App,
	orderService *service.OrderService,
	authService *service.AuthService,
	productService *service.ProductService, // Tambahan Cia
	reviewService *service.ReviewService, // Tambahan Cia
	slotService *service.FlashSaleSlotService, // Tambahan Cia
	jwtManager *helper.JWTManager,
) {
	// Endpoint Swagger UI (Dari Cia) - Bebas diakses tanpa token
	app.Get("/swagger/*", swagger.HandlerDefault)

	api := app.Group("/api/v1")

	// -----------------------------------------------------
	// 1. Endpoint Auth & User (Dari Faza)
	// -----------------------------------------------------
	auth := api.Group("/auth")
	auth.Post("/register", authService.Register)
	auth.Post("/login", middleware.LoginRateLimiter(), authService.Login)
	auth.Get("/me", middleware.RequireAuth(jwtManager), authService.Me)

	// -----------------------------------------------------
	// 2. Endpoint Product (Dari Cia)
	// -----------------------------------------------------
	products := api.Group("/products")

	// Ubah FindAll menjadi GetProducts
	products.Get("/", productService.GetProducts)

	// Ubah FindByID menjadi GetProduct
	products.Get("/:id", productService.GetProduct)

	// Ubah Create menjadi CreateProduct
	products.Post("/", middleware.RequireAuth(jwtManager), productService.CreateProduct)

	// Ubah Update menjadi UpdateProduct
	products.Put("/:id", middleware.RequireAuth(jwtManager), productService.UpdateProduct)

	// Ubah Delete menjadi DeleteProduct
	products.Delete("/:id", middleware.RequireAuth(jwtManager), productService.DeleteProduct)

	// -----------------------------------------------------
	// 3. Endpoint Review MongoDB (Dari Cia)
	// -----------------------------------------------------
	reviews := api.Group("/reviews")

	// Ubah FindByProductID menjadi GetReviews
	reviews.Get("/product/:product_id", reviewService.GetReviews)

	// GetRatingSummary sudah benar, biarkan saja
	reviews.Get("/product/:product_id/summary", reviewService.GetRatingSummary)

	// Ubah Create menjadi CreateReview
	reviews.Post("/", middleware.RequireAuth(jwtManager), reviewService.CreateReview)

	// Ubah Delete menjadi DeleteReview
	reviews.Delete("/:id", middleware.RequireAuth(jwtManager), reviewService.DeleteReview)

	// -----------------------------------------------------
	// 4. Endpoint Transaksi Flash Sale (Dari Sukma)
	// -----------------------------------------------------
	api.Post("/checkout", middleware.RequireAuth(jwtManager), orderService.Checkout)

	// -----------------------------------------------------
	// 5. Endpoint Slot (Jadwal) Flash Sale (Dari Cia)
	// -----------------------------------------------------
	RegisterFlashSaleSlotRoutes(api, slotService, jwtManager)
}
