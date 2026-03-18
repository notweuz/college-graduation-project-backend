package service

import (
	"college-graduation-project-backend/internal/database"
	"college-graduation-project-backend/internal/errs"
	"college-graduation-project-backend/internal/model"
	"college-graduation-project-backend/internal/model/enum"
	"college-graduation-project-backend/internal/model/response"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/johnfercher/maroto/pkg/color"
	"github.com/johnfercher/maroto/pkg/consts"
	"github.com/johnfercher/maroto/pkg/pdf"
	"github.com/johnfercher/maroto/pkg/props"
)

const studioName = "Фотостудия «Культ»"

type reportService struct {
	reportDatabase database.ReportDatabase
	userService    UserService
}

type salesAggregation struct {
	bucketStart   time.Time
	bucketEnd     time.Time
	hallID        *uint64
	hallName      *string
	bookedHours   float64
	revenue       float64
	bookingsCount uint64
}

func NewReportService(reportDatabase database.ReportDatabase, userService UserService) ReportService {
	return &reportService{
		reportDatabase: reportDatabase,
		userService:    userService,
	}
}

func (s *reportService) GetSalesReport(userID uint64, from, to time.Time, hallID *uint64, groupBy string, metrics []string) (*response.SalesReport, error) {
	user, err := s.userService.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if user.Role != enum.RoleAdmin {
		return nil, errs.Forbidden("Cannot get sales report", "You do not have permission to access admin reports")
	}
	if !from.Before(to) {
		return nil, errs.BadRequest("Cannot get sales report", "'from' must be before 'to'")
	}

	groupBy = normalizeGroupBy(groupBy)
	if groupBy == "" {
		return nil, errs.BadRequest("Cannot get sales report", "group_by must be one of: day, week, month, hall")
	}

	metrics, err = normalizeMetrics(metrics)
	if err != nil {
		return nil, err
	}

	hallCount, err := s.reportDatabase.CountHalls(hallID)
	if err != nil {
		return nil, errs.InternalServerError("Cannot get sales report", err.Error())
	}
	if hallCount == 0 {
		if hallID != nil {
			return nil, errs.NotFound("Cannot get sales report", "hall not found")
		}
		return nil, errs.BadRequest("Cannot get sales report", "no halls found for report")
	}

	bookings, err := s.reportDatabase.FindSalesBookings(from, to, hallID)
	if err != nil {
		return nil, errs.InternalServerError("Cannot get sales report", err.Error())
	}

	aggregations := s.aggregateSales(bookings, from, to, groupBy)
	rows, totals, summary := buildSalesResponse(aggregations, from, to, groupBy, metrics, hallCount)

	report := &response.SalesReport{
		StudioName:  studioName,
		Title:       "Административный отчет о продажах",
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		From:        from.Format(time.RFC3339),
		To:          to.Format(time.RFC3339),
		GroupBy:     groupBy,
		Metrics:     metrics,
		Filters: response.SalesReportFilters{
			From:    from.Format(time.RFC3339),
			To:      to.Format(time.RFC3339),
			HallID:  hallID,
			GroupBy: groupBy,
			Metrics: metrics,
		},
		Rows:    rows,
		Totals:  totals,
		Summary: summary,
	}

	return report, nil
}

