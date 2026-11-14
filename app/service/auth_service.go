package service

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"flash-sale-api/app/model"
	"flash-sale-api/app/repository"
	"flash-sale-api/helper"
)

type AuthService struct {
	users repository.UserRepository
	jwt   *helper.JWTManager
}

func NewAuthService(
	users repository.UserRepository,
	jwtManager *helper.JWTManager,
) *AuthService {
	return &AuthService{
		users: users,
		jwt:   jwtManager,
	}
}

func (s *AuthService) Register(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var req model.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(req.Email)

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	hashed, err := helper.HashPassword(req.Password)
	if err != nil {
		return helper.Internal(errors.New("gagal memproses password"))
	}

	// Default role "user" untuk mencegah mass assignment
	created, err := s.users.Create(ctx, model.User{
		Username: req.Username,
		Email:    req.Email,
		Password: hashed,
		Role:     "user",
		IsActive: true,
	})
	if err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return helper.Conflict("username atau email sudah dipakai")
		}
		return helper.Internal(err)
	}

	return helper.Created(c, "pendaftaran berhasil", created,
		"/api/v1/users/"+strconv.Itoa(created.ID))
}

func (s *AuthService) Login(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var req model.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	// Mencari user berdasarkan email sesuai LoginRequest
	user, err := s.users.FindByEmail(ctx, strings.TrimSpace(req.Email))
	if err != nil {
		helper.VerifyDummyPassword(req.Password)
		return helper.Unauthorized("email atau password salah")
	}

	if !helper.VerifyPassword(user.Password, req.Password) {
		return helper.Unauthorized("email atau password salah")
	}

	if !user.IsActive {
		return helper.Forbidden("akun dinonaktifkan")
	}

	accessToken, err := s.jwt.GenerateAccess(user)
	if err != nil {
		return helper.Internal(err)
	}

	resp := model.TokenResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int(s.jwt.AccessTTL().Seconds()),
	}

	return helper.Success(c, fiber.StatusOK, "login berhasil", resp)
}

func (s *AuthService) Me(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	authUser, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	user, err := s.users.FindByID(ctx, authUser.UserID)
	if err != nil {
		return helper.Unauthorized("user tidak ditemukan")
	}

	return helper.Success(c, fiber.StatusOK, "profil berhasil diambil", user)
}