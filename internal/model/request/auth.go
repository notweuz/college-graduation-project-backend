package request

type Register struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
}

type Login struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
