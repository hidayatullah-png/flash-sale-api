package service

import (
	"errors"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"

	"flash-sale-api/app/model"
	"flash-sale-api/app/repository"
)

// FlashSaleSlotService berisi handler HTTP untuk jadwal (slot) flash sale.
// Helper errJSON dan parseID dipakai bersama dari product_service.go.
type FlashSaleSlotService struct {
	repo *repository.FlashSaleSlotRepository
}

func NewFlashSaleSlotService(repo *repository.FlashSaleSlotRepository) *FlashSaleSlotService {
	return &FlashSaleSlotService{repo: repo}
}

func validateSlotSchedule(flashPrice float64, quota int, startTime, endTime time.Time, isCreate bool) string {
	if flashPrice < 0 {
		return "flash_price tidak boleh negatif"
	}
	if isCreate && quota <= 0 {
		return "quota harus lebih dari 0"
	}
	if !isCreate && quota < 0 {
		return "quota tidak boleh negatif"
	}
	if startTime.IsZero() || endTime.IsZero() {
		return "start_time dan end_time wajib diisi (format RFC3339, contoh: 2026-11-25T19:00:00+07:00)"
	}
	if !endTime.After(startTime) {
		return "end_time harus lebih besar dari start_time"
	}
	return ""
}

// CreateSlot godoc
// @Summary      Buat slot flash sale
// @Description  Membuat jadwal flash sale untuk satu produk (harga flash, kuota, waktu mulai dan selesai). Produk harus ada di PostgreSQL. Format waktu RFC3339.
// @Tags         Flash Sale
// @Accept       json
// @Produce      json
// @Param        body  body      model.CreateFlashSaleSlotRequest  true  "Data slot flash sale"
// @Success      201   {object}  model.FlashSaleSlot
// @Failure      400   {object}  model.ErrorResponse
// @Failure      404   {object}  model.ErrorResponse
// @Failure      500   {object}  model.ErrorResponse
// @Router       /flash-sales [post]
func (s *FlashSaleSlotService) CreateSlot(c *fiber.Ctx) error {
	var req model.CreateFlashSaleSlotRequest
	if err := c.BodyParser(&req); err != nil {
		return errJSON(c, fiber.StatusBadRequest, "format JSON tidak valid (pastikan start_time dan end_time berformat RFC3339)")
	}

	if req.ProductID <= 0 {
		return errJSON(c, fiber.StatusBadRequest, "product_id wajib diisi dan lebih dari 0")
	}
	if msg := validateSlotSchedule(req.FlashPrice, req.Quota, req.StartTime, req.EndTime, true); msg != "" {
		return errJSON(c, fiber.StatusBadRequest, msg)
	}

	slot, err := s.repo.Create(c.UserContext(), req)
	if err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			return errJSON(c, fiber.StatusNotFound, err.Error())
		}
		log.Printf("CreateSlot error: %v", err)
		return errJSON(c, fiber.StatusInternalServerError, "gagal membuat slot flash sale")
	}

	return c.Status(fiber.StatusCreated).JSON(slot)
}

// GetSlots godoc
// @Summary      Daftar slot flash sale
// @Description  Mengambil daftar slot flash sale beserta status (upcoming, active, ended, sold_out). Gunakan active=true untuk hanya slot yang sedang berlangsung dan masih punya kuota.
// @Tags         Flash Sale
// @Produce      json
// @Param        active  query     bool  false  "true = hanya slot yang sedang aktif"
// @Success      200     {array}   model.FlashSaleSlot
// @Failure      500     {object}  model.ErrorResponse
// @Router       /flash-sales [get]
func (s *FlashSaleSlotService) GetSlots(c *fiber.Ctx) error {
	activeOnly := c.Query("active") == "true"

	slots, err := s.repo.FindAll(c.UserContext(), activeOnly)
	if err != nil {
		log.Printf("GetSlots error: %v", err)
		return errJSON(c, fiber.StatusInternalServerError, "gagal mengambil daftar slot flash sale")
	}

	return c.JSON(slots)
}

