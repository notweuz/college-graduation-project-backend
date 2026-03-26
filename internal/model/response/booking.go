package response

import (
	"college-graduation-project-backend/internal/datetime"
	"time"
)

type BookingFull struct {
	ID            uint64    `json:"id"`
	Hall          HallFull  `json:"hall"`
	User          UserShort `json:"user"`
	StartDateTime string    `json:"start_date_time"`
	EndDateTime   string    `json:"end_date_time"`
	TotalPrice    float64   `json:"total_price"`
	Comment       string    `json:"comment"`
	CreatedAt     time.Time `json:"created_at"`
}

func NewBookingFull(id uint64, hall HallFull, user UserShort, startDateTime time.Time, endDateTime time.Time, totalPrice float64, comment string, createdAt time.Time) BookingFull {
	return BookingFull{
		ID:            id,
		Hall:          hall,
		User:          user,
		StartDateTime: datetime.Format(startDateTime),
		EndDateTime:   datetime.Format(endDateTime),
		TotalPrice:    totalPrice,
		Comment:       comment,
		CreatedAt:     createdAt,
	}
}

type CalculatedPrice struct {
	Price    float64 `json:"price"`
	Discount float64 `json:"discount"`
	Total    float64 `json:"total"`
}

func NewCalculatedPrice(price, discount, total float64) *CalculatedPrice {
	return &CalculatedPrice{
		Price:    price,
		Discount: discount,
		Total:    total,
	}
}
