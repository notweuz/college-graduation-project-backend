package request

type ReviewCreate struct {
	Rating  uint8  `json:"rating" validate:"required,range=1-5"`
	Comment string `json:"comment" validate:"max=255"`
}
