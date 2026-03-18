package handler

import (
	"college-graduation-project-backend/internal/model/request"
	"college-graduation-project-backend/internal/model/response"
	"college-graduation-project-backend/internal/service"
	"college-graduation-project-backend/internal/validation"

	"github.com/gofiber/fiber/v3"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register godoc
// @Summary Register user
// @Description Регистрирует нового пользователя и возвращает JWT токен.
// @Tags auth
// @Accept json
// @Produce json
// @Param payload body request.Register true "Данные регистрации"
// @Success 200 {object} response.Token
// @Failure 400 {object} BadRequestErrorResponse
// @Failure 409 {object} ConflictErrorResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /api/auth/register [post]
func (h *AuthHandler) Register(c fiber.Ctx) error {
	var req request.Register
	if err := c.Bind().Body(&req); err != nil {
		return err
	}
	if err := validation.Validate(&req); err != nil {
		return err
	}
	token, err := h.authService.Register(&req)
	if err != nil {
		return err
	}
	return c.JSON(response.Token{Token: token})
}

// Login godoc
// @Summary Login user
// @Description Аутентифицирует пользователя и возвращает JWT токен.
// @Tags auth
// @Accept json
// @Produce json
// @Param payload body request.Login true "Данные входа"
// @Success 200 {object} response.Token
// @Failure 400 {object} BadRequestErrorResponse
// @Failure 401 {object} UnauthorizedErrorResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /api/auth/login [post]
func (h *AuthHandler) Login(c fiber.Ctx) error {
	var req request.Login
	if err := c.Bind().Body(&req); err != nil {
		return err
	}
	if err := validation.Validate(&req); err != nil {
		return err
	}
	token, err := h.authService.Login(&req)
	if err != nil {
		return err
	}
	return c.JSON(response.Token{Token: token})
}
