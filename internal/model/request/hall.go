package request

type HallCreate struct {
	Name        string   `json:"name" validate:"required"`
	Description string   `json:"description"`
	PricePerDay *float64 `json:"price_per_day" validate:"required,gt=0" minimum:"0" example:"100"`
}

type HallUpdate struct {
	Name        *string  `json:"name"`
	Description *string  `json:"description"`
	PricePerDay *float64 `json:"price_per_day" validate:"omitempty,gt=0" minimum:"0" example:"100"`
	IsActive    *bool    `json:"is_active"`
}
