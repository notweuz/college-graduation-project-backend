package database

import (
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

func (d *bookingDatabase) FindAll() ([]model.Booking, error) {
	var bookings []model.Booking
	if err := d.db.Find(&bookings).Error; err != nil {
		return nil, err
	}
	return bookings, nil
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
	var count int64

	err := d.db.Model(&model.Booking{}).
		Where("hall_id = ?", hallID).
		Where("(start_date_time < ? AND end_date_time > ?) OR (start_date_time < ? AND end_date_time > ?) OR (start_date_time >= ? AND end_date_time <= ?)",
			endDateTime, startDateTime,
			startDateTime, endDateTime,
			startDateTime, endDateTime).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}
