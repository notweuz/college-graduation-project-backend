package model

import "time"

type Review struct {
	ID        uint64  `gorm:"primaryKey;autoIncrement"`
	BookingID uint64  `gorm:"not null"`
	Booking   Booking `gorm:"foreignKey:BookingID;references:id"`
	UserID    uint64  `gorm:"not null"`
	User      User    `gorm:"foreignKey:UserID;references:id"`
	Rating    uint8   `gorm:"not null"`
	Comment   string
	CreatedAt time.Time `gorm:"autoCreateTime"`
}