func (s *reportService) GetSalesReportPDF(userID uint64, from, to time.Time, hallID *uint64, groupBy string, metrics []string) ([]byte, error) {
	report, err := s.GetSalesReport(userID, from, to, hallID, groupBy, metrics)
	if err != nil {
		return nil, err
	}

	m := pdf.NewMaroto(consts.Portrait, consts.A4)
	if err = configureUTF8Font(m); err != nil {
		return nil, errs.InternalServerError("Cannot generate sales report PDF", err.Error())
	}
	m.SetPageMargins(10, 12, 10)

	dark := color.Color{Red: 28, Green: 45, Blue: 68}
	light := color.Color{Red: 244, Green: 247, Blue: 252}
	white := color.Color{Red: 255, Green: 255, Blue: 255}
	muted := color.Color{Red: 112, Green: 122, Blue: 136}

	m.SetBackgroundColor(dark)
	m.Row(9, func() {
		m.Col(12, func() {
			m.Text(report.StudioName, props.Text{
				Top:   2,
				Style: consts.Bold,
				Size:  15,
				Align: consts.Center,
				Color: white,
			})
		})
	})
	m.SetBackgroundColor(white)

	m.Row(8, func() {
		m.Col(12, func() {
			m.Text(report.Title, props.Text{Top: 1, Style: consts.Bold, Size: 12, Align: consts.Center, Color: dark})
		})
	})
	m.Line(1, props.Line{Color: dark, Width: 0.4})

	m.Row(6, func() {
		m.Col(12, func() {
			m.Text(fmt.Sprintf("Период: %s - %s", report.From, report.To), props.Text{Size: 9, Color: dark})
		})
	})
	m.Row(6, func() {
		m.Col(12, func() {
			m.Text(fmt.Sprintf("Сгруппировано по: %s | Метрики: %s", report.GroupBy, strings.Join(report.Metrics, ", ")), props.Text{Size: 9, Color: dark})
		})
	})
	m.Row(6, func() {
		m.Col(12, func() {
			m.Text(fmt.Sprintf("Сформирован: %s", report.GeneratedAt), props.Text{Size: 9, Color: dark})
		})
	})

	m.SetBackgroundColor(light)
	m.Row(8, func() {
		m.Col(12, func() {
			m.Text("Итоги периода", props.Text{Top: 1.5, Style: consts.Bold, Size: 11, Align: consts.Center, Color: dark})
		})
	})
	m.SetBackgroundColor(white)

	summaryRows := [][]string{
		{
			"Выручка: " + formatMetric("revenue", report.Totals),
			"Бронирований: " + formatMetric("bookings_count", report.Totals),
			"Средний чек: " + formatMetric("avg_check", report.Totals),
		},
		{
			fmt.Sprintf("Загружено часов: %.2f", report.Summary.TotalBookedHours),
			fmt.Sprintf("Доступно часов: %.2f", report.Summary.AvailableHours),
			"Загрузка: " + formatMetric("occupancy", report.Totals),
		},
		{
			fmt.Sprintf("Строк отчета: %d", report.Summary.RowsCount),
			fmt.Sprintf("Залов в отчете: %d", report.Summary.HallsCount),
			fmt.Sprintf("Интервал: %.1f ч", to.Sub(from).Hours()),
		},
	}

	for _, row := range summaryRows {
		values := row
		m.Row(6, func() {
			for _, value := range values {
				text := value
				m.Col(4, func() {
					m.Text(text, props.Text{Size: 8.7, Color: dark})
				})
			}
		})
	}
	m.Line(1, props.Line{Color: dark, Width: 0.2})

	headers, gridSizes := buildTableHeader(report.GroupBy, report.Metrics)
	contents := buildTableContent(report.Rows, report.GroupBy, report.Metrics)

	m.SetBackgroundColor(dark)
	m.TableList(headers, contents, props.TableList{
		HeaderProp: props.TableListContent{
			Style:     consts.Bold,
			Size:      9,
			Color:     white,
			GridSizes: gridSizes,
		},
		ContentProp: props.TableListContent{
			Size:      8,
			Color:     dark,
			GridSizes: gridSizes,
		},
		Align:                  consts.Center,
		AlternatedBackground:   &light,
		HeaderContentSpace:     2,
		VerticalContentPadding: 1,
		Line:                   true,
		LineProp:               props.Line{Color: color.Color{Red: 215, Green: 223, Blue: 236}, Width: 0.15},
	})
	m.SetBackgroundColor(white)

	m.Row(6, func() {
		m.Col(12, func() {
			m.Text("Документ сгенерирован автоматически системой администрирования "+report.StudioName, props.Text{
				Size:  7.5,
				Align: consts.Right,
				Color: muted,
			})
		})
	})

	out, err := m.Output()
	if err != nil {
		return nil, errs.InternalServerError("Cannot generate sales report PDF", err.Error())
	}

	return out.Bytes(), nil
}

func (s *reportService) aggregateSales(bookings []model.Booking, from, to time.Time, groupBy string) map[string]*salesAggregation {
	aggregations := map[string]*salesAggregation{}

	for _, booking := range bookings {
		clampedStart := maxTime(booking.StartDateTime, from)
		clampedEnd := minTime(booking.EndDateTime, to)
		if !clampedStart.Before(clampedEnd) {
			continue
		}

		bookingDuration := booking.EndDateTime.Sub(booking.StartDateTime).Hours()
		if bookingDuration <= 0 {
			continue
		}

		if groupBy == "hall" {
			key := fmt.Sprintf("hall:%d", booking.HallID)
			agg := getAggregation(aggregations, key, from, to, &booking.HallID, &booking.Hall.Name)
			agg.bookedHours += clampedEnd.Sub(clampedStart).Hours()
			agg.revenue += booking.TotalPrice * (clampedEnd.Sub(clampedStart).Hours() / bookingDuration)
			agg.bookingsCount++
			continue
		}

		for segmentStart := clampedStart; segmentStart.Before(clampedEnd); {
			bucketStart := truncateToBucket(segmentStart, groupBy)
			bucketEnd := nextBucketStart(bucketStart, groupBy)
			segmentEnd := minTime(clampedEnd, bucketEnd)
			segmentHours := segmentEnd.Sub(segmentStart).Hours()

			key := bucketKey(bucketStart, groupBy)
			agg := getAggregation(aggregations, key, maxTime(bucketStart, from), minTime(bucketEnd, to), nil, nil)
			agg.bookedHours += segmentHours
			agg.revenue += booking.TotalPrice * (segmentHours / bookingDuration)
			agg.bookingsCount++

			segmentStart = segmentEnd
		}
	}

	return aggregations
}

