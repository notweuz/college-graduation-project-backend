package request

type HallCreate struct {
	Name        string  `json:"name" validate:"required"`
	Description string  `json:"description"`
	PricePerDay float64 `json:"price_per_day" validate:"required"`
}

type HallUpdate struct {
	Name        *string  `json:"name"`
	Description *string  `json:"description"`
	PricePerDay *float64 `json:"price_per_day"`
	IsActive    *bool    `json:"is_active"`
}
