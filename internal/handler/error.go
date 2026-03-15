package handler

import (
	"college-graduation-project-backend/internal/errs"
	"errors"

	"github.com/gofiber/fiber/v3"
)

func ErrorHandler(c fiber.Ctx, err error) error {
	var appErr *errs.AppError
	if errors.As(err, &appErr) {
		return c.Status(appErr.Status).JSON(fiber.Map{
			"status":  appErr.Status,
			"message": appErr.Message,
			"reason":  appErr.Reason,
		})
	}

	if errors.As(err, &fiber.ErrNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  fiber.StatusNotFound,
			"message": fiber.ErrNotFound.Message,
			"reason":  fiber.ErrNotFound.Error(),
		})
	}

	return c.Status(500).JSON(fiber.Map{
		"status":  500,
		"message": err.Error(),
		"reason":  "Internal server error",
	})
}
