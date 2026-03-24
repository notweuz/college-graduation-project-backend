package response

import "time"

type UserAgreement struct {
	Text      string    `json:"text"`
	UpdatedAt time.Time `json:"updated_at"`
	Version   uint64    `json:"version"`
}

func NewUserAgreement(text string, updatedAt time.Time, version uint64) *UserAgreement {
	return &UserAgreement{Text: text, UpdatedAt: updatedAt, Version: version}
}
