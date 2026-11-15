package service

import (
	"errors"
	"log"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"flash-sale-api/app/model"
	"flash-sale-api/app/repository"
)

// ProductService berisi handler HTTP untuk modul Product.
type ProductService struct {
	repo *repository.ProductRepository
}

func NewProductService(repo *repository.ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

func errJSON(c *fiber.Ctx, status int, msg string) error {
	return c.Status(status).JSON(model.ErrorResponse{Error: msg})
}

func validateProduct(req *model.ProductRequest) string {
	req.Name = strings.TrimSpace(req.Name)
	switch {
	case req.Name == "":
		return "name wajib diisi"
	case len(req.Name) > 255:
		return "name maksimal 255 karakter"
	case req.Price < 0:
		return "price tidak boleh negatif"
	case req.Stock < 0:
		return "stock tidak boleh negatif"
	}
	return ""
}

func parseID(c *fiber.Ctx) (int, error) {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return 0, errors.New("id tidak valid")
	}
	return id, nil
}

// CreateProduct godoc
// @Summary      Tambah produk flash sale
// @Description  Membuat produk baru di PostgreSQL. Field version otomatis bernilai 1.
// @Tags         Products
// @Accept       json
// @Produce      json
// @Param        body  body      model.ProductRequest  true  "Data produk"
// @Success      201   {object}  model.Product
// @Failure      400   {object}  model.ErrorResponse
// @Failure      500   {object}  model.ErrorResponse
// @Router       /products [post]
func (s *ProductService) CreateProduct(c *fiber.Ctx) error {
	var req model.ProductRequest
	if err := c.BodyParser(&req); err != nil {
		return errJSON(c, fiber.StatusBadRequest, "body JSON tidak valid")
	}
	if msg := validateProduct(&req); msg != "" {
		return errJSON(c, fiber.StatusBadRequest, msg)
	}

	product, err := s.repo.Create(c.UserContext(), req)
	if err != nil {
		log.Printf("CreateProduct error: %v", err)
		return errJSON(c, fiber.StatusInternalServerError, "gagal menyimpan produk")
	}
	return c.Status(fiber.StatusCreated).JSON(product)
}

// GetProducts godoc
// @Summary      Daftar produk flash sale
// @Description  Mengambil daftar produk dengan paginasi.
// @Tags         Products
// @Produce      json
// @Param        page   query     int  false  "Halaman (default 1)"
// @Param        limit  query     int  false  "Jumlah per halaman (default 10, maks 100)"
// @Success      200    {array}   model.Product
// @Failure      500    {object}  model.ErrorResponse
// @Router       /products [get]
func (s *ProductService) GetProducts(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	products, err := s.repo.FindAll(c.UserContext(), limit, (page-1)*limit)
	if err != nil {
		return errJSON(c, fiber.StatusInternalServerError, "gagal mengambil produk")
	}
	return c.JSON(products)
}

// GetProduct godoc
// @Summary      Detail produk
// @Description  Mengambil satu produk berdasarkan ID.
// @Tags         Products
// @Produce      json
// @Param        id   path      int  true  "ID produk"
// @Success      200  {object}  model.Product
// @Failure      400  {object}  model.ErrorResponse
// @Failure      404  {object}  model.ErrorResponse
// @Failure      500  {object}  model.ErrorResponse
// @Router       /products/{id} [get]
func (s *ProductService) GetProduct(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return errJSON(c, fiber.StatusBadRequest, err.Error())
	}

	product, err := s.repo.FindByID(c.UserContext(), id)
	if err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			return errJSON(c, fiber.StatusNotFound, err.Error())
		}
		return errJSON(c, fiber.StatusInternalServerError, "gagal mengambil produk")
	}
	return c.JSON(product)
}

// UpdateProduct godoc
// @Summary      Edit produk
// @Description  Mengubah data produk. Field version otomatis naik +1 setiap update.
// @Tags         Products
// @Accept       json
// @Produce      json
// @Param        id    path      int                   true  "ID produk"
// @Param        body  body      model.ProductRequest  true  "Data produk baru"
// @Success      200   {object}  model.Product
// @Failure      400   {object}  model.ErrorResponse
// @Failure      404   {object}  model.ErrorResponse
// @Failure      500   {object}  model.ErrorResponse
// @Router       /products/{id} [put]
func (s *ProductService) UpdateProduct(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return errJSON(c, fiber.StatusBadRequest, err.Error())
	}

	var req model.ProductRequest
	if err := c.BodyParser(&req); err != nil {
		return errJSON(c, fiber.StatusBadRequest, "body JSON tidak valid")
	}
	if msg := validateProduct(&req); msg != "" {
		return errJSON(c, fiber.StatusBadRequest, msg)
	}

	product, err := s.repo.Update(c.UserContext(), id, req)
	if err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			return errJSON(c, fiber.StatusNotFound, err.Error())
		}
		return errJSON(c, fiber.StatusInternalServerError, "gagal mengubah produk")
	}
	return c.JSON(product)
}

// DeleteProduct godoc
// @Summary      Hapus produk
// @Description  Menghapus produk. Ditolak (409) jika produk masih dipakai oleh order.
// @Tags         Products
// @Produce      json
// @Param        id   path      int  true  "ID produk"
// @Success      200  {object}  model.MessageResponse
// @Failure      400  {object}  model.ErrorResponse
// @Failure      404  {object}  model.ErrorResponse
// @Failure      409  {object}  model.ErrorResponse
// @Failure      500  {object}  model.ErrorResponse
// @Router       /products/{id} [delete]
func (s *ProductService) DeleteProduct(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return errJSON(c, fiber.StatusBadRequest, err.Error())
	}

	if err := s.repo.Delete(c.UserContext(), id); err != nil {
		switch {
		case errors.Is(err, repository.ErrProductNotFound):
			return errJSON(c, fiber.StatusNotFound, err.Error())
		case errors.Is(err, repository.ErrProductInUse):
			return errJSON(c, fiber.StatusConflict, err.Error())
		}
		return errJSON(c, fiber.StatusInternalServerError, "gagal menghapus produk")
	}
	return c.JSON(model.MessageResponse{Message: "produk berhasil dihapus"})
}
