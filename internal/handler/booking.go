package handler

import (
	"college-graduation-project-backend/internal/middleware"
	"college-graduation-project-backend/internal/model/request"
	"college-graduation-project-backend/internal/model/response"
	"college-graduation-project-backend/internal/service"
	"college-graduation-project-backend/internal/validation"
	"time"

	"github.com/gofiber/fiber/v3"
)

type BookingHandler struct {
	bookingService service.BookingService
}

func NewBookingHandler(bookingService service.BookingService) *BookingHandler {
	return &BookingHandler{bookingService: bookingService}
}

func (h *BookingHandler) Create(c fiber.Ctx) error {
	userID, err := middleware.GetCurrentUserID(c)
	if err != nil {
		return err
	}

	var req request.BookingCreate
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	if err := validation.Validate(&req); err != nil {
		return err
	}

	booking, err := h.bookingService.Create(userID, &req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(response.NewBookingFull(
		booking.ID,
		*response.NewHallFull(
			booking.Hall.ID,
			booking.Hall.Name,
			booking.Hall.Description,
			booking.Hall.PricePerHour,
			booking.Hall.IsActive,
		),
		*response.NewUserShort(
			booking.User.ID,
			booking.User.FullName,
			booking.User.Email,
		),
		booking.StartDateTime,
		booking.EndDateTime,
		booking.TotalPrice,
		booking.Comment,
		booking.CreatedAt,
	))
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
		hallFull := response.NewHallFull(booking.Hall.ID, booking.Hall.Name, booking.Hall.Description, booking.Hall.PricePerHour, booking.Hall.IsActive)
		userShort := response.NewUserShort(booking.User.ID, booking.User.FullName, booking.User.Email)
		bookingsFull[i] = response.NewBookingFull(booking.ID, *hallFull, *userShort, booking.StartDateTime, booking.EndDateTime, booking.TotalPrice, booking.Comment, booking.CreatedAt)
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
