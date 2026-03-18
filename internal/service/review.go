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
