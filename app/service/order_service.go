package service

import (
	"flash-sale-api/app/repository"

	"github.com/gofiber/fiber/v2"
)

type OrderService struct {
	repo *repository.OrderRepository
}

// Inisiasi Service dengan menyuntikkan Repository
func NewOrderService(repo *repository.OrderRepository) *OrderService {
	return &OrderService{repo: repo}
}

// Struct untuk menangkap data JSON dari user
type CheckoutRequest struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
	// Catatan: UserID sementara kita minta dari body untuk tes. 
	// Nanti setelah Faza selesai dengan fitur Auth, UserID ini akan
	// diambil secara gaib dari JWT Token (c.Locals).
	UserID    int `json:"user_id"` 
}

// Checkout adalah metode yang langsung dipanggil oleh Router Fiber
func (s *OrderService) Checkout(c *fiber.Ctx) error {
	var req CheckoutRequest

	// 1. Parsing Body Request (JSON)
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Format data tidak valid",
		})
	}

	// Validasi kuantitas dasar
	if req.Quantity <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Kuantitas pembelian minimal 1",
		})
	}

	// 2. Lempar ke Repository untuk eksekusi Row-Level Locking
	// c.Context() adalah context bawaan Fiber untuk membatalkan proses jika user tiba-tiba disconnect
	err := s.repo.CheckoutFlashSale(c.Context(), req.UserID, req.ProductID, req.Quantity)
	if err != nil {
		// Jika error karena stok habis atau barang tidak ada
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	// 3. Kembalikan Response Sukses
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Berhasil mengamankan barang Flash Sale!",
	})
}