package errs

import "github.com/gofiber/fiber/v3"

type AppError struct {
	Status  int
	Message string
	Reason  string
}

func (e *AppError) Error() string {
	return e.Message
}

func NotFound(message, reason string) *AppError {
	return &AppError{
		Status:  fiber.StatusNotFound,
		Message: message,
		Reason:  reason,
	}
}

func Unauthorized(message, reason string) *AppError {
	return &AppError{
		Status:  fiber.StatusUnauthorized,
		Message: message,
		Reason:  reason,
	}
}

func BadRequest(message, reason string) *AppError {
	return &AppError{
		Status:  fiber.StatusBadRequest,
		Message: message,
		Reason:  reason,
	}
}

func Conflict(message, reason string) *AppError {
	return &AppError{
		Status:  fiber.StatusConflict,
		Message: message,
		Reason:  reason,
	}
}

func InternalServerError(message, reason string) *AppError {
	return &AppError{
		Status:  fiber.StatusInternalServerError,
		Message: message,
		Reason:  reason,
	}
}

func Forbidden(message, reason string) *AppError {
	return &AppError{
		Status:  fiber.StatusForbidden,
		Message: message,
		Reason:  reason,
	}
}

func UnprocessableEntity(message, reason string) *AppError {
	return &AppError{
		Status:  fiber.StatusUnprocessableEntity,
		Message: message,
		Reason:  reason,
	}
}
