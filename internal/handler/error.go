package handler

import (
	"college-graduation-project-backend/internal/errs"

	"github.com/gofiber/fiber/v3"
)

func ErrorHandler(c fiber.Ctx, err error) error {
	if appErr, ok := err.(*errs.AppError); ok {
		return c.Status(appErr.Status).JSON(fiber.Map{
			"status":  appErr.Status,
			"message": appErr.Message,
			"reason":  appErr.Reason,
		})
	}

	return c.Status(500).JSON(fiber.Map{
		"status":  500,
		"message": err.Error(),
		"reason":  "Internal server error",
	})
}
