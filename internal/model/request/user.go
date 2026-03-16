package request

type UpdateProfile struct {
	Email    *string `json:"email"`
	FullName *string `json:"full_name"`
}
