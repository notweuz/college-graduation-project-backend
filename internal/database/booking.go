package database

import (
	"college-graduation-project-backend/internal/model"

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
	if err := d.db.First(&booking, id).Error; err != nil {
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

func (d *bookingDatabase) Update(booking *model.Booking) error {
	return d.db.Save(booking).Error
}

func (d *bookingDatabase) Delete(id uint64) error {
	return d.db.Delete(&model.Booking{}, id).Error
}
