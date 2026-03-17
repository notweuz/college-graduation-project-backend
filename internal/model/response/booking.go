package response

import "time"

type BookingFull struct {
	ID            uint64    `json:"id"`
	Hall          HallFull  `json:"hall"`
	User          UserShort `json:"user"`
	StartDateTime time.Time `json:"start_date_time"`
	EndDateTime   time.Time `json:"end_date_time"`
	TotalPrice    float64   `json:"total_price"`
	Comment       string    `json:"comment"`
	CreatedAt     time.Time `json:"created_at"`
}

func NewBookingFull(id uint64, hall HallFull, user UserShort, startDateTime time.Time, endDateTime time.Time, totalPrice float64, comment string, createdAt time.Time) BookingFull {
	return BookingFull{
		ID:            id,
		Hall:          hall,
		User:          user,
		StartDateTime: startDateTime,
		EndDateTime:   endDateTime,
		TotalPrice:    totalPrice,
		Comment:       comment,
		CreatedAt:     createdAt,
	}
}
