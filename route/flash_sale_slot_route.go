package route

import (
	"github.com/gofiber/fiber/v2"

	"flash-sale-api/app/service"
	"flash-sale-api/helper"
	"flash-sale-api/middleware"
)

// RegisterFlashSaleSlotRoutes mendaftarkan endpoint jadwal flash sale.
// GET bebas diakses, sedangkan POST, PUT, DELETE butuh token JWT.
func RegisterFlashSaleSlotRoutes(api fiber.Router, svc *service.FlashSaleSlotService, jwtManager *helper.JWTManager) {
	slots := api.Group("/flash-sales")
	slots.Get("/", svc.GetSlots)
	slots.Get("/:id", svc.GetSlot)
	slots.Post("/", middleware.RequireAuth(jwtManager), svc.CreateSlot)
	slots.Put("/:id", middleware.RequireAuth(jwtManager), svc.UpdateSlot)
	slots.Delete("/:id", middleware.RequireAuth(jwtManager), svc.DeleteSlot)
}
