package handler

import (
	"college-graduation-project-backend/internal/errs"
	"college-graduation-project-backend/internal/middleware"
	"college-graduation-project-backend/internal/service"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
)

type ReportHandler struct {
	reportService service.ReportService
}

func NewReportHandler(reportService service.ReportService) *ReportHandler {
	return &ReportHandler{reportService: reportService}
}

// GetSalesReport godoc
// @Summary Get sales report (admin)
// @Description Возвращает отчет по продажам в JSON.
// @Tags admin-reports
// @Produce json
// @Security BearerAuth
// @Param from query string true "Дата/время начала (RFC3339)"
// @Param to query string true "Дата/время конца (RFC3339)"
// @Param hall_id query int false "ID зала"
// @Param group_by query string false "Группировка: day|week|month"
// @Param metrics query string false "Метрики через запятую: revenue,bookings_count,avg_check,occupancy"
// @Success 200 {object} response.SalesReport
// @Failure 400 {object} BadRequestErrorResponse
// @Failure 401 {object} UnauthorizedErrorResponse
// @Failure 403 {object} ForbiddenErrorResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /api/admin/reports/sales [get]
func (h *ReportHandler) GetSalesReport(c fiber.Ctx) error {
	userID, err := middleware.GetCurrentUserID(c)
	if err != nil {
		return err
	}

	from, to, hallID, groupBy, metrics, err := parseSalesReportQuery(c)
	if err != nil {
		return err
	}

	report, err := h.reportService.GetSalesReport(userID, from, to, hallID, groupBy, metrics)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(report)
}

// GetSalesReportPDF godoc
// @Summary Get sales report PDF (admin)
// @Description Возвращает PDF отчет по продажам.
// @Tags admin-reports
// @Produce application/pdf
// @Security BearerAuth
// @Param from query string true "Дата/время начала (RFC3339)"
// @Param to query string true "Дата/время конца (RFC3339)"
// @Param hall_id query int false "ID зала"
// @Param group_by query string false "Группировка: day|week|month"
// @Param metrics query string false "Метрики через запятую: revenue,bookings_count,avg_check,occupancy"
// @Success 200 {file} binary
// @Failure 400 {object} BadRequestErrorResponse
// @Failure 401 {object} UnauthorizedErrorResponse
// @Failure 403 {object} ForbiddenErrorResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /api/admin/reports/sales/pdf [get]
func (h *ReportHandler) GetSalesReportPDF(c fiber.Ctx) error {
	userID, err := middleware.GetCurrentUserID(c)
	if err != nil {
		return err
	}

	from, to, hallID, groupBy, metrics, err := parseSalesReportQuery(c)
	if err != nil {
		return err
	}

	pdfBytes, err := h.reportService.GetSalesReportPDF(userID, from, to, hallID, groupBy, metrics)
	if err != nil {
		return err
	}

	filename := "sales-report.pdf"
	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	return c.Send(pdfBytes)
}

// GetHallsLoadReport godoc
// @Summary Get halls load report (admin)
// @Description Возвращает отчет по загрузке залов в JSON.
// @Tags admin-reports
// @Produce json
// @Security BearerAuth
// @Param from query string true "Дата/время начала (RFC3339)"
// @Param to query string true "Дата/время конца (RFC3339)"
// @Param hall_id query int false "ID зала"
// @Success 200 {object} response.HallsLoadReport
// @Failure 400 {object} BadRequestErrorResponse
// @Failure 401 {object} UnauthorizedErrorResponse
// @Failure 403 {object} ForbiddenErrorResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /api/admin/reports/halls-load [get]
func (h *ReportHandler) GetHallsLoadReport(c fiber.Ctx) error {
	userID, err := middleware.GetCurrentUserID(c)
	if err != nil {
		return err
	}

	from, to, hallID, err := parseReportPeriodAndHall(c)
	if err != nil {
		return err
	}

	report, err := h.reportService.GetHallsLoadReport(userID, from, to, hallID)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(report)
}

