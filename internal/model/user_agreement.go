package model

import "time"

type UserAgreement struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement"`
	Text      string    `gorm:"type:text;not null"`
	Version   uint64    `gorm:"not null;default:1"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}
