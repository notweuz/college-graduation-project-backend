package model

import "time"

type Booking struct {
	ID            uint64    `gorm:"primaryKey;autoIncrement"`
	HallID        uint64    `gorm:"not null"`
	Hall          Hall      `gorm:"foreignKey:HallID;references:id"`
	StartDateTime time.Time `gorm:"not null"`
	EndDateTime   time.Time `gorm:"not null"`
	TotalPrice    float64   `gorm:"not null"`
	Comment       string
	CreatedAt     time.Time `gorm:"autoCreateTime"`
}
