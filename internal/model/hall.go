package model

type Hall struct {
	ID          uint64  `gorm:"primaryKey;autoIncrement"`
	Name        string  `gorm:"size:30;not null"`
	Description string  `gorm:"size255"`
	PricePerDay float64 `gorm:"not null"`
	IsActive    bool
}

func NewHall(id uint64, name, description string, pricePerDay float64, isActive bool) *Hall {
	return &Hall{
		ID:          id,
		Name:        name,
		Description: description,
		PricePerDay: pricePerDay,
		IsActive:    isActive,
	}
}
