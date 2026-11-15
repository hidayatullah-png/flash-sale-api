package route

import (
	"github.com/gofiber/fiber/v2"

	"flash-sale-api/app/service"
)

// RegisterProductRoutes mendaftarkan endpoint CRUD produk (PostgreSQL).
func RegisterProductRoutes(api fiber.Router, svc *service.ProductService) {
	products := api.Group("/products")
	products.Post("/", svc.CreateProduct)
	products.Get("/", svc.GetProducts)
	products.Get("/:id", svc.GetProduct)
	products.Put("/:id", svc.UpdateProduct)
	products.Delete("/:id", svc.DeleteProduct)
}
