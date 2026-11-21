package route

import (
	"github.com/gofiber/fiber/v2"

	"flash-sale-api/app/service"
)

// RegisterFlashSaleSlotRoutes mendaftarkan endpoint jadwal flash sale.
// POST, PUT, DELETE nantinya khusus Admin (pasang middleware JWT dari modul Auth).
func RegisterFlashSaleSlotRoutes(api fiber.Router, svc *service.FlashSaleSlotService) {
	slots := api.Group("/flash-sales")
	slots.Post("/", svc.CreateSlot)
	slots.Get("/", svc.GetSlots)
	slots.Get("/:id", svc.GetSlot)
	slots.Put("/:id", svc.UpdateSlot)
	slots.Delete("/:id", svc.DeleteSlot)
}
