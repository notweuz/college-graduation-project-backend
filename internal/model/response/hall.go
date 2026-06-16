package response

import (
	"college-graduation-project-backend/internal/datetime"
	"time"
)

type HallFull struct {
	ID          uint64   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	PricePerDay float64  `json:"price_per_day"`
	IsActive    bool     `json:"is_active"`
	Images      []string `json:"images"`
}

func NewHallFull(id uint64, name, description string, pricePerDay float64, isActive bool, images []string) *HallFull {
	if images == nil {
		images = []string{}
	}
	return &HallFull{
		ID:          id,
		Name:        name,
		Description: description,
		PricePerDay: pricePerDay,
		IsActive:    isActive,
		Images:      images,
	}
}

type HallImageWithID struct {
	ID   uint64 `json:"id"`
	Path string `json:"path"`
}

type HallAvailability struct {
	StartDateTime string `json:"start_date_time"`
	EndDateTime   string `json:"end_date_time"`
	Status        string `json:"status"`
}

func NewHallAvailability(startDateTime, endDateTime time.Time, status string) HallAvailability {
	return HallAvailability{
		StartDateTime: datetime.Format(startDateTime),
		EndDateTime:   datetime.Format(endDateTime),
		Status:        status,
	}
}
