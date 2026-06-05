package request

type UpdateProfile struct {
	Email    *string `json:"email" validate:"required,email"`
	FullName *string `json:"full_name" validate:"required,min=2,max=150"`
}
