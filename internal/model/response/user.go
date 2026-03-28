package response

type UserShort struct {
	ID       uint64  `json:"id"`
	FullName *string `json:"full_name"`
	Email    *string `json:"email"`
}

func NewUserShort(id uint64, fullName, email *string) *UserShort {
	return &UserShort{ID: id, FullName: fullName, Email: email}
}

type UserPublicShort struct {
	ID       uint64  `json:"id"`
	FullName *string `json:"full_name"`
	Avatar   *string `json:"avatar"`
}

func NewUserPublicShort(id uint64, fullName, avatar *string) *UserPublicShort {
	return &UserPublicShort{ID: id, FullName: fullName, Avatar: avatar}
}

type Role struct {
	Role string `json:"role"`
}
