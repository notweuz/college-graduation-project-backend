package request

type UserAgreementUpdate struct {
	Text string `json:"text" validate:"required"`
}
