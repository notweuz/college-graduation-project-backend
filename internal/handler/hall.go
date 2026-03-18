package handler

import (
	"college-graduation-project-backend/internal/errs"
	"college-graduation-project-backend/internal/middleware"
	"college-graduation-project-backend/internal/model/request"
	"college-graduation-project-backend/internal/model/response"
	"college-graduation-project-backend/internal/service"
	"college-graduation-project-backend/internal/validation"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
			images, err := h.hallService.GetHallImages(hall.ID)
			if err != nil {
				return err
			}
			halls = append(halls, *response.NewHallFull(hall.ID, hall.Name, hall.Description, hall.PricePerHour, hall.IsActive, images))
		}
	} else {
		hallsFull, err := h.hallService.FindAll()
		if err != nil {
			return err
		}
		for _, hall := range hallsFull {
			images, err := h.hallService.GetHallImages(hall.ID)
			if err != nil {
				return err
			}
			halls = append(halls, *response.NewHallFull(hall.ID, hall.Name, hall.Description, hall.PricePerHour, hall.IsActive, images))
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

	images, err := h.hallService.GetHallImages(hall.ID)
	if err != nil {
		return err
	}

	hallResponse := response.NewHallFull(hall.ID, hall.Name, hall.Description, hall.PricePerHour, hall.IsActive, images)
	return c.Status(fiber.StatusOK).JSON(hallResponse)
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

	images, err := h.hallService.GetHallImages(hall.ID)
	if err != nil {
		return err
	}

	hallResponse := response.NewHallFull(hall.ID, hall.Name, hall.Description, hall.PricePerHour, hall.IsActive, images)
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

	images, err := h.hallService.GetHallImages(hall.ID)
	if err != nil {
		return err
	}

	hallResponse := response.NewHallFull(hall.ID, hall.Name, hall.Description, hall.PricePerHour, hall.IsActive, images)
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

func (h *HallHandler) UploadImage(c fiber.Ctx) error {
	id := fiber.Params[uint64](c, "id")

	userID, err := middleware.GetCurrentUserID(c)
	if err != nil {
		return err
	}

	file, err := c.FormFile("image")
	if err != nil {
		return errs.BadRequest("Invalid file", "No file uploaded")
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".gif" && ext != ".webp" {
		return errs.BadRequest("Invalid file type", "Only jpg, jpeg, png, gif, webp files are allowed")
	}

	if file.Size > 10*1024*1024 {
		return errs.BadRequest("File too large", "Maximum file size is 10MB")
	}

	uploadsDir := "uploads/halls"
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		return errs.InternalServerError("Cannot create upload directory", err.Error())
	}

	filename := fmt.Sprintf("%d_%d%s", id, time.Now().Unix(), ext)
	filePath := filepath.Join(uploadsDir, filename)

	if err := c.SaveFile(file, filePath); err != nil {
		return errs.InternalServerError("Cannot save file", err.Error())
	}

	imagePath := fmt.Sprintf("/api/images/%s", filename)
	err = h.hallService.UploadImage(userID, id, imagePath)
	if err != nil {
		os.Remove(filePath)
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"image_path": imagePath,
	})
}

func (h *HallHandler) ServeImage(c fiber.Ctx) error {
	filename := c.Params("filename")
	filePath := filepath.Join("uploads/halls", filename)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return errs.NotFound("Image not found", "Image file does not exist")
	}

	return c.SendFile(filePath)
}
