package handler

type SimpleErrorResponse struct {
	Error string `json:"error" example:"Invalid hall ID"`
}

type UploadImageResponse struct {
	ImagePath string `json:"image_path" example:"/api/images/1_1738662000.jpg"`
}

type BadRequestErrorResponse struct {
	Status  int    `json:"status" example:"400"`
	Message string `json:"message" example:"Bad request"`
	Reason  string `json:"reason" example:"invalid request payload"`
}

type UnauthorizedErrorResponse struct {
	Status  int    `json:"status" example:"401"`
	Message string `json:"message" example:"Unauthorized"`
	Reason  string `json:"reason" example:"missing or invalid bearer token"`
}

type ForbiddenErrorResponse struct {
	Status  int    `json:"status" example:"403"`
	Message string `json:"message" example:"Forbidden"`
	Reason  string `json:"reason" example:"insufficient permissions"`
}

type NotFoundErrorResponse struct {
	Status  int    `json:"status" example:"404"`
	Message string `json:"message" example:"Not found"`
	Reason  string `json:"reason" example:"resource not found"`
}

type ConflictErrorResponse struct {
	Status  int    `json:"status" example:"409"`
	Message string `json:"message" example:"Conflict"`
	Reason  string `json:"reason" example:"resource already exists or state conflict"`
}

type UnprocessableEntityErrorResponse struct {
	Status  int    `json:"status" example:"422"`
	Message string `json:"message" example:"Unprocessable entity"`
	Reason  string `json:"reason" example:"validation failed"`
}

type InternalServerErrorResponse struct {
	Status  int    `json:"status" example:"500"`
	Message string `json:"message" example:"Internal server error"`
	Reason  string `json:"reason" example:"unexpected error"`
}
