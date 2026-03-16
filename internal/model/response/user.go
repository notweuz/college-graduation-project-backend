package response

type UserShort struct {
	ID       uint64  `json:"id"`
	FullName *string `json:"full_name"`
	Email    *string `json:"email"`
}

func NewUserShort(id uint64, fullName, email *string) *UserShort {
	return &UserShort{ID: id, FullName: fullName, Email: email}
}
