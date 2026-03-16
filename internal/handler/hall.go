package handler

import (
	"college-graduation-project-backend/internal/service"

	"github.com/gofiber/fiber/v3"
)

type HallHandler struct {
	hallService service.HallService
}

func NewHallHandler(hallService service.HallService) *HallHandler {
	return &HallHandler{hallService: hallService}
}

func (h *HallHandler) GetAllHalls(c fiber.Ctx) error {
	halls, err := h.hallService.FindAll()
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusOK).JSON(halls)
}

func (h *HallHandler) GetHallById(c fiber.Ctx) error {
	id := fiber.Params[uint64](c, "id")
	hall, err := h.hallService.FindByID(id)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusOK).JSON(hall)
}
