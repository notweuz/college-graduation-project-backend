package handler

import (
	"college-graduation-project-backend/internal/errs"
	"college-graduation-project-backend/internal/middleware"
	"college-graduation-project-backend/internal/model/request"
	"college-graduation-project-backend/internal/model/response"
	"college-graduation-project-backend/internal/service"
	"college-graduation-project-backend/internal/validation"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
)

type BookingHandler struct {
	bookingService service.BookingService
}

func NewBookingHandler(bookingService service.BookingService) *BookingHandler {
	return &BookingHandler{bookingService: bookingService}
}

// Create godoc
// @Summary Create booking
// @Description Создает бронирование для текущего пользователя.
// @Tags bookings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payload body request.BookingCreate true "Данные бронирования"
// @Success 201 {object} response.BookingFull
// @Failure 400 {object} BadRequestErrorResponse
// @Failure 401 {object} UnauthorizedErrorResponse
// @Failure 404 {object} NotFoundErrorResponse
// @Failure 409 {object} ConflictErrorResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /api/bookings/ [post]
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
			booking.Hall.PricePerDay,
			booking.Hall.IsActive,
			[]string{},
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

// GetAllFromUser godoc
// @Summary List my bookings
// @Description Возвращает бронирования текущего пользователя с необязательным фильтром по периоду (RFC3339).
// @Tags bookings
// @Produce json
// @Security BearerAuth
// @Param from query string false "Дата/время начала (RFC3339)"
// @Param to query string false "Дата/время конца (RFC3339)"
// @Success 200 {array} response.BookingFull
// @Failure 400 {object} BadRequestErrorResponse
// @Failure 401 {object} UnauthorizedErrorResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /api/bookings/my [get]
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
		hallFull := response.NewHallFull(booking.Hall.ID, booking.Hall.Name, booking.Hall.Description, booking.Hall.PricePerDay, booking.Hall.IsActive, []string{})
		userShort := response.NewUserShort(booking.User.ID, booking.User.FullName, booking.User.Email)
		bookingsFull[i] = response.NewBookingFull(booking.ID, *hallFull, *userShort, booking.StartDateTime, booking.EndDateTime, booking.TotalPrice, booking.Comment, booking.CreatedAt)
	}
	return c.Status(fiber.StatusOK).JSON(bookingsFull)
}

// GetByID godoc
// @Summary Get booking by ID
// @Description Возвращает бронирование по ID для текущего пользователя или администратора.
// @Tags bookings,admin-bookings
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID бронирования"
// @Success 200 {object} response.BookingFull
// @Failure 401 {object} UnauthorizedErrorResponse
// @Failure 403 {object} ForbiddenErrorResponse
// @Failure 404 {object} NotFoundErrorResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /api/bookings/{id} [get]
// @Router /api/admin/bookings/{id} [get]
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
	hallFull := response.NewHallFull(booking.Hall.ID, booking.Hall.Name, booking.Hall.Description, booking.Hall.PricePerDay, booking.Hall.IsActive, []string{})
	userShort := response.NewUserShort(booking.User.ID, booking.User.FullName, booking.User.Email)
	return c.Status(fiber.StatusOK).JSON(response.NewBookingFull(booking.ID, *hallFull, *userShort, booking.StartDateTime, booking.EndDateTime, booking.TotalPrice, booking.Comment, booking.CreatedAt))
}

// DeleteByAuthor godoc
// @Summary Delete booking by author
// @Description Удаляет бронирование, если текущий пользователь является автором.
// @Tags bookings
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID бронирования"
// @Success 204 {string} string "No Content"
// @Failure 401 {object} UnauthorizedErrorResponse
// @Failure 403 {object} ForbiddenErrorResponse
// @Failure 404 {object} NotFoundErrorResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /api/bookings/{id} [delete]
func (h *BookingHandler) DeleteByAuthor(c fiber.Ctx) error {
	userID, err := middleware.GetCurrentUserID(c)
	if err != nil {
		return err
	}
	id := fiber.Params[uint64](c, "id")
	err = h.bookingService.DeleteByAuthor(userID, id)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusNoContent).JSON(nil)
}

