package model

import "time"

type Review struct {
	ID        uint64  `gorm:"primaryKey;autoIncrement"`
	BookingID uint64  `gorm:"not null;uniqueIndex:idx_reviews_booking_id"`
	Booking   Booking `gorm:"foreignKey:BookingID;references:id"`
	UserID    uint64  `gorm:"not null"`
	User      User    `gorm:"foreignKey:UserID;references:id"`
	Rating    uint8   `gorm:"not null"`
	Comment   string
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

func NewReview(id uint64, booking Booking, user User, rating uint8, comment string) *Review {
	return &Review{
		ID:        id,
		Booking:   booking,
		BookingID: booking.ID,
		User:      user,
		UserID:    user.ID,
		Rating:    rating,
		Comment:   comment,
	}
}
