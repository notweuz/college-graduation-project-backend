package request

type UpdateProfile struct {
	Email    *string `json:"email" validate:"omitempty,email"`
	FullName *string `json:"full_name" validate:"omitempty,min=2,max=150"`
}
