package route

import (
	"flash-sale-api/app/service"

	"github.com/gofiber/fiber/v2"
)

// SetupRoutes mendaftarkan semua endpoint API
func SetupRoutes(app *fiber.App, orderService *service.OrderService) {
	api := app.Group("/api/v1")

	// Endpoint Transaksi Flash Sale
	api.Post("/checkout", orderService.Checkout)
}