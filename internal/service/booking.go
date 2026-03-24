package service

import (
	"college-graduation-project-backend/internal/config"
	"college-graduation-project-backend/internal/database"
	"college-graduation-project-backend/internal/errs"
	"college-graduation-project-backend/internal/model"
	"college-graduation-project-backend/internal/model/enum"
	"college-graduation-project-backend/internal/model/request"
	"errors"
	"time"

	"gorm.io/gorm"
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

func (b *bookingService) Create(userID uint64, req *request.BookingCreate) (*model.Booking, error) {
	user, err := b.userService.FindByID(userID)
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

	normalizedStart, normalizedEnd := normalizeBookingRange(req.StartDateTime, req.EndDateTime)
	if !normalizedStart.Before(normalizedEnd) {
		return nil, errs.BadRequest("Cannot create booking", "Booking duration must include at least one full day")
	}

	if normalizedStart.Before(time.Now()) {
		return nil, errs.BadRequest("Cannot create booking", "Start datetime cannot be in the past")
	}

	hasConflict, err := b.bookingDatabase.CheckConflict(req.HallID, normalizedStart, normalizedEnd)
	if err != nil {
		return nil, errs.InternalServerError("Cannot check booking conflict", err.Error())
	}

	if hasConflict {
		return nil, errs.Conflict("Cannot create booking", "Hall is already booked for this time period")
	}

	duration := normalizedEnd.Sub(normalizedStart)
	days := duration.Hours() / 24
	totalPrice := hall.PricePerDay * days

	booking := &model.Booking{
		User:          *user,
		Hall:          *hall,
		HallID:        req.HallID,
		UserID:        userID,
		StartDateTime: normalizedStart,
		EndDateTime:   normalizedEnd,
		TotalPrice:    totalPrice,
		Comment:       req.Comment,
	}

	err = b.bookingDatabase.Create(booking)
	if err != nil {
		return nil, errs.InternalServerError("Cannot create booking", err.Error())
	}

	return booking, nil
}

func normalizeBookingRange(start, end time.Time) (time.Time, time.Time) {
	loc := bookingCalendarLocation()
	normalizedStart := startOfDayInLocation(start, loc)
	normalizedEnd := startOfDayInLocation(end, loc)
	if !isStartOfDayInLocation(end, loc) {
		normalizedEnd = normalizedEnd.Add(24 * time.Hour)
	}
	return normalizedStart, normalizedEnd
}

func bookingCalendarLocation() *time.Location {
	if config.AppLocation != nil {
		return config.AppLocation
	}
	return time.UTC
}

func startOfDayInLocation(t time.Time, loc *time.Location) time.Time {
	localTime := t.In(loc)
	return time.Date(localTime.Year(), localTime.Month(), localTime.Day(), 0, 0, 0, 0, loc)
}

func isStartOfDayInLocation(t time.Time, loc *time.Location) bool {
	localTime := t.In(loc)
	return localTime.Hour() == 0 && localTime.Minute() == 0 && localTime.Second() == 0 && localTime.Nanosecond() == 0
}

func (b *bookingService) FindAllFromUser(userID uint64, from, to *time.Time) ([]model.Booking, error) {
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

func (b *bookingService) FindByID(userID, id uint64) (*model.Booking, error) {
	user, err := b.userService.FindByID(userID)
	if err != nil {
		return nil, err
	}

	booking, err := b.bookingDatabase.FindByID(id)
	if err != nil {
		return nil, err
	}

	if user.Role == enum.RoleAdmin {
		return booking, nil
	}

	if booking.User.ID != user.ID {
		return nil, errs.Forbidden("Cannot find booking by ID", "This booking does not belong to you")
	}

	return booking, nil
}

func (b *bookingService) DeleteByAuthor(userID, id uint64) error {
	_, err := b.userService.FindByID(userID)
	if err != nil {
		return err
	}
	booking, err := b.FindByID(userID, id)
	if err != nil {
		return err
	}
	if booking.StartDateTime.Before(time.Now()) {
		return errs.Forbidden("Cannot delete booking", "Cannot delete booking that was already started/done")
	}
	err = b.bookingDatabase.Delete(id)
	if err != nil {
		return err
	}
	return nil
}

func (b *bookingService) FindAll(userID uint64, hallID, bookingUserID *uint64, from, to *time.Time) ([]model.Booking, error) {
	user, err := b.userService.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if user.Role != enum.RoleAdmin {
		return nil, errs.Forbidden("Cannot get all bookings", "You do not have permission to get all bookings")
	}
	bookings, err := b.bookingDatabase.FindAll(hallID, bookingUserID, from, to)
	if err != nil {
		return nil, errs.InternalServerError("Cannot get all bookings", err.Error())
	}
	return bookings, nil
}

func (b *bookingService) Update(userID, id uint64, req *request.BookingUpdate) (*model.Booking, error) {
	user, err := b.userService.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if user.Role != enum.RoleAdmin {
		return nil, errs.Forbidden("Cannot update booking", "You do not have permission to update this booking")
	}
	booking, err := b.FindByID(userID, id)
	if err != nil {
		return nil, err
	}
	if req.Comment != nil {
		booking.Comment = *req.Comment
	}
	if req.TotalPrice != nil {
		booking.TotalPrice = *req.TotalPrice
	}
	err = b.bookingDatabase.Update(booking)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NotFound("Cannot update booking", "Booking not found")
		}
		return nil, errs.InternalServerError("Cannot update booking", err.Error())
	}
	return booking, nil
}