// GetHallsLoadReportPDF godoc
// @Summary Get halls load report PDF (admin)
// @Description Возвращает PDF отчет по загрузке залов.
// @Tags admin-reports
// @Produce application/pdf
// @Security BearerAuth
// @Param from query string true "Дата/время начала (RFC3339)"
// @Param to query string true "Дата/время конца (RFC3339)"
// @Param hall_id query int false "ID зала"
// @Success 200 {file} binary
// @Failure 400 {object} BadRequestErrorResponse
// @Failure 401 {object} UnauthorizedErrorResponse
// @Failure 403 {object} ForbiddenErrorResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /api/admin/reports/halls-load/pdf [get]
func (h *ReportHandler) GetHallsLoadReportPDF(c fiber.Ctx) error {
	userID, err := middleware.GetCurrentUserID(c)
	if err != nil {
		return err
	}

	from, to, hallID, err := parseReportPeriodAndHall(c)
	if err != nil {
		return err
	}

	pdfBytes, err := h.reportService.GetHallsLoadReportPDF(userID, from, to, hallID)
	if err != nil {
		return err
	}

	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", "attachment; filename=\"halls-load-report.pdf\"")
	return c.Send(pdfBytes)
}

// GetClientsReport godoc
// @Summary Get clients report (admin)
// @Description Возвращает отчет по клиентам в JSON.
// @Tags admin-reports
// @Produce json
// @Security BearerAuth
// @Param from query string true "Дата/время начала (RFC3339)"
// @Param to query string true "Дата/время конца (RFC3339)"
// @Param hall_id query int false "ID зала"
// @Param limit query int false "Ограничение числа строк (по умолчанию 20)"
// @Success 200 {object} response.ClientsReport
// @Failure 400 {object} BadRequestErrorResponse
// @Failure 401 {object} UnauthorizedErrorResponse
// @Failure 403 {object} ForbiddenErrorResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /api/admin/reports/clients [get]
func (h *ReportHandler) GetClientsReport(c fiber.Ctx) error {
	userID, err := middleware.GetCurrentUserID(c)
	if err != nil {
		return err
	}

	from, to, hallID, err := parseReportPeriodAndHall(c)
	if err != nil {
		return err
	}

	limit := uint64(20)
	if limitRaw := strings.TrimSpace(c.Query("limit")); limitRaw != "" {
		value, parseErr := strconv.ParseUint(limitRaw, 10, 64)
		if parseErr != nil {
			return errs.BadRequest("Cannot get clients report", "'limit' must be uint64")
		}
		limit = value
	}

	report, err := h.reportService.GetClientsReport(userID, from, to, hallID, limit)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(report)
}

// GetClientsReportPDF godoc
// @Summary Get clients report PDF (admin)
// @Description Возвращает PDF отчет по клиентам.
// @Tags admin-reports
// @Produce application/pdf
// @Security BearerAuth
// @Param from query string true "Дата/время начала (RFC3339)"
// @Param to query string true "Дата/время конца (RFC3339)"
// @Param hall_id query int false "ID зала"
// @Param limit query int false "Ограничение числа строк (по умолчанию 20)"
// @Success 200 {file} binary
// @Failure 400 {object} BadRequestErrorResponse
// @Failure 401 {object} UnauthorizedErrorResponse
// @Failure 403 {object} ForbiddenErrorResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /api/admin/reports/clients/pdf [get]
func (h *ReportHandler) GetClientsReportPDF(c fiber.Ctx) error {
	userID, err := middleware.GetCurrentUserID(c)
	if err != nil {
		return err
	}

	from, to, hallID, err := parseReportPeriodAndHall(c)
	if err != nil {
		return err
	}

	limit := uint64(20)
	if limitRaw := strings.TrimSpace(c.Query("limit")); limitRaw != "" {
		value, parseErr := strconv.ParseUint(limitRaw, 10, 64)
		if parseErr != nil {
			return errs.BadRequest("Cannot get clients report", "'limit' must be uint64")
		}
		limit = value
	}

	pdfBytes, err := h.reportService.GetClientsReportPDF(userID, from, to, hallID, limit)
	if err != nil {
		return err
	}

	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", "attachment; filename=\"clients-report.pdf\"")
	return c.Send(pdfBytes)
}

