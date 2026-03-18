package handler

import (
	"college-graduation-project-backend/internal/middleware"
	"college-graduation-project-backend/internal/model/request"
	"college-graduation-project-backend/internal/model/response"
	"college-graduation-project-backend/internal/service"

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
	bookingID := fiber.Params[uint64](c, "id")
	review, err := h.reviewService.Create(userID, bookingID, &req)
	if err != nil {
		return err
	}
	userShort := response.NewUserShort(review.User.ID, review.User.Email, review.User.FullName)
	reviewFull := response.NewReviewShort(review.ID, *userShort, review.Rating, review.Comment, review.CreatedAt)
	return c.Status(fiber.StatusOK).JSON(reviewFull)
}
