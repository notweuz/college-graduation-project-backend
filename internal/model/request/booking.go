package request

import "time"

type BookingCreate struct {
	HallID        uint64    `json:"hall_id" validate:"required"`
	StartDateTime time.Time `json:"start_date_time" validate:"required"`
	EndDateTime   time.Time `json:"end_date_time" validate:"required"`
	Comment       string    `json:"comment" validate:"max:255"`
}

type BookingUpdate struct {
	TotalPrice *float64 `json:"total_price" validate:"required"`
	Comment    *string  `json:"comment" validate:"max:255"`
}