// GetSlot godoc
// @Summary      Detail slot flash sale
// @Description  Mengambil satu slot flash sale beserta nama dan harga normal produknya.
// @Tags         Flash Sale
// @Produce      json
// @Param        id   path      int  true  "ID slot"
// @Success      200  {object}  model.FlashSaleSlot
// @Failure      400  {object}  model.ErrorResponse
// @Failure      404  {object}  model.ErrorResponse
// @Failure      500  {object}  model.ErrorResponse
// @Router       /flash-sales/{id} [get]
func (s *FlashSaleSlotService) GetSlot(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return errJSON(c, fiber.StatusBadRequest, "id slot tidak valid")
	}

	slot, err := s.repo.FindByID(c.UserContext(), id)
	if err != nil {
		if errors.Is(err, repository.ErrSlotNotFound) {
			return errJSON(c, fiber.StatusNotFound, err.Error())
		}
		log.Printf("GetSlot error: %v", err)
		return errJSON(c, fiber.StatusInternalServerError, "gagal mengambil detail slot flash sale")
	}

	return c.JSON(slot)
}

// UpdateSlot godoc
// @Summary      Edit slot flash sale
// @Description  Mengubah harga flash, kuota, dan jadwal slot. Produk pada slot tidak bisa diganti.
// @Tags         Flash Sale
// @Accept       json
// @Produce      json
// @Param        id    path      int                               true  "ID slot"
// @Param        body  body      model.UpdateFlashSaleSlotRequest  true  "Data slot baru"
// @Success      200   {object}  model.FlashSaleSlot
// @Failure      400   {object}  model.ErrorResponse
// @Failure      404   {object}  model.ErrorResponse
// @Failure      500   {object}  model.ErrorResponse
// @Router       /flash-sales/{id} [put]
func (s *FlashSaleSlotService) UpdateSlot(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return errJSON(c, fiber.StatusBadRequest, "id slot tidak valid")
	}

	var req model.UpdateFlashSaleSlotRequest
	if err := c.BodyParser(&req); err != nil {
		return errJSON(c, fiber.StatusBadRequest, "format JSON tidak valid")
	}
	if msg := validateSlotSchedule(req.FlashPrice, req.Quota, req.StartTime, req.EndTime, false); msg != "" {
		return errJSON(c, fiber.StatusBadRequest, msg)
	}

	slot, err := s.repo.Update(c.UserContext(), id, req)
	if err != nil {
		if errors.Is(err, repository.ErrSlotNotFound) {
			return errJSON(c, fiber.StatusNotFound, err.Error())
		}
		log.Printf("UpdateSlot error: %v", err)
		return errJSON(c, fiber.StatusInternalServerError, "gagal memperbarui slot flash sale")
	}

	return c.JSON(slot)
}

// DeleteSlot godoc
// @Summary      Hapus slot flash sale
// @Description  Menghapus slot. Ditolak (409) jika slot sudah memiliki pesanan.
// @Tags         Flash Sale
// @Produce      json
// @Param        id   path      int  true  "ID slot"
// @Success      200  {object}  model.MessageResponse
// @Failure      400  {object}  model.ErrorResponse
// @Failure      404  {object}  model.ErrorResponse
// @Failure      409  {object}  model.ErrorResponse
// @Failure      500  {object}  model.ErrorResponse
// @Router       /flash-sales/{id} [delete]
func (s *FlashSaleSlotService) DeleteSlot(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return errJSON(c, fiber.StatusBadRequest, "id slot tidak valid")
	}

	if err := s.repo.Delete(c.UserContext(), id); err != nil {
		switch {
		case errors.Is(err, repository.ErrSlotNotFound):
			return errJSON(c, fiber.StatusNotFound, err.Error())
		case errors.Is(err, repository.ErrSlotInUse):
			return errJSON(c, fiber.StatusConflict, err.Error())
		}
		log.Printf("DeleteSlot error: %v", err)
		return errJSON(c, fiber.StatusInternalServerError, "gagal menghapus slot flash sale")
	}

	return c.JSON(model.MessageResponse{Message: "slot flash sale berhasil dihapus"})
}
