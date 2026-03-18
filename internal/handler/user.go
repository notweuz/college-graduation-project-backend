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

// GetProfile godoc
// @Summary Get current user profile
// @Description Возвращает профиль текущего авторизованного пользователя.
// @Tags users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.UserShort
// @Failure 401 {object} UnauthorizedErrorResponse
// @Failure 404 {object} NotFoundErrorResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /api/users/me [get]
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

// UpdateProfile godoc
// @Summary Update current user profile
// @Description Обновляет профиль текущего авторизованного пользователя.
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payload body request.UpdateProfile true "Поля профиля для обновления"
// @Success 200 {object} response.UserShort
// @Failure 400 {object} BadRequestErrorResponse
// @Failure 401 {object} UnauthorizedErrorResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /api/users/me [patch]
func (h *UserHandler) UpdateProfile(c fiber.Ctx) error {
	var updateProfile request.UpdateProfile
	if err := c.Bind().Body(&updateProfile); err != nil {
		return err
	}
	userID, err := middleware.GetCurrentUserID(c)
	if err != nil {
		return err
	}
	user, err := h.userService.UpdateProfile(userID, updateProfile)
	if err != nil {
		return err
	}
	userShort := response.NewUserShort(user.ID, user.FullName, user.Email)
	return c.Status(fiber.StatusOK).JSON(userShort)
}