func buildSalesResponse(aggregations map[string]*salesAggregation, from, to time.Time, groupBy string, metrics []string, hallsCount uint64) ([]response.SalesReportRow, response.SalesReportMetrics, response.SalesReportSummary) {
	keys := make([]string, 0, len(aggregations))
	for key := range aggregations {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	rows := make([]response.SalesReportRow, 0, len(keys))
	var totalRevenue float64
	var totalBookedHours float64
	var totalBookings uint64

	for _, key := range keys {
		agg := aggregations[key]
		totalRevenue += agg.revenue
		totalBookedHours += agg.bookedHours
		totalBookings += agg.bookingsCount

		availableHours := agg.bucketEnd.Sub(agg.bucketStart).Hours()
		if groupBy != "hall" {
			availableHours *= float64(hallsCount)
		}
		metricsValue := selectMetrics(metrics, agg.revenue, agg.bookingsCount, agg.bookedHours, availableHours)

		row := response.SalesReportRow{
			Bucket:      key,
			HallID:      agg.hallID,
			HallName:    agg.hallName,
			DateFrom:    agg.bucketStart.Format(time.RFC3339),
			DateTo:      agg.bucketEnd.Format(time.RFC3339),
			BookedHours: round2(agg.bookedHours),
			Metrics:     metricsValue,
		}
		rows = append(rows, row)
	}

	totalAvailableHours := to.Sub(from).Hours()
	if groupBy != "hall" {
		totalAvailableHours *= float64(hallsCount)
	}
	totals := selectMetrics(metrics, totalRevenue, totalBookings, totalBookedHours, totalAvailableHours)
	summary := response.SalesReportSummary{
		RowsCount:        uint64(len(rows)),
		HallsCount:       hallsCount,
		TotalBookedHours: round2(totalBookedHours),
		AvailableHours:   round2(totalAvailableHours),
	}

	return rows, totals, summary
}

func getAggregation(store map[string]*salesAggregation, key string, from, to time.Time, hallID *uint64, hallName *string) *salesAggregation {
	if agg, ok := store[key]; ok {
		return agg
	}

	store[key] = &salesAggregation{
		bucketStart: from,
		bucketEnd:   to,
		hallID:      hallID,
		hallName:    hallName,
	}
	return store[key]
}

func normalizeGroupBy(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "day":
		return "day"
	case "week":
		return "week"
	case "month":
		return "month"
	case "hall":
		return "hall"
	default:
		return ""
	}
}

func normalizeMetrics(metrics []string) ([]string, error) {
	if len(metrics) == 0 {
		return []string{"revenue", "bookings_count", "avg_check", "occupancy"}, nil
	}

	allowed := map[string]struct{}{
		"revenue":        {},
		"bookings_count": {},
		"avg_check":      {},
		"occupancy":      {},
	}

	result := make([]string, 0, len(metrics))
	seen := make(map[string]struct{}, len(metrics))
	for _, metric := range metrics {
		name := strings.TrimSpace(strings.ToLower(metric))
		if _, ok := allowed[name]; !ok {
			return nil, errs.BadRequest("Cannot get sales report", "metrics must include only: revenue, bookings_count, avg_check, occupancy")
		}
		if _, duplicated := seen[name]; duplicated {
			continue
		}
		seen[name] = struct{}{}
		result = append(result, name)
	}

	return result, nil
}

func selectMetrics(selected []string, revenue float64, bookingsCount uint64, bookedHours, availableHours float64) response.SalesReportMetrics {
	m := response.SalesReportMetrics{}
	for _, metric := range selected {
		switch metric {
		case "revenue":
			value := round2(revenue)
			m.Revenue = &value
		case "bookings_count":
			value := bookingsCount
			m.BookingsCount = &value
		case "avg_check":
			var avg float64
			if bookingsCount > 0 {
				avg = revenue / float64(bookingsCount)
			}
			avg = round2(avg)
			m.AvgCheck = &avg
		case "occupancy":
			var occupancy float64
			if availableHours > 0 {
				occupancy = (bookedHours / availableHours) * 100
			}
			occupancy = round2(occupancy)
			m.Occupancy = &occupancy
		}
	}
	return m
}

