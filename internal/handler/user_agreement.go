package handler

import (
	"college-graduation-project-backend/internal/middleware"
	"college-graduation-project-backend/internal/model/request"
	"college-graduation-project-backend/internal/model/response"
	"college-graduation-project-backend/internal/service"
	"college-graduation-project-backend/internal/validation"

	"github.com/gofiber/fiber/v3"
)

type UserAgreementHandler struct {
	userAgreementService service.UserAgreementService
}

func NewUserAgreementHandler(userAgreementService service.UserAgreementService) *UserAgreementHandler {
	return &UserAgreementHandler{userAgreementService: userAgreementService}
}

// GetPublic godoc
// @Summary Get user agreement
// @Description Возвращает актуальный текст пользовательского соглашения.
// @Tags public
// @Produce json
// @Success 200 {object} response.UserAgreement
// @Failure 404 {object} NotFoundErrorResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /api/public/user-agreement [get]
func (h *UserAgreementHandler) GetPublic(c fiber.Ctx) error {
	agreement, err := h.userAgreementService.Get()
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(response.NewUserAgreement(agreement.Text, agreement.UpdatedAt, agreement.Version))
}

// UpdateAdmin godoc
// @Summary Update user agreement (admin)
// @Description Обновляет текст пользовательского соглашения (требуется авторизация администратора).
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payload body request.UserAgreementUpdate true "Новая редакция соглашения"
// @Success 200 {object} response.UserAgreement
// @Failure 400 {object} BadRequestErrorResponse
// @Failure 401 {object} UnauthorizedErrorResponse
// @Failure 403 {object} ForbiddenErrorResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /api/admin/user-agreement [put]
func (h *UserAgreementHandler) UpdateAdmin(c fiber.Ctx) error {
	var req request.UserAgreementUpdate
	if err := c.Bind().Body(&req); err != nil {
		return err
	}
	if err := validation.Validate(&req); err != nil {
		return err
	}

	userID, err := middleware.GetCurrentUserID(c)
	if err != nil {
		return err
	}

	agreement, err := h.userAgreementService.Update(userID, req.Text)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(response.NewUserAgreement(agreement.Text, agreement.UpdatedAt, agreement.Version))
}
