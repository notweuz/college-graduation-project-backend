package handler

import (
	"college-graduation-project-backend/internal/errs"
	"college-graduation-project-backend/internal/middleware"
	"college-graduation-project-backend/internal/model/request"
	"college-graduation-project-backend/internal/model/response"
	"college-graduation-project-backend/internal/service"
	"college-graduation-project-backend/internal/validation"
	"time"

	"github.com/gofiber/fiber/v3"
)

type HallHandler struct {
	hallService service.HallService
}

func NewHallHandler(hallService service.HallService) *HallHandler {
	return &HallHandler{hallService: hallService}
}

func (h *HallHandler) GetAllHalls(c fiber.Ctx) error {
	onlyActive := fiber.Query[bool](c, "active")
	var halls []response.HallFull
	if onlyActive {
		hallsFull, err := h.hallService.FindAllActive()
		if err != nil {
			return err
		}
		for _, hall := range hallsFull {
			halls = append(halls, *response.NewHallFull(hall.ID, hall.Name, hall.Description, hall.PricePerHour, hall.IsActive))
		}
	} else {
		hallsFull, err := h.hallService.FindAll()
		if err != nil {
			return err
		}
		for _, hall := range hallsFull {
			halls = append(halls, *response.NewHallFull(hall.ID, hall.Name, hall.Description, hall.PricePerHour, hall.IsActive))
		}
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
	userID, err := middleware.GetCurrentUserID(c)
	if err != nil {
		return err
	}
	hall, err := h.hallService.Create(userID, &hallCreate)
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
	userID, err := middleware.GetCurrentUserID(c)
	if err != nil {
		return err
	}
	id := fiber.Params[uint64](c, "id")
	hall, err := h.hallService.Update(userID, id, &hallUpdate)
	if err != nil {
		return err
	}
	hallResponse := response.NewHallFull(hall.ID, hall.Name, hall.Description, hall.PricePerHour, hall.IsActive)
	return c.Status(fiber.StatusOK).JSON(hallResponse)
}

func (h *HallHandler) Delete(c fiber.Ctx) error {
	id := fiber.Params[uint64](c, "id")
	userID, err := middleware.GetCurrentUserID(c)
	if err != nil {
		return err
	}
	err = h.hallService.Delete(userID, id)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusNoContent).JSON(nil)
}

func (h *HallHandler) GetHallAvailability(c fiber.Ctx) error {
	id := fiber.Params[uint64](c, "id")

	dateStr := c.Query("date")
	fromStr := c.Query("from")
	toStr := c.Query("to")

	var from, to time.Time
	var err error

	if dateStr != "" {
		from, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			return errs.BadRequest("Invalid 'date' format, use YYYY-MM-DD", err.Error())
		}
		to = from.Add(24 * time.Hour)
	} else if fromStr != "" && toStr != "" {
		from, err = time.Parse("2006-01-02", fromStr)
		if err != nil {
			return errs.BadRequest("Invalid 'from' format, use YYYY-MM-DD", err.Error())
		}
		to, err = time.Parse("2006-01-02", toStr)
		if err != nil {
			return errs.BadRequest("Invalid 'to' format, use YYYY-MM-DD", err.Error())
		}
		to = to.Add(24 * time.Hour)
	} else {
		return errs.BadRequest("Invalid parameters", "provide either 'date' or both 'from' and 'to' parameters")
	}

	_, err = h.hallService.FindByID(id)
	if err != nil {
		return err
	}

	availability, err := h.hallService.GetAvailability(id, from, to)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(availability)
}
