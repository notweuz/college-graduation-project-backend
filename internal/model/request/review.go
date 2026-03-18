package request

type ReviewCreate struct {
	Rating  uint8  `json:"rating" validate:"required,min=1,max=5"`
	Comment string `json:"comment" validate:"max=255"`
}
