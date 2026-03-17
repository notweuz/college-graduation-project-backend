package service

import (
	"college-graduation-project-backend/internal/database"
	"college-graduation-project-backend/internal/errs"
	"college-graduation-project-backend/internal/model"
	"college-graduation-project-backend/internal/model/request"
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

func (b bookingService) Create(userID uint64, req *request.BookingCreate) (*model.Booking, error) {
	_, err := b.userService.FindByID(userID)
	if err != nil {
		return nil, err
	}

	hall, err := b.hallService.FindByID(req.HallID)
	if err != nil {
		return nil, err
	}

	if !hall.IsActive {
		return nil, errs.BadRequest("Cannot create booking", "Hall is not active")
	}

	if req.StartDateTime.After(req.EndDateTime) {
		return nil, errs.BadRequest("Cannot create booking", "Start datetime must be before end datetime")
	}

	if req.StartDateTime.Before(time.Now()) {
		return nil, errs.BadRequest("Cannot create booking", "Start datetime cannot be in the past")
	}

	hasConflict, err := b.bookingDatabase.CheckConflict(req.HallID, req.StartDateTime, req.EndDateTime)
	if err != nil {
		return nil, errs.InternalServerError("Cannot check booking conflict", err.Error())
	}

	if hasConflict {
		return nil, errs.Conflict("Cannot create booking", "Hall is already booked for this time period")
	}

	duration := req.EndDateTime.Sub(req.StartDateTime)
	totalPrice := hall.PricePerHour * duration.Hours()

	booking := &model.Booking{
		HallID:        req.HallID,
		UserID:        userID,
		StartDateTime: req.StartDateTime,
		EndDateTime:   req.EndDateTime,
		TotalPrice:    totalPrice,
		Comment:       req.Comment,
	}

	err = b.bookingDatabase.Create(booking)
	if err != nil {
		return nil, errs.InternalServerError("Cannot create booking", err.Error())
	}

	return booking, nil
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
