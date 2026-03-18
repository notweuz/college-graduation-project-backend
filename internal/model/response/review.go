package response

import "time"

type ReviewShort struct {
	ID        uint64    `json:"id"`
	User      UserShort `json:"user"`
	Rating    uint8     `json:"rating"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
}

func NewReviewShort(id uint64, user UserShort, rating uint8, comment string, createdAt time.Time) *ReviewShort {
	return &ReviewShort{
		ID:        id,
		User:      user,
		Rating:    rating,
		Comment:   comment,
		CreatedAt: createdAt,
	}
}