func formatMetric(metric string, value response.SalesReportMetrics) string {
	switch metric {
	case "revenue":
		if value.Revenue == nil {
			return "-"
		}
		return fmt.Sprintf("%.2f", *value.Revenue)
	case "bookings_count":
		if value.BookingsCount == nil {
			return "-"
		}
		return fmt.Sprintf("%d", *value.BookingsCount)
	case "avg_check":
		if value.AvgCheck == nil {
			return "-"
		}
		return fmt.Sprintf("%.2f", *value.AvgCheck)
	case "occupancy":
		if value.Occupancy == nil {
			return "-"
		}
		return fmt.Sprintf("%.2f%%", *value.Occupancy)
	default:
		return "-"
	}
}

func buildTableHeader(groupBy string, metrics []string) ([]string, []uint) {
	headers := []string{"Период"}
	if groupBy == "hall" {
		headers = append(headers, "Зал")
	}
	for _, metric := range metrics {
		switch metric {
		case "revenue":
			headers = append(headers, "Выручка")
		case "bookings_count":
			headers = append(headers, "Кол-во брон.")
		case "avg_check":
			headers = append(headers, "Средний чек")
		case "occupancy":
			headers = append(headers, "Загрузка")
		}
	}
	return headers, buildGridSizes(len(headers))
}

func buildGridSizes(cols int) []uint {
	if cols <= 0 {
		return []uint{12}
	}
	if cols == 1 {
		return []uint{12}
	}

	first := uint(4)
	if cols >= 6 {
		first = 3
	}

	restCols := cols - 1
	restTotal := int(12 - first)
	base := restTotal / restCols
	rem := restTotal % restCols

	grid := make([]uint, 0, cols)
	grid = append(grid, first)
	for i := 0; i < restCols; i++ {
		size := base
		if i < rem {
			size++
		}
		if size < 1 {
			size = 1
		}
		grid = append(grid, uint(size))
	}

	return grid
}

func buildTableContent(rows []response.SalesReportRow, groupBy string, metrics []string) [][]string {
	contents := make([][]string, 0, len(rows))
	for _, row := range rows {
		record := []string{row.Bucket}
		if groupBy == "hall" {
			hallName := "-"
			if row.HallName != nil {
				hallName = *row.HallName
			}
			record = append(record, hallName)
		}
		for _, metric := range metrics {
			record = append(record, formatMetric(metric, row.Metrics))
		}
		contents = append(contents, record)
	}
	return contents
}

func bucketKey(bucketStart time.Time, groupBy string) string {
	switch groupBy {
	case "day":
		return bucketStart.Format("2006-01-02")
	case "week":
		year, week := bucketStart.ISOWeek()
		return fmt.Sprintf("%d-W%02d", year, week)
	case "month":
		return bucketStart.Format("2006-01")
	default:
		return bucketStart.Format(time.RFC3339)
	}
}

func truncateToBucket(t time.Time, groupBy string) time.Time {
	switch groupBy {
	case "day":
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	case "week":
		weekday := int(t.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		dayStart := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
		return dayStart.AddDate(0, 0, -(weekday - 1))
	case "month":
		return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
	default:
		return t
	}
}

func nextBucketStart(bucketStart time.Time, groupBy string) time.Time {
	switch groupBy {
	case "day":
		return bucketStart.AddDate(0, 0, 1)
	case "week":
		return bucketStart.AddDate(0, 0, 7)
	case "month":
		return bucketStart.AddDate(0, 1, 0)
	default:
		return bucketStart
	}
}

func maxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

func minTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

func configureUTF8Font(m pdf.Maroto) error {
	fontCandidates := []string{
		"assets/fonts/Arial.ttf",
		"assets/fonts/DejaVuSans.ttf",
		"assets/fonts/NotoSans-Regular.ttf",
		"assets/fonts/ArialUnicode.ttf",
		"./assets/fonts/Arial.ttf",
		"./assets/fonts/DejaVuSans.ttf",
		"./assets/fonts/NotoSans-Regular.ttf",
		"./assets/fonts/ArialUnicode.ttf",
	}

	for _, fontPath := range fontCandidates {
		if _, err := os.Stat(fontPath); err != nil {
			continue
		}

		m.AddUTF8Font("studio_utf8", consts.Normal, fontPath)
		m.AddUTF8Font("studio_utf8", consts.Bold, fontPath)
		m.AddUTF8Font("studio_utf8", consts.Italic, fontPath)
		m.AddUTF8Font("studio_utf8", consts.BoldItalic, fontPath)
		m.SetDefaultFontFamily("studio_utf8")
		return nil
	}

	return fmt.Errorf("utf-8 font for russian text not found; add one of: assets/fonts/Arial.ttf, assets/fonts/DejaVuSans.ttf, assets/fonts/NotoSans-Regular.ttf")
}
