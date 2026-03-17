package handler

import (
	"college-graduation-project-backend/internal/model/request"
	"college-graduation-project-backend/internal/model/response"
	"college-graduation-project-backend/internal/service"
	"college-graduation-project-backend/internal/validation"

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

func (h *HallHandler) Create(c fiber.Ctx) error {
	var hallCreate request.HallCreate
	if err := c.Bind().Body(&hallCreate); err != nil {
		return err
	}
	if err := validation.Validate(&hallCreate); err != nil {
		return err
	}
	hall, err := h.hallService.Create(c, &hallCreate)
	if err != nil {
		return err
	}
	hallResponse := response.NewHallFull(hall.ID, hall.Name, hall.Description, hall.PricePerHour, hall.IsActive)
	return c.Status(fiber.StatusCreated).JSON(hallResponse)
}

func (h *HallHandler) Update(c fiber.Ctx) error {
	var hallUpdate request.HallUpdate
	if err := c.Bind().Body(&hallUpdate); err != nil {
		return err
	}
	if err := validation.Validate(&hallUpdate); err != nil {
		return err
	}
	hall, err := h.hallService.Update(c, &hallUpdate)
	if err != nil {
		return err
	}
	hallResponse := response.NewHallFull(hall.ID, hall.Name, hall.Description, hall.PricePerHour, hall.IsActive)
	return c.Status(fiber.StatusOK).JSON(hallResponse)
}
