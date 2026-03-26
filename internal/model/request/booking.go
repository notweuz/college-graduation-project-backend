package request

type BookingCreate struct {
	HallID        uint64 `json:"hall_id" validate:"required"`
	StartDateTime string `json:"start_date_time" validate:"required,datetime=2006-01-02"`
	EndDateTime   string `json:"end_date_time" validate:"required,datetime=2006-01-02"`
	Comment       string `json:"comment" validate:"max=255"`
}

type BookingUpdate struct {
	TotalPrice *float64 `json:"total_price"`
	Comment    *string  `json:"comment" validate:"max=255"`
}
