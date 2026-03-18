package database

import (
	"college-graduation-project-backend/internal/model"
	"time"

	"gorm.io/gorm"
)

type reportDatabase struct {
	db *gorm.DB
}

func NewReportDatabase(db *gorm.DB) ReportDatabase {
	return &reportDatabase{db: db}
}

func (d *reportDatabase) FindSalesBookings(from, to time.Time, hallID *uint64) ([]model.Booking, error) {
	query := d.db.Model(&model.Booking{}).
		Preload("Hall").
		Where("start_date_time < ? AND end_date_time > ?", to, from)

	if hallID != nil {
		query = query.Where("hall_id = ?", *hallID)
	}

	var bookings []model.Booking
	if err := query.Order("start_date_time").Find(&bookings).Error; err != nil {
		return nil, err
	}

	return bookings, nil
}

func (d *reportDatabase) CountHalls(hallID *uint64) (uint64, error) {
	query := d.db.Model(&model.Hall{})
	if hallID != nil {
		query = query.Where("id = ?", *hallID)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}

	return uint64(count), nil
}
