package model

type Hall struct {
	ID           uint64  `gorm:"primaryKey;autoIncrement"`
	Name         string  `gorm:"size:30;not null"`
	Description  string  `gorm:"size255"`
	PricePerHour float64 `gorm:"not null"`
	IsActive     bool
}
