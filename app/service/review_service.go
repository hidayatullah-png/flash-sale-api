package service

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"

	"flash-sale-api/app/model"
	"flash-sale-api/app/repository"
)

// ReviewService berisi handler HTTP untuk modul Review.
// Ulasan disimpan di MongoDB, sedangkan pengecekan produk memakai PostgreSQL (polyglot).
type ReviewService struct {
	reviewRepo  *repository.ReviewRepository
	productRepo *repository.ProductRepository
}

func NewReviewService(reviewRepo *repository.ReviewRepository, productRepo *repository.ProductRepository) *ReviewService {
	return &ReviewService{reviewRepo: reviewRepo, productRepo: productRepo}
}

// CreateReview godoc
// @Summary      Tambah ulasan produk
// @Description  Menyimpan komentar & rating (1-5) ke MongoDB. Produk harus ada di PostgreSQL.
// @Tags         Reviews
// @Accept       json
// @Produce      json
// @Param        id    path      int                  true  "ID produk"
// @Param        body  body      model.ReviewRequest  true  "Data ulasan"
// @Success      201   {object}  model.Review
// @Failure      400   {object}  model.ErrorResponse
// @Failure      404   {object}  model.ErrorResponse
// @Failure      500   {object}  model.ErrorResponse
// @Router       /products/{id}/reviews [post]
func (s *ReviewService) CreateReview(c *fiber.Ctx) error {
	productID, err := parseID(c)
	if err != nil {
		return errJSON(c, fiber.StatusBadRequest, err.Error())
	}

	var req model.ReviewRequest
	if err := c.BodyParser(&req); err != nil {
		return errJSON(c, fiber.StatusBadRequest, "body JSON tidak valid")
	}
	req.Comment = strings.TrimSpace(req.Comment)
	switch {
	case req.UserID <= 0:
		return errJSON(c, fiber.StatusBadRequest, "user_id wajib diisi")
	case req.Rating < 1 || req.Rating > 5:
		return errJSON(c, fiber.StatusBadRequest, "rating harus antara 1 sampai 5")
	case len(req.Comment) > 1000:
		return errJSON(c, fiber.StatusBadRequest, "comment maksimal 1000 karakter")
	}

	// Cek produk di PostgreSQL sebelum menyimpan ulasan ke MongoDB.
	if _, err := s.productRepo.FindByID(c.UserContext(), productID); err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			return errJSON(c, fiber.StatusNotFound, err.Error())
		}
		return errJSON(c, fiber.StatusInternalServerError, "gagal memeriksa produk")
	}

	review := &model.Review{
		ProductID: productID,
		UserID:    req.UserID,
		Rating:    req.Rating,
		Comment:   req.Comment,
	}
	if err := s.reviewRepo.Create(c.UserContext(), review); err != nil {
		return errJSON(c, fiber.StatusInternalServerError, "gagal menyimpan ulasan")
	}
	return c.Status(fiber.StatusCreated).JSON(review)
}

// GetReviews godoc
// @Summary      Daftar ulasan produk
// @Description  Mengambil ulasan satu produk dari MongoDB, terbaru lebih dulu.
// @Tags         Reviews
// @Produce      json
// @Param        id     path      int  true   "ID produk"
// @Param        page   query     int  false  "Halaman (default 1)"
// @Param        limit  query     int  false  "Jumlah per halaman (default 10, maks 100)"
// @Success      200    {array}   model.Review
// @Failure      400    {object}  model.ErrorResponse
// @Failure      500    {object}  model.ErrorResponse
// @Router       /products/{id}/reviews [get]
func (s *ReviewService) GetReviews(c *fiber.Ctx) error {
	productID, err := parseID(c)
	if err != nil {
		return errJSON(c, fiber.StatusBadRequest, err.Error())
	}
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	reviews, err := s.reviewRepo.FindByProductID(c.UserContext(), productID, int64(limit), int64((page-1)*limit))
	if err != nil {
		return errJSON(c, fiber.StatusInternalServerError, "gagal mengambil ulasan")
	}
	return c.JSON(reviews)
}

// GetRatingSummary godoc
// @Summary      Ringkasan rating produk
// @Description  Rata-rata rating dan jumlah ulasan, dihitung dengan aggregation pipeline MongoDB.
// @Tags         Reviews
// @Produce      json
// @Param        id   path      int  true  "ID produk"
// @Success      200  {object}  model.RatingSummary
// @Failure      400  {object}  model.ErrorResponse
// @Failure      500  {object}  model.ErrorResponse
// @Router       /products/{id}/reviews/summary [get]
func (s *ReviewService) GetRatingSummary(c *fiber.Ctx) error {
	productID, err := parseID(c)
	if err != nil {
		return errJSON(c, fiber.StatusBadRequest, err.Error())
	}

	summary, err := s.reviewRepo.GetRatingSummary(c.UserContext(), productID)
	if err != nil {
		return errJSON(c, fiber.StatusInternalServerError, "gagal menghitung rating")
	}
	return c.JSON(summary)
}

// DeleteReview godoc
// @Summary      Hapus ulasan
// @Description  Menghapus satu ulasan dari MongoDB berdasarkan ObjectID.
// @Tags         Reviews
// @Produce      json
// @Param        reviewId  path      string  true  "ObjectID ulasan"
// @Success      200       {object}  model.MessageResponse
// @Failure      400       {object}  model.ErrorResponse
// @Failure      404       {object}  model.ErrorResponse
// @Failure      500       {object}  model.ErrorResponse
// @Router       /reviews/{reviewId} [delete]
func (s *ReviewService) DeleteReview(c *fiber.Ctx) error {
	err := s.reviewRepo.Delete(c.UserContext(), c.Params("reviewId"))
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrInvalidReviewID):
			return errJSON(c, fiber.StatusBadRequest, err.Error())
		case errors.Is(err, repository.ErrReviewNotFound):
			return errJSON(c, fiber.StatusNotFound, err.Error())
		}
		return errJSON(c, fiber.StatusInternalServerError, "gagal menghapus ulasan")
	}
	return c.JSON(model.MessageResponse{Message: "ulasan berhasil dihapus"})
}
