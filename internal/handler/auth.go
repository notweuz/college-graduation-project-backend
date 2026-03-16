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
