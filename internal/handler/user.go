package handler

import (
	"college-graduation-project-backend/internal/middleware"
	"college-graduation-project-backend/internal/model/request"
	"college-graduation-project-backend/internal/model/response"
	"college-graduation-project-backend/internal/service"

	"github.com/gofiber/fiber/v3"
)

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) GetProfile(c fiber.Ctx) error {
	userId, err := middleware.GetCurrentUserID(c)
	if err != nil {
		return err
	}

	user, err := h.userService.FindByID(userId)
	if err != nil {
		return err
	}
	userShort := response.NewUserShort(user.ID, user.FullName, user.Email)

	return c.Status(fiber.StatusOK).JSON(userShort)
}

func (h *UserHandler) UpdateProfile(c fiber.Ctx) error {
	var updateProfile request.UpdateProfile
	if err := c.Bind().Body(&updateProfile); err != nil {
		return err
	}
	user, err := h.userService.UpdateProfile(c, updateProfile)
	if err != nil {
		return err
	}
	userShort := response.NewUserShort(user.ID, user.FullName, user.Email)
	return c.Status(fiber.StatusOK).JSON(userShort)
}