// GetAll godoc
// @Summary List bookings (admin)
// @Description Возвращает список бронирований с фильтрами (требуется авторизация администратора).
// @Tags admin-bookings
// @Produce json
// @Security BearerAuth
// @Param from query string false "Дата/время начала (RFC3339)"
// @Param to query string false "Дата/время конца (RFC3339)"
// @Param hall_id query int false "ID зала"
// @Param user_id query int false "ID пользователя"
// @Success 200 {array} response.BookingFull
// @Failure 400 {object} BadRequestErrorResponse
// @Failure 401 {object} UnauthorizedErrorResponse
// @Failure 403 {object} ForbiddenErrorResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /api/admin/bookings/ [get]
func (h *BookingHandler) GetAll(c fiber.Ctx) error {
	userID, err := middleware.GetCurrentUserID(c)
	if err != nil {
		return err
	}
	var hallID, bookingUserID *uint64
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
	if s := c.Query("hall_id"); s != "" {
		hallIDVal, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid 'hall_id' format")
		}
		hallID = &hallIDVal
	}
	if s := c.Query("user_id"); s != "" {
		userIDVal, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid 'user_id' format")
		}
		bookingUserID = &userIDVal
	}

	bookings, err := h.bookingService.FindAll(userID, hallID, bookingUserID, from, to)
	if err != nil {
		return err
	}
	bookingFulls := make([]response.BookingFull, len(bookings))
	for i, booking := range bookings {
		hallFull := response.NewHallFull(booking.Hall.ID, booking.Hall.Name, booking.Hall.Description, booking.Hall.PricePerDay, booking.Hall.IsActive, []string{})
		userShort := response.NewUserShort(booking.User.ID, booking.User.FullName, booking.User.Email)
		bookingFulls[i] = response.NewBookingFull(booking.ID, *hallFull, *userShort, booking.StartDateTime, booking.EndDateTime, booking.TotalPrice, booking.Comment, booking.CreatedAt)
	}
	return c.Status(fiber.StatusOK).JSON(bookingFulls)
}

// Update godoc
// @Summary Update booking (admin)
// @Description Обновляет бронирование по ID (требуется авторизация администратора).
// @Tags admin-bookings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID бронирования"
// @Param payload body request.BookingUpdate true "Поля бронирования для обновления"
// @Success 200 {object} response.BookingFull
// @Failure 400 {object} BadRequestErrorResponse
// @Failure 401 {object} UnauthorizedErrorResponse
// @Failure 403 {object} ForbiddenErrorResponse
// @Failure 404 {object} NotFoundErrorResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /api/admin/bookings/{id} [patch]
func (h *BookingHandler) Update(c fiber.Ctx) error {
	userID, err := middleware.GetCurrentUserID(c)
	if err != nil {
		return err
	}
	id := fiber.Params[uint64](c, "id")
	var req request.BookingUpdate
	if err := c.Bind().Body(&req); err != nil {
		return err
	}
	if err = validation.Validate(&req); err != nil {
		return err
	}
	booking, err := h.bookingService.Update(userID, id, &req)
	if err != nil {
		return err
	}
	userShort := response.NewUserShort(booking.User.ID, booking.User.FullName, booking.User.Email)
	hallFull := response.NewHallFull(booking.Hall.ID, booking.Hall.Name, booking.Hall.Description, booking.Hall.PricePerDay, booking.Hall.IsActive, []string{})
	bookingFull := response.NewBookingFull(booking.ID, *hallFull, *userShort, booking.StartDateTime, booking.EndDateTime, booking.TotalPrice, booking.Comment, booking.CreatedAt)
	return c.Status(fiber.StatusOK).JSON(bookingFull)
}

func (h *BookingHandler) CalculatePrice(c fiber.Ctx) error {
	hallID := fiber.Params[uint64](c, "hall_id")

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

	calculated, err := h.bookingService.CalculatePrice(hallID, from, to)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(calculated)
}
