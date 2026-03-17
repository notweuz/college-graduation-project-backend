package service

import (
	"college-graduation-project-backend/internal/database"
	"college-graduation-project-backend/internal/errs"
	"college-graduation-project-backend/internal/model"
	"time"
)

type bookingService struct {
	bookingDatabase database.BookingDatabase
	userService     UserService
	hallService     HallService
}

func NewBookingService(bookingDatabase database.BookingDatabase, userService UserService, hallService HallService) BookingService {
	return &bookingService{
		bookingDatabase: bookingDatabase,
		userService:     userService,
		hallService:     hallService,
	}
}

func (b bookingService) FindAllFromUser(userID uint64, from, to *time.Time) ([]model.Booking, error) {
	_, err := b.userService.FindByID(userID)
	if err != nil {
		return nil, err
	}
	bookings, err := b.bookingDatabase.FindAllFromUser(userID, from, to)
	if err != nil {
		return nil, errs.InternalServerError("Cannot get bookings from user", err.Error())
	}
	return bookings, nil
}

func (b bookingService) FindByID(userID, id uint64) (*model.Booking, error) {
	user, err := b.userService.FindByID(userID)
	if err != nil {
		return nil, err
	}

	booking, err := b.bookingDatabase.FindByID(id)
	if err != nil {
		return nil, err
	}
	if booking.User.ID != user.ID {
		return nil, errs.Forbidden("Cannot find booking by ID", "This booking does not belong to you")
	}

	return booking, nil
}
