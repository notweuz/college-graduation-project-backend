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

func parseSalesReportQuery(c fiber.Ctx) (time.Time, time.Time, *uint64, string, []string, error) {
	fromRaw := strings.TrimSpace(c.Query("from"))
	if fromRaw == "" {
		return time.Time{}, time.Time{}, nil, "", nil, errs.BadRequest("Cannot get sales report", "'from' query param is required (RFC3339)")
	}
	from, err := time.Parse(time.RFC3339, fromRaw)
	if err != nil {
		return time.Time{}, time.Time{}, nil, "", nil, errs.BadRequest("Cannot get sales report", "'from' must be RFC3339")
	}

	toRaw := strings.TrimSpace(c.Query("to"))
	if toRaw == "" {
		return time.Time{}, time.Time{}, nil, "", nil, errs.BadRequest("Cannot get sales report", "'to' query param is required (RFC3339)")
	}
	to, err := time.Parse(time.RFC3339, toRaw)
	if err != nil {
		return time.Time{}, time.Time{}, nil, "", nil, errs.BadRequest("Cannot get sales report", "'to' must be RFC3339")
	}

	var hallID *uint64
	if hallRaw := strings.TrimSpace(c.Query("hall_id")); hallRaw != "" {
		value, parseErr := strconv.ParseUint(hallRaw, 10, 64)
		if parseErr != nil {
			return time.Time{}, time.Time{}, nil, "", nil, errs.BadRequest("Cannot get sales report", "'hall_id' must be uint64")
		}
		hallID = &value
	}

	groupBy := strings.TrimSpace(c.Query("group_by", "day"))

	metrics := []string{}
	if metricsRaw := strings.TrimSpace(c.Query("metrics")); metricsRaw != "" {
		metrics = strings.Split(metricsRaw, ",")
	}

	return from, to, hallID, groupBy, metrics, nil
}
