package model

import (
	"college-graduation-project-backend/internal/model/enum"
	"time"
)

type User struct {
	ID           uint64        `gorm:"primaryKey;autoIncrement"`
	PasswordHash string        `gorm:"size:255;not null"`
	Email        *string       `gorm:"size:256"`
	FullName     *string       `gorm:"size:150"`
	Role         enum.UserRole `gorm:"type:varchar(20);default:'client'"`
	CreatedAt    time.Time     `gorm:"autoCreateTime"`
}
