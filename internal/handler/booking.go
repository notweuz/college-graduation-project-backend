package handler

import (
	"college-graduation-project-backend/internal/middleware"
	"college-graduation-project-backend/internal/model/response"
	"college-graduation-project-backend/internal/service"
	"time"

	"github.com/gofiber/fiber/v3"
)

type BookingHandler struct {
	bookingService service.BookingService
}

func NewBookingHandler(bookingService service.BookingService) *BookingHandler {
	return &BookingHandler{bookingService: bookingService}
}

func (h *BookingHandler) GetAllFromUser(c fiber.Ctx) error {
	userID, err := middleware.GetCurrentUserID(c)
	if err != nil {
		return err
	}

	var from, to *time.Time

	if s := c.Query("from"); s != "" {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid 'from' date format, use RFC3339")
		}
		from = &t
	}

	if s := c.Query("to"); s != "" {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid 'to' date format, use RFC3339")
		}
		to = &t
	}

	bookings, err := h.bookingService.FindAllFromUser(userID, from, to)
	if err != nil {
		return err
	}
	bookingsFull := make([]response.BookingFull, len(bookings))
	for i, booking := range bookings {
		bookingsFull[i] = response.NewBookingFull(booking.ID, response.HallFull{}, response.UserShort{}, booking.StartDateTime, booking.EndDateTime, booking.TotalPrice, booking.Comment, booking.CreatedAt)
	}
	return c.Status(fiber.StatusOK).JSON(bookingsFull)
}

func (h *BookingHandler) GetByID(c fiber.Ctx) error {
	userID, err := middleware.GetCurrentUserID(c)
	if err != nil {
		return err
	}
	id := fiber.Params[uint64](c, "id")
	booking, err := h.bookingService.FindByID(userID, id)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusOK).JSON(response.NewBookingFull(booking.ID, response.HallFull{}, response.UserShort{}, booking.StartDateTime, booking.EndDateTime, booking.TotalPrice, booking.Comment, booking.CreatedAt))
}
