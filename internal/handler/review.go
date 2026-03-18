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

func (h *ReviewHandler) DeleteAdmin(c fiber.Ctx) error {
	reviewID := fiber.Params[uint64](c, "id")
	err := h.reviewService.Delete(reviewID)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusNoContent).Send(nil)
}
