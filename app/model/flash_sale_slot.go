package model

import "time"

// FlashSaleSlot adalah jadwal sesi flash sale untuk satu produk.
type FlashSaleSlot struct {
	ID          int       `json:"id" example:"1"`
	ProductID   int       `json:"product_id" example:"4"`
	ProductName string    `json:"product_name,omitempty" example:"Sepatu Lari Limited"`
	NormalPrice float64   `json:"normal_price,omitempty" example:"599000"`
	FlashPrice  float64   `json:"flash_price" example:"299000"`
	Quota       int       `json:"quota" example:"50"`
	StartTime   time.Time `json:"start_time" example:"2026-11-25T19:00:00+07:00"`
	EndTime     time.Time `json:"end_time" example:"2026-11-25T21:00:00+07:00"`
	Status      string    `json:"status" enums:"upcoming,active,ended,sold_out" example:"upcoming"`
	CreatedAt   time.Time `json:"created_at"`
}

// CreateFlashSaleSlotRequest adalah body untuk membuat slot flash sale.
type CreateFlashSaleSlotRequest struct {
	ProductID  int       `json:"product_id" example:"4"`
	FlashPrice float64   `json:"flash_price" example:"299000"`
	Quota      int       `json:"quota" example:"50"`
	StartTime  time.Time `json:"start_time" example:"2026-11-25T19:00:00+07:00"`
	EndTime    time.Time `json:"end_time" example:"2026-11-25T21:00:00+07:00"`
}

// UpdateFlashSaleSlotRequest adalah body untuk mengubah slot flash sale.
type UpdateFlashSaleSlotRequest struct {
	FlashPrice float64   `json:"flash_price" example:"299000"`
	Quota      int       `json:"quota" example:"50"`
	StartTime  time.Time `json:"start_time" example:"2026-11-25T19:00:00+07:00"`
	EndTime    time.Time `json:"end_time" example:"2026-11-25T21:00:00+07:00"`
}

// ComputeStatus menentukan status sesi flash sale berdasarkan waktu saat ini dan sisa kuota.
func (s *FlashSaleSlot) ComputeStatus(now time.Time) string {
	switch {
	case now.Before(s.StartTime):
		return "upcoming"
	case now.After(s.EndTime):
		return "ended"
	case s.Quota <= 0:
		return "sold_out"
	default:
		return "active"
	}
}
