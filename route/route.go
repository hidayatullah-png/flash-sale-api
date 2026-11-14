package route

import (
	"github.com/gofiber/fiber/v2"

	"flash-sale-api/app/service"
	"flash-sale-api/helper"
	"flash-sale-api/middleware"
)

func SetupRoutes(
	app *fiber.App,
	orderService *service.OrderService,
	authService *service.AuthService,
	jwtManager *helper.JWTManager,
) {
	api := app.Group("/api/v1")

	// Endpoint Auth 
	auth := api.Group("/auth")
	auth.Post("/register", authService.Register)
	auth.Post("/login", middleware.LoginRateLimiter(), authService.Login)
	auth.Get("/me", middleware.RequireAuth(jwtManager), authService.Me)

	// Endpoint Transaksi Flash Sale 
	api.Post("/checkout", middleware.RequireAuth(jwtManager), orderService.Checkout)
}