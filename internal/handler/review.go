package handler

import (
	"college-graduation-project-backend/internal/middleware"
	"college-graduation-project-backend/internal/model/request"
	"college-graduation-project-backend/internal/model/response"
	"college-graduation-project-backend/internal/service"
	"college-graduation-project-backend/internal/validation"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

type ReviewHandler struct {
	reviewService service.ReviewService
}

func NewReviewHandler(reviewService service.ReviewService) *ReviewHandler {
	return &ReviewHandler{
		reviewService: reviewService,
	}
}

// Create godoc
// @Summary Create review for booking
// @Description Создает отзыв к бронированию от текущего пользователя.
// @Tags reviews
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID бронирования"
// @Param payload body request.ReviewCreate true "Данные отзыва"
// @Success 200 {object} response.ReviewShort
// @Failure 400 {object} BadRequestErrorResponse
// @Failure 401 {object} UnauthorizedErrorResponse
// @Failure 403 {object} ForbiddenErrorResponse
// @Failure 404 {object} NotFoundErrorResponse
// @Failure 409 {object} ConflictErrorResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /api/bookings/{id}/review [post]
func (h *ReviewHandler) Create(c fiber.Ctx) error {
	userID, err := middleware.GetCurrentUserID(c)
	if err != nil {
		return err
	}
	var req request.ReviewCreate
	if err := c.Bind().Body(&req); err != nil {
		return err
	}
	if err := validation.Validate(&req); err != nil {
		return err
	}
	bookingID := fiber.Params[uint64](c, "id")
	review, err := h.reviewService.Create(userID, bookingID, &req)
	if err != nil {
		return err
	}
	userShort := response.NewUserShort(review.User.ID, review.User.Email, review.User.FullName)
	reviewFull := response.NewReviewShort(review.ID, *userShort, review.Rating, review.Comment, review.CreatedAt)
	return c.Status(fiber.StatusOK).JSON(reviewFull)
}

// Update godoc
// @Summary Update review
// @Description Обновляет отзыв текущего пользователя.
// @Tags reviews
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID отзыва"
// @Param payload body request.ReviewCreate true "Данные отзыва"
// @Success 200 {object} response.ReviewShort
// @Failure 400 {object} BadRequestErrorResponse
// @Failure 401 {object} UnauthorizedErrorResponse
// @Failure 403 {object} ForbiddenErrorResponse
// @Failure 404 {object} NotFoundErrorResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /api/reviews/{id} [patch]
func (h *ReviewHandler) Update(c fiber.Ctx) error {
	userID, err := middleware.GetCurrentUserID(c)
	if err != nil {
		return err
	}
	var req request.ReviewCreate
	if err := c.Bind().Body(&req); err != nil {
		return err
	}
	if err := validation.Validate(&req); err != nil {
		return err
	}
	reviewID := fiber.Params[uint64](c, "id")
	review, err := h.reviewService.Update(userID, reviewID, &req)
	if err != nil {
		return err
	}
	userShort := response.NewUserShort(review.User.ID, review.User.Email, review.User.FullName)
	reviewFull := response.NewReviewShort(review.ID, *userShort, review.Rating, review.Comment, review.CreatedAt)
	return c.Status(fiber.StatusOK).JSON(reviewFull)
}

// GetByHallID godoc
// @Summary Get reviews by hall ID
// @Description Возвращает отзывы для конкретного зала.
// @Tags reviews
// @Produce json
// @Param id path int true "ID зала"
// @Success 200 {array} response.ReviewShort
// @Failure 400 {object} SimpleErrorResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /api/halls/{id}/reviews [get]
func (h *ReviewHandler) GetByHallID(c fiber.Ctx) error {
	hallID := fiber.Params[uint64](c, "id")

	if hallID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid hall ID"})
	}

	reviews, err := h.reviewService.GetByHallID(hallID)
	if err != nil {
		return err
	}

	reviewShorts := make([]response.ReviewShort, 0)
	for _, review := range reviews {
		userShort := response.NewUserShort(review.User.ID, review.User.Email, review.User.FullName)
		reviewShort := response.NewReviewShort(review.ID, *userShort, review.Rating, review.Comment, review.CreatedAt)
		reviewShorts = append(reviewShorts, *reviewShort)
	}

	return c.Status(fiber.StatusOK).JSON(reviewShorts)
}

// GetAllAdmin godoc
// @Summary List reviews (admin)
// @Description Возвращает список отзывов с фильтрами (требуется авторизация администратора).
// @Tags admin-reviews
// @Produce json
// @Security BearerAuth
// @Param hall_id query int false "ID зала"
// @Param min_rating query int false "Минимальный рейтинг"
// @Success 200 {object} response.ReviewListResponse
// @Failure 401 {object} UnauthorizedErrorResponse
// @Failure 403 {object} ForbiddenErrorResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /api/admin/reviews/ [get]
func (h *ReviewHandler) GetAllAdmin(c fiber.Ctx) error {
	var hallID, minRating *uint64

	if hallIDStr := c.Query("hall_id"); hallIDStr != "" {
		if id, err := strconv.ParseUint(hallIDStr, 10, 64); err == nil {
			hallID = &id
		}
	}

	if minRatingStr := c.Query("min_rating"); minRatingStr != "" {
		if rating, err := strconv.ParseUint(minRatingStr, 10, 64); err == nil {
			minRating = &rating
		}
	}

	reviews, err := h.reviewService.GetAllWithFilters(hallID, minRating)
	if err != nil {
		return err
	}

	reviewShorts := make([]response.ReviewShort, 0)
	for _, review := range reviews {
		userShort := response.NewUserShort(review.User.ID, review.User.Email, review.User.FullName)
		reviewShort := response.NewReviewShort(review.ID, *userShort, review.Rating, review.Comment, review.CreatedAt)
		reviewShorts = append(reviewShorts, *reviewShort)
	}

	resp := response.NewReviewListResponse(reviewShorts)
	return c.Status(fiber.StatusOK).JSON(resp)
}

// DeleteAdmin godoc
// @Summary Delete review (admin)
// @Description Удаляет отзыв по ID (требуется авторизация администратора).
// @Tags admin-reviews
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID отзыва"
// @Success 204 {string} string "No Content"
// @Failure 401 {object} UnauthorizedErrorResponse
// @Failure 403 {object} ForbiddenErrorResponse
// @Failure 404 {object} NotFoundErrorResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /api/admin/reviews/{id} [delete]
func (h *ReviewHandler) DeleteAdmin(c fiber.Ctx) error {
	reviewID := fiber.Params[uint64](c, "id")
	err := h.reviewService.Delete(reviewID)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusNoContent).Send(nil)
}