// GetBookingsDynamicsReport godoc
// @Summary Get bookings dynamics report (admin)
// @Description Возвращает отчет по динамике бронирований в JSON.
// @Tags admin-reports
// @Produce json
// @Security BearerAuth
// @Param from query string true "Дата/время начала (RFC3339)"
// @Param to query string true "Дата/время конца (RFC3339)"
// @Param hall_id query int false "ID зала"
// @Param group_by query string false "Группировка: day|week|month"
// @Success 200 {object} response.BookingsDynamicsReport
// @Failure 400 {object} BadRequestErrorResponse
// @Failure 401 {object} UnauthorizedErrorResponse
// @Failure 403 {object} ForbiddenErrorResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /api/admin/reports/bookings-dynamics [get]
func (h *ReportHandler) GetBookingsDynamicsReport(c fiber.Ctx) error {
	userID, err := middleware.GetCurrentUserID(c)
	if err != nil {
		return err
	}

	from, to, hallID, err := parseReportPeriodAndHall(c)
	if err != nil {
		return err
	}

	groupBy := strings.TrimSpace(c.Query("group_by", "day"))
	report, err := h.reportService.GetBookingsDynamicsReport(userID, from, to, hallID, groupBy)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(report)
}

// GetBookingsDynamicsReportPDF godoc
// @Summary Get bookings dynamics report PDF (admin)
// @Description Возвращает PDF отчет по динамике бронирований.
// @Tags admin-reports
// @Produce application/pdf
// @Security BearerAuth
// @Param from query string true "Дата/время начала (RFC3339)"
// @Param to query string true "Дата/время конца (RFC3339)"
// @Param hall_id query int false "ID зала"
// @Param group_by query string false "Группировка: day|week|month"
// @Success 200 {file} binary
// @Failure 400 {object} BadRequestErrorResponse
// @Failure 401 {object} UnauthorizedErrorResponse
// @Failure 403 {object} ForbiddenErrorResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /api/admin/reports/bookings-dynamics/pdf [get]
func (h *ReportHandler) GetBookingsDynamicsReportPDF(c fiber.Ctx) error {
	userID, err := middleware.GetCurrentUserID(c)
	if err != nil {
		return err
	}

	from, to, hallID, err := parseReportPeriodAndHall(c)
	if err != nil {
		return err
	}

	groupBy := strings.TrimSpace(c.Query("group_by", "day"))
	pdfBytes, err := h.reportService.GetBookingsDynamicsReportPDF(userID, from, to, hallID, groupBy)
	if err != nil {
		return err
	}

	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", "attachment; filename=\"bookings-dynamics-report.pdf\"")
	return c.Send(pdfBytes)
}

func parseSalesReportQuery(c fiber.Ctx) (time.Time, time.Time, *uint64, string, []string, error) {
	from, to, hallID, err := parseReportPeriodAndHall(c)
	if err != nil {
		return time.Time{}, time.Time{}, nil, "", nil, err
	}

	groupBy := strings.TrimSpace(c.Query("group_by", "day"))

	metrics := []string{}
	if metricsRaw := strings.TrimSpace(c.Query("metrics")); metricsRaw != "" {
		metrics = strings.Split(metricsRaw, ",")
	}

	return from, to, hallID, groupBy, metrics, nil
}

func parseReportPeriodAndHall(c fiber.Ctx) (time.Time, time.Time, *uint64, error) {
	fromRaw := strings.TrimSpace(c.Query("from"))
	if fromRaw == "" {
		return time.Time{}, time.Time{}, nil, errs.BadRequest("Cannot get report", "'from' query param is required (RFC3339)")
	}
	from, err := time.Parse(time.RFC3339, fromRaw)
	if err != nil {
		return time.Time{}, time.Time{}, nil, errs.BadRequest("Cannot get report", "'from' must be RFC3339")
	}

	toRaw := strings.TrimSpace(c.Query("to"))
	if toRaw == "" {
		return time.Time{}, time.Time{}, nil, errs.BadRequest("Cannot get report", "'to' query param is required (RFC3339)")
	}
	to, err := time.Parse(time.RFC3339, toRaw)
	if err != nil {
		return time.Time{}, time.Time{}, nil, errs.BadRequest("Cannot get report", "'to' must be RFC3339")
	}

	var hallID *uint64
	if hallRaw := strings.TrimSpace(c.Query("hall_id")); hallRaw != "" {
		value, parseErr := strconv.ParseUint(hallRaw, 10, 64)
		if parseErr != nil {
			return time.Time{}, time.Time{}, nil, errs.BadRequest("Cannot get report", "'hall_id' must be uint64")
		}
		hallID = &value
	}

	return from, to, hallID, nil
}
