package database

import (
	"college-graduation-project-backend/internal/config"
	"college-graduation-project-backend/internal/model"
	"time"

	"gorm.io/gorm"
)

type bookingDatabase struct {
	db *gorm.DB
}

func NewBookingDatabase(database *gorm.DB) BookingDatabase {
	return &bookingDatabase{db: database}
}

func (d *bookingDatabase) Create(booking *model.Booking) error {
	return d.db.Create(booking).Error
}

func (d *bookingDatabase) FindByID(id uint64) (*model.Booking, error) {
	var booking model.Booking
	if err := d.db.Preload("Hall").Preload("User").First(&booking, id).Error; err != nil {
		return nil, err
	}
	return &booking, nil
}

func (d *bookingDatabase) FindAllFromUser(userID uint64, from, to *time.Time) ([]model.Booking, error) {
	var bookings []model.Booking
	query := d.db.Where("user_id = ?", userID)

	if from != nil {
		query = query.Where("start_date_time >= ?", *from)
	}

	if to != nil {
		query = query.Where("end_date_time <= ?", *to)
	}

	if err := query.Preload("Hall").Preload("User").Find(&bookings).Error; err != nil {
		return nil, err
	}
	return bookings, nil
}

func (d *bookingDatabase) Update(booking *model.Booking) error {
	return d.db.Save(booking).Error
}

func (d *bookingDatabase) Delete(id uint64) error {
	return d.db.Delete(&model.Booking{}, id).Error
}

func (d *bookingDatabase) CheckConflict(hallID uint64, startDateTime, endDateTime time.Time) (bool, error) {
	var bookings []model.Booking
	searchFrom := startDateTime.Add(-24 * time.Hour)
	searchTo := endDateTime.Add(24 * time.Hour)
	err := d.db.Model(&model.Booking{}).
		Where("hall_id = ?", hallID).
		Where("start_date_time < ? AND end_date_time > ?", searchTo, searchFrom).
		Find(&bookings).Error

	if err != nil {
		return false, err
	}

	loc := config.AppLocation
	if loc == nil {
		loc = time.UTC
	}

	for _, booking := range bookings {
		existingStart := startOfDayInLocation(booking.StartDateTime, loc)
		existingEnd := startOfDayInLocation(booking.EndDateTime, loc)
		if !isStartOfDayInLocation(booking.EndDateTime, loc) {
			existingEnd = existingEnd.Add(24 * time.Hour)
		}

		if existingStart.Before(endDateTime) && existingEnd.After(startDateTime) {
			return true, nil
		}
	}

	return false, nil
}

func startOfDayInLocation(t time.Time, loc *time.Location) time.Time {
	localTime := t.In(loc)
	return time.Date(localTime.Year(), localTime.Month(), localTime.Day(), 0, 0, 0, 0, loc)
}

func isStartOfDayInLocation(t time.Time, loc *time.Location) bool {
	localTime := t.In(loc)
	return localTime.Hour() == 0 && localTime.Minute() == 0 && localTime.Second() == 0 && localTime.Nanosecond() == 0
}

func (d *bookingDatabase) FindBookingsForHall(hallID uint64, from, to time.Time) ([]model.Booking, error) {
	var bookings []model.Booking

	err := d.db.Where("hall_id = ?", hallID).
		Where("start_date_time < ? AND end_date_time > ?", to, from).
		Order("start_date_time").
		Find(&bookings).Error

	if err != nil {
		return nil, err
	}

	return bookings, nil
}

func (d *bookingDatabase) FindAll(hallID, bookingUserID *uint64, from, to *time.Time) ([]model.Booking, error) {
	query := d.db
	if hallID != nil {
		query = query.Where("hall_id = ?", *hallID)
	}
	if bookingUserID != nil {
		query = query.Where("user_id = ?", *bookingUserID)
	}
	if from != nil {
		query = query.Where("start_date_time >= ?", *from)
	}
	if to != nil {
		query = query.Where("end_date_time <= ?", *to)
	}
	var bookings []model.Booking
	err := query.Preload("Hall").Preload("User").Order("start_date_time").Find(&bookings).Error
	if err != nil {
		return nil, err
	}
	return bookings, nil
}
