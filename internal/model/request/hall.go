package request

type HallCreate struct {
	Name         string  `json:"name" validate:"required"`
	Description  string  `json:"description"`
	PricePerHour float64 `json:"price_per_hour" validate:"required"`
}
