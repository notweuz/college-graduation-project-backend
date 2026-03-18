package response

import "time"

type HallFull struct {
	ID           uint64   `json:"id"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	PricePerHour float64  `json:"price_per_hour"`
	IsActive     bool     `json:"is_active"`
	Images       []string `json:"images"`
}

func NewHallFull(id uint64, name, description string, pricePerHour float64, isActive bool, images []string) *HallFull {
	if images == nil {
		images = []string{}
	}
	return &HallFull{
		ID:           id,
		Name:         name,
		Description:  description,
		PricePerHour: pricePerHour,
		IsActive:     isActive,
		Images:       images,
	}
}

type HallAvailability struct {
	StartDateTime time.Time `json:"start_date_time"`
	EndDateTime   time.Time `json:"end_date_time"`
	Status        string    `json:"status"`
}

func NewHallAvailability(startDateTime, endDateTime time.Time, status string) HallAvailability {
	return HallAvailability{
		StartDateTime: startDateTime,
		EndDateTime:   endDateTime,
		Status:        status,
	}
}
