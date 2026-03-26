package service

import (
	"college-graduation-project-backend/internal/config"
	"college-graduation-project-backend/internal/database"
	"college-graduation-project-backend/internal/datetime"
	"college-graduation-project-backend/internal/errs"
	"college-graduation-project-backend/internal/model"
	"college-graduation-project-backend/internal/model/enum"
	"college-graduation-project-backend/internal/model/request"
	"college-graduation-project-backend/internal/model/response"
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

	startDate, err := datetime.Parse(req.StartDateTime)
	if err != nil {
		return nil, errs.BadRequest("Cannot create booking", "'start_date_time' must be in YYYY-MM-DD format")
	}
	endDate, err := datetime.Parse(req.EndDateTime)
	if err != nil {
		return nil, errs.BadRequest("Cannot create booking", "'end_date_time' must be in YYYY-MM-DD format")
	}

	normalizedStart, normalizedEnd := normalizeBookingRange(startDate, endDate)
	if normalizedStart.After(normalizedEnd) {
		return nil, errs.BadRequest("Cannot create booking", "Start date must be before or equal to end date")
	}

	if normalizedStart.Before(datetime.StartOfDay(time.Now())) {
		return nil, errs.BadRequest("Cannot create booking", "Start date cannot be in the past")
	}

	hasConflict, err := b.bookingDatabase.CheckConflict(req.HallID, normalizedStart, normalizedEnd)
	if err != nil {
		return nil, errs.InternalServerError("Cannot create booking", err.Error())
	}
	if hasConflict {
		return nil, errs.Conflict("Cannot create booking", "Selected dates conflict with an existing booking")
	}

	calculatedPrice, err := b.CalculatePrice(req.HallID, normalizedStart, normalizedEnd)
	if err != nil {
		return nil, errs.InternalServerError("Cannot calculate booking price", err.Error())
	}

	booking := &model.Booking{
		User:          *user,
		Hall:          *hall,
		HallID:        req.HallID,
		UserID:        userID,
		StartDateTime: normalizedStart,
		EndDateTime:   normalizedEnd,
		TotalPrice:    calculatedPrice.Total,
		Comment:       req.Comment,
	}

	err = b.bookingDatabase.Create(booking)
	if err != nil {
		return nil, errs.InternalServerError("Cannot create booking", err.Error())
	}

	return booking, nil
}

func normalizeBookingRange(start, end time.Time) (time.Time, time.Time) {
	return datetime.StartOfDay(start), datetime.EndOfDay(end)
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

func (b *bookingService) CalculatePrice(hallID uint64, from, to time.Time) (*response.CalculatedPrice, error) {
	hall, err := b.hallService.FindByID(hallID)
	if err != nil {
		return nil, err
	}
	normalizedStart, normalizedEnd := normalizeBookingRange(from, to)
	days := billableDaysBetween(normalizedStart, normalizedEnd)
	defaultPrice := hall.PricePerDay * days
	discount := 0.0
	newPrice := 0.0
	if days >= 2 {
		newPrice = defaultPrice * 0.7
		discount = newPrice - defaultPrice
	}
	return response.NewCalculatedPrice(defaultPrice, discount, newPrice), nil
}
