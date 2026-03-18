package service

import (
	"college-graduation-project-backend/internal/database"
	"college-graduation-project-backend/internal/errs"
	"college-graduation-project-backend/internal/model"
	"college-graduation-project-backend/internal/model/request"
	"errors"

	"gorm.io/gorm"
)

type reviewService struct {
	database       database.ReviewDatabase
	userService    UserService
	bookingService BookingService
}

func NewReviewService(Database database.ReviewDatabase, bookingService BookingService, userService UserService) ReviewService {
	return &reviewService{
		database:       Database,
		bookingService: bookingService,
		userService:    userService,
	}
}

func (r reviewService) Create(userID, bookingID uint64, req *request.ReviewCreate) (*model.Review, error) {
	user, err := r.userService.FindByID(userID)
	if err != nil {
		return nil, err
	}
	booking, err := r.bookingService.FindByID(userID, bookingID)
	if err != nil {
		return nil, err
	}
	if booking.User.ID != user.ID {
		return nil, errs.Forbidden("Cannot create review", "This booking does not belong to you")
	}

	existingReview, err := r.database.FindByUserIDAndBookingID(userID, bookingID)
	if err == nil && existingReview != nil {
		return nil, errs.Conflict("Cannot create review", "You have already reviewed this booking")
	}

	review := model.NewReview(0, *booking, *user, req.Rating, req.Comment)
	err = r.database.Create(review)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, errs.Conflict("Cannot create review", "review already exists")
		}
		return nil, errs.InternalServerError("Cannot create review", "internal server error")
	}
	return review, nil
}

func (r reviewService) GetByHallID(hallID uint64) ([]model.Review, error) {
	return r.database.FindByHallID(hallID)
}

func (r reviewService) GetAllWithFilters(hallID, minRating *uint64) ([]model.Review, error) {
	return r.database.FindAllWithFilters(hallID, minRating)
}

func (r reviewService) Update(userID, reviewID uint64, req *request.ReviewCreate) (*model.Review, error) {
	review, err := r.database.FindByID(reviewID)
	if err != nil {
		return nil, errs.NotFound("Review not found", "Review does not exist")
	}

	if review.UserID != userID {
		return nil, errs.Forbidden("Cannot update review", "This review does not belong to you")
	}

	review.Rating = req.Rating
	review.Comment = req.Comment

	err = r.database.Update(review)
	if err != nil {
		return nil, errs.InternalServerError("Cannot update review", "internal server error")
	}

	return review, nil
}

func (r reviewService) Delete(id uint64) error {
	return r.database.Delete(id)
}
