package response

type HallFull struct {
	ID           uint64  `json:"id"`
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	PricePerHour float64 `json:"price_per_hour"`
	IsActive     bool    `json:"is_active"`
}

func NewHallFull(id uint64, name, description string, pricePerHour float64, isActive bool) *HallFull {
	return &HallFull{
		ID:           id,
		Name:         name,
		Description:  description,
		PricePerHour: pricePerHour,
		IsActive:     isActive,
	}
}
