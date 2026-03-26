package service

import (
	"bytes"
	"college-graduation-project-backend/internal/database"
	"college-graduation-project-backend/internal/datetime"
	"college-graduation-project-backend/internal/errs"
	"college-graduation-project-backend/internal/model"
	"college-graduation-project-backend/internal/model/enum"
	"college-graduation-project-backend/internal/model/response"
	"encoding/base64"
	"fmt"
	"image"
	stdcolor "image/color"
	imagedraw "image/draw"
	"image/png"
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
	bookedDays    float64
	revenue       float64
	bookingsCount uint64
}

type hallLoadAggregation struct {
	hallID        uint64
	hallName      string
	bookingsCount uint64
	revenue       float64
	bookedDays    float64
}

type clientAggregation struct {
	userID        uint64
	fullName      *string
	email         *string
	bookingsCount uint64
	totalSpent    float64
	lastBookingAt time.Time
}

type dynamicAggregation struct {
	bucketStart   time.Time
	bucketEnd     time.Time
	bookingsCount uint64
	revenue       float64
	bookedDays    float64
}

func NewReportService(reportDatabase database.ReportDatabase, userService UserService) ReportService {
	return &reportService{
		reportDatabase: reportDatabase,
		userService:    userService,
	}
}

func (s *reportService) GetSalesReport(userID uint64, from, to time.Time, hallID *uint64, groupBy string, metrics []string) (*response.SalesReport, error) {
	if err := s.ensureAdmin(userID); err != nil {
		return nil, err
	}
	if !from.Before(to) {
		return nil, errs.BadRequest("Cannot get sales report", "'from' must be before 'to'")
	}

	groupBy = normalizeGroupBy(groupBy)
	if groupBy == "" {
		return nil, errs.BadRequest("Cannot get sales report", "group_by must be one of: day, week, month, hall")
	}

	metrics, err := normalizeMetrics(metrics)
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
	uniqueBookingsCount := countOverlappingBookings(bookings, from, to)
	rows, totals, summary := buildSalesResponse(aggregations, from, to, groupBy, metrics, hallCount, uniqueBookingsCount)

	report := &response.SalesReport{
		StudioName:  studioName,
		Title:       "Административный отчет о продажах",
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		From:        datetime.Format(from),
		To:          datetime.Format(to),
		GroupBy:     groupBy,
		Metrics:     metrics,
		Filters: response.SalesReportFilters{
			From:    datetime.Format(from),
			To:      datetime.Format(to),
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

func (s *reportService) GetHallsLoadReport(userID uint64, from, to time.Time, hallID *uint64) (*response.HallsLoadReport, error) {
	if err := s.ensureAdmin(userID); err != nil {
		return nil, err
	}
	if !from.Before(to) {
		return nil, errs.BadRequest("Cannot get halls load report", "'from' must be before 'to'")
	}

	hallCount, err := s.reportDatabase.CountHalls(hallID)
	if err != nil {
		return nil, errs.InternalServerError("Cannot get halls load report", err.Error())
	}
	if hallCount == 0 {
		if hallID != nil {
			return nil, errs.NotFound("Cannot get halls load report", "hall not found")
		}
		return nil, errs.BadRequest("Cannot get halls load report", "no halls found for report")
	}

	bookings, err := s.reportDatabase.FindReportBookings(from, to, hallID)
	if err != nil {
		return nil, errs.InternalServerError("Cannot get halls load report", err.Error())
	}

	aggregated := map[uint64]*hallLoadAggregation{}
	for _, booking := range bookings {
		clampedStart := maxTime(booking.StartDateTime, from)
		clampedEnd := minTime(booking.EndDateTime, to)
		if !clampedStart.Before(clampedEnd) {
			continue
		}
		durationDays := billableDaysBetween(booking.StartDateTime, booking.EndDateTime)
		if durationDays <= 0 {
			continue
		}

		agg, ok := aggregated[booking.HallID]
		if !ok {
			agg = &hallLoadAggregation{
				hallID:   booking.HallID,
				hallName: booking.Hall.Name,
			}
			aggregated[booking.HallID] = agg
		}

		overlapDays := billableDaysBetween(clampedStart, clampedEnd)
		agg.bookingsCount++
		agg.bookedDays += overlapDays
		agg.revenue += booking.TotalPrice * (overlapDays / durationDays)
	}

	rows := make([]response.HallsLoadRow, 0, len(aggregated))
	var totalBookings uint64
	var totalRevenue float64
	var totalBookedDays float64
	availablePerHallDays := billableDaysBetween(from, to)
	for _, agg := range aggregated {
		totalBookings += agg.bookingsCount
		totalRevenue += agg.revenue
		totalBookedDays += agg.bookedDays

		row := response.HallsLoadRow{
			HallID:        agg.hallID,
			HallName:      agg.hallName,
			BookingsCount: agg.bookingsCount,
			Revenue:       round2(agg.revenue),
			BookedDays:    round2(agg.bookedDays),
			AvgCheck:      safeAvg(agg.revenue, agg.bookingsCount),
			Occupancy:     safePercent(agg.bookedDays, availablePerHallDays),
		}
		rows = append(rows, row)
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Revenue == rows[j].Revenue {
			return rows[i].BookingsCount > rows[j].BookingsCount
		}
		return rows[i].Revenue > rows[j].Revenue
	})

	totals := response.HallsLoadTotals{
		HallsCount:       hallCount,
		BookingsCount:    totalBookings,
		Revenue:          round2(totalRevenue),
		BookedDays:       round2(totalBookedDays),
		AvgCheck:         safeAvg(totalRevenue, totalBookings),
		AverageOccupancy: safePercent(totalBookedDays, availablePerHallDays*float64(hallCount)),
	}

	return &response.HallsLoadReport{
		StudioName:  studioName,
		Title:       "Отчет по загрузке залов",
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		From:        datetime.Format(from),
		To:          datetime.Format(to),
		Filters: response.HallsLoadFilters{
			From:   datetime.Format(from),
			To:     datetime.Format(to),
			HallID: hallID,
		},
		Rows:   rows,
		Totals: totals,
	}, nil
}

func (s *reportService) GetClientsReport(userID uint64, from, to time.Time, hallID *uint64, limit uint64) (*response.ClientsReport, error) {
	if err := s.ensureAdmin(userID); err != nil {
		return nil, err
	}
	if !from.Before(to) {
		return nil, errs.BadRequest("Cannot get clients report", "'from' must be before 'to'")
	}
	if limit == 0 {
		limit = 20
	}
	if limit > 200 {
		limit = 200
	}

	bookings, err := s.reportDatabase.FindReportBookings(from, to, hallID)
	if err != nil {
		return nil, errs.InternalServerError("Cannot get clients report", err.Error())
	}

	aggregated := map[uint64]*clientAggregation{}
	var totalBookings uint64
	var totalRevenue float64

	for _, booking := range bookings {
		clampedStart := maxTime(booking.StartDateTime, from)
		clampedEnd := minTime(booking.EndDateTime, to)
		if !clampedStart.Before(clampedEnd) {
			continue
		}

		durationDays := billableDaysBetween(booking.StartDateTime, booking.EndDateTime)
		if durationDays <= 0 {
			continue
		}
		overlapDays := billableDaysBetween(clampedStart, clampedEnd)
		overlapRevenue := booking.TotalPrice * (overlapDays / durationDays)

		agg, ok := aggregated[booking.UserID]
		if !ok {
			agg = &clientAggregation{
				userID:   booking.UserID,
				fullName: booking.User.FullName,
				email:    booking.User.Email,
			}
			aggregated[booking.UserID] = agg
		}

		agg.bookingsCount++
		agg.totalSpent += overlapRevenue
		if booking.StartDateTime.After(agg.lastBookingAt) {
			agg.lastBookingAt = booking.StartDateTime
		}

		totalBookings++
		totalRevenue += overlapRevenue
	}

	rows := make([]response.ClientsRow, 0, len(aggregated))
	for _, agg := range aggregated {
		row := response.ClientsRow{
			UserID:        agg.userID,
			FullName:      agg.fullName,
			Email:         agg.email,
			BookingsCount: agg.bookingsCount,
			TotalSpent:    round2(agg.totalSpent),
			AvgCheck:      safeAvg(agg.totalSpent, agg.bookingsCount),
			LastBookingAt: datetime.Format(agg.lastBookingAt),
		}
		rows = append(rows, row)
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].TotalSpent == rows[j].TotalSpent {
			return rows[i].BookingsCount > rows[j].BookingsCount
		}
		return rows[i].TotalSpent > rows[j].TotalSpent
	})
	if uint64(len(rows)) > limit {
		rows = rows[:limit]
	}

	return &response.ClientsReport{
		StudioName:  studioName,
		Title:       "Отчет по клиентам",
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		From:        datetime.Format(from),
		To:          datetime.Format(to),
		Filters: response.ClientsFilters{
			From:   datetime.Format(from),
			To:     datetime.Format(to),
			HallID: hallID,
			Limit:  limit,
		},
		Rows: rows,
		Summary: response.ClientsSummary{
			UniqueClients: uint64(len(aggregated)),
			BookingsCount: totalBookings,
			Revenue:       round2(totalRevenue),
			AvgCheck:      safeAvg(totalRevenue, totalBookings),
		},
	}, nil
}

func (s *reportService) GetBookingsDynamicsReport(userID uint64, from, to time.Time, hallID *uint64, groupBy string) (*response.BookingsDynamicsReport, error) {
	if err := s.ensureAdmin(userID); err != nil {
		return nil, err
	}
	if !from.Before(to) {
		return nil, errs.BadRequest("Cannot get bookings dynamics report", "'from' must be before 'to'")
	}
	groupBy = normalizeGroupBy(groupBy)
	if groupBy == "" || groupBy == "hall" {
		return nil, errs.BadRequest("Cannot get bookings dynamics report", "group_by must be one of: day, week, month")
	}

	hallCount, err := s.reportDatabase.CountHalls(hallID)
	if err != nil {
		return nil, errs.InternalServerError("Cannot get bookings dynamics report", err.Error())
	}
	if hallCount == 0 {
		if hallID != nil {
			return nil, errs.NotFound("Cannot get bookings dynamics report", "hall not found")
		}
		return nil, errs.BadRequest("Cannot get bookings dynamics report", "no halls found for report")
	}

	bookings, err := s.reportDatabase.FindReportBookings(from, to, hallID)
	if err != nil {
		return nil, errs.InternalServerError("Cannot get bookings dynamics report", err.Error())
	}

	aggregated := map[string]*dynamicAggregation{}
	for _, booking := range bookings {
		clampedStart := maxTime(booking.StartDateTime, from)
		clampedEnd := minTime(booking.EndDateTime, to)
		if !clampedStart.Before(clampedEnd) {
			continue
		}
		durationDays := billableDaysBetween(booking.StartDateTime, booking.EndDateTime)
		if durationDays <= 0 {
			continue
		}

		for segmentStart := clampedStart; segmentStart.Before(clampedEnd); {
			bucketStart := truncateToBucket(segmentStart, groupBy)
			bucketEnd := nextBucketStart(bucketStart, groupBy)
			segmentEnd := minTime(clampedEnd, bucketEnd)
			segmentDays := billableDaysBetween(segmentStart, segmentEnd)

			key := bucketKey(bucketStart, groupBy)
			agg, ok := aggregated[key]
			if !ok {
				agg = &dynamicAggregation{
					bucketStart: maxTime(bucketStart, from),
					bucketEnd:   minTime(bucketEnd, to),
				}
				aggregated[key] = agg
			}
			agg.bookingsCount++
			agg.bookedDays += segmentDays
			agg.revenue += booking.TotalPrice * (segmentDays / durationDays)

			segmentStart = segmentEnd
		}
	}

	keys := make([]string, 0, len(aggregated))
	for key := range aggregated {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	rows := make([]response.BookingsDynamicsRow, 0, len(keys))
	uniqueBookingsCount := countOverlappingBookings(bookings, from, to)
	var totalRevenue float64
	var totalBookedDays float64
	var occupancySum float64
	for _, key := range keys {
		agg := aggregated[key]
		availableDays := billableDaysBetween(agg.bucketStart, agg.bucketEnd) * float64(hallCount)
		occupancy := safePercent(agg.bookedDays, availableDays)
		occupancySum += occupancy
		totalRevenue += agg.revenue
		totalBookedDays += agg.bookedDays

		rows = append(rows, response.BookingsDynamicsRow{
			Bucket:        key,
			DateFrom:      datetime.Format(agg.bucketStart),
			DateTo:        datetime.Format(agg.bucketEnd),
			BookingsCount: agg.bookingsCount,
			Revenue:       round2(agg.revenue),
			BookedDays:    round2(agg.bookedDays),
			Occupancy:     occupancy,
		})
	}

	avgOccupancy := 0.0
	if len(rows) > 0 {
		avgOccupancy = round2(occupancySum / float64(len(rows)))
	}

	return &response.BookingsDynamicsReport{
		StudioName:  studioName,
		Title:       "Динамика бронирований",
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		From:        datetime.Format(from),
		To:          datetime.Format(to),
		GroupBy:     groupBy,
		Filters: response.BookingsDynamicsFilters{
			From:    datetime.Format(from),
			To:      datetime.Format(to),
			HallID:  hallID,
			GroupBy: groupBy,
		},
		Rows: rows,
		Totals: response.BookingsDynamicsTotals{
			RowsCount:        uint64(len(rows)),
			BookingsCount:    uniqueBookingsCount,
			Revenue:          round2(totalRevenue),
			BookedDays:       round2(totalBookedDays),
			AvgCheck:         safeAvg(totalRevenue, uniqueBookingsCount),
			AverageOccupancy: avgOccupancy,
		},
	}, nil
}

func (s *reportService) GetHallsLoadReportPDF(userID uint64, from, to time.Time, hallID *uint64) ([]byte, error) {
	report, err := s.GetHallsLoadReport(userID, from, to, hallID)
	if err != nil {
		return nil, err
	}
	headers := []string{"Зал", "Брон.", "Выручка", "Дни", "Ср. чек", "Загрузка"}
	rows := make([][]string, 0, len(report.Rows))
	for _, row := range report.Rows {
		rows = append(rows, []string{
			row.HallName,
			fmt.Sprintf("%d", row.BookingsCount),
			fmt.Sprintf("%.2f", row.Revenue),
			fmt.Sprintf("%.2f", row.BookedDays),
			fmt.Sprintf("%.2f", row.AvgCheck),
			fmt.Sprintf("%.2f%%", row.Occupancy),
		})
	}
	meta := []string{
		fmt.Sprintf("Период: %s", formatPeriod(report.From, report.To)),
		fmt.Sprintf("Итого: выручка %.2f, броней %d", report.Totals.Revenue, report.Totals.BookingsCount),
	}
	summaryRows := [][]string{
		{
			fmt.Sprintf("Залов: %d", report.Totals.HallsCount),
			fmt.Sprintf("Бронирований: %d", report.Totals.BookingsCount),
			fmt.Sprintf("Выручка: %.2f", report.Totals.Revenue),
		},
		{
			fmt.Sprintf("Занято дней: %.2f", report.Totals.BookedDays),
			fmt.Sprintf("Средний чек: %.2f", report.Totals.AvgCheck),
			fmt.Sprintf("Средняя загрузка: %.2f%%", report.Totals.AverageOccupancy),
		},
	}
	return buildSimpleReportPDF(report.StudioName, report.Title, report.GeneratedAt, meta, summaryRows, headers, rows)
}

func (s *reportService) GetClientsReportPDF(userID uint64, from, to time.Time, hallID *uint64, limit uint64) ([]byte, error) {
	report, err := s.GetClientsReport(userID, from, to, hallID, limit)
	if err != nil {
		return nil, err
	}
	headers := []string{"ID", "Клиент", "Контакт", "Брон.", "Потрачено", "Последняя"}
	rows := make([][]string, 0, len(report.Rows))
	for _, row := range report.Rows {
		name := "-"
		if row.FullName != nil {
			name = *row.FullName
		}
		email := "-"
		if row.Email != nil {
			email = *row.Email
		}
		rows = append(rows, []string{
			fmt.Sprintf("%d", row.UserID),
			truncateText(name, 18),
			truncateText(email, 20),
			fmt.Sprintf("%d", row.BookingsCount),
			fmt.Sprintf("%.2f", row.TotalSpent),
			formatDateOnly(row.LastBookingAt),
		})
	}
	meta := []string{
		fmt.Sprintf("Период: %s", formatPeriod(report.From, report.To)),
		fmt.Sprintf("Клиентов: %d | Выручка: %.2f", report.Summary.UniqueClients, report.Summary.Revenue),
	}
	summaryRows := [][]string{
		{
			fmt.Sprintf("Уникальных клиентов: %d", report.Summary.UniqueClients),
			fmt.Sprintf("Бронирований: %d", report.Summary.BookingsCount),
			fmt.Sprintf("Выручка: %.2f", report.Summary.Revenue),
		},
		{
			fmt.Sprintf("Средний чек: %.2f", report.Summary.AvgCheck),
			fmt.Sprintf("TOP лимит: %d", report.Filters.Limit),
			fmt.Sprintf("Период: %s", formatPeriod(report.From, report.To)),
		},
	}
	return buildSimpleReportPDF(report.StudioName, report.Title, report.GeneratedAt, meta, summaryRows, headers, rows)
}

func (s *reportService) GetBookingsDynamicsReportPDF(userID uint64, from, to time.Time, hallID *uint64, groupBy string) ([]byte, error) {
	report, err := s.GetBookingsDynamicsReport(userID, from, to, hallID, groupBy)
	if err != nil {
		return nil, err
	}
	return buildBookingsDynamicsReportPDF(report)
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
			m.Text(fmt.Sprintf("Период: %s", formatPeriod(report.From, report.To)), props.Text{Size: 9, Color: dark})
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
			fmt.Sprintf("Загружено дней: %.2f", report.Summary.TotalBookedDays),
			fmt.Sprintf("Доступно дней: %.2f", report.Summary.AvailableDays),
			"Загрузка: " + formatMetric("occupancy", report.Totals),
		},
		{
			fmt.Sprintf("Строк отчета: %d", report.Summary.RowsCount),
			fmt.Sprintf("Залов в отчете: %d", report.Summary.HallsCount),
			fmt.Sprintf("Интервал: %.0f дн", billableDaysBetween(from, to)),
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

		bookingDuration := billableDaysBetween(booking.StartDateTime, booking.EndDateTime)
		if bookingDuration <= 0 {
			continue
		}

		if groupBy == "hall" {
			key := fmt.Sprintf("hall:%d", booking.HallID)
			agg := getAggregation(aggregations, key, from, to, &booking.HallID, &booking.Hall.Name)
			overlapDays := billableDaysBetween(clampedStart, clampedEnd)
			agg.bookedDays += overlapDays
			agg.revenue += booking.TotalPrice * (overlapDays / bookingDuration)
			agg.bookingsCount++
			continue
		}

		for segmentStart := clampedStart; segmentStart.Before(clampedEnd); {
			bucketStart := truncateToBucket(segmentStart, groupBy)
			bucketEnd := nextBucketStart(bucketStart, groupBy)
			segmentEnd := minTime(clampedEnd, bucketEnd)
			segmentDays := billableDaysBetween(segmentStart, segmentEnd)

			key := bucketKey(bucketStart, groupBy)
			agg := getAggregation(aggregations, key, maxTime(bucketStart, from), minTime(bucketEnd, to), nil, nil)
			agg.bookedDays += segmentDays
			agg.revenue += booking.TotalPrice * (segmentDays / bookingDuration)
			agg.bookingsCount++

			segmentStart = segmentEnd
		}
	}

	return aggregations
}

func buildSalesResponse(aggregations map[string]*salesAggregation, from, to time.Time, groupBy string, metrics []string, hallsCount uint64, uniqueBookingsCount uint64) ([]response.SalesReportRow, response.SalesReportMetrics, response.SalesReportSummary) {
	keys := make([]string, 0, len(aggregations))
	for key := range aggregations {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	rows := make([]response.SalesReportRow, 0, len(keys))
	var totalRevenue float64
	var totalBookedDays float64

	for _, key := range keys {
		agg := aggregations[key]
		totalRevenue += agg.revenue
		totalBookedDays += agg.bookedDays

		availableDays := billableDaysBetween(agg.bucketStart, agg.bucketEnd)
		if groupBy != "hall" {
			availableDays *= float64(hallsCount)
		}
		metricsValue := selectMetrics(metrics, agg.revenue, agg.bookingsCount, agg.bookedDays, availableDays)

		row := response.SalesReportRow{
			Bucket:     key,
			HallID:     agg.hallID,
			HallName:   agg.hallName,
			DateFrom:   datetime.Format(agg.bucketStart),
			DateTo:     datetime.Format(agg.bucketEnd),
			BookedDays: round2(agg.bookedDays),
			Metrics:    metricsValue,
		}
		rows = append(rows, row)
	}

	totalAvailableDays := billableDaysBetween(from, to)
	if groupBy != "hall" {
		totalAvailableDays *= float64(hallsCount)
	}
	totals := selectMetrics(metrics, totalRevenue, uniqueBookingsCount, totalBookedDays, totalAvailableDays)
	summary := response.SalesReportSummary{
		RowsCount:       uint64(len(rows)),
		HallsCount:      hallsCount,
		TotalBookedDays: round2(totalBookedDays),
		AvailableDays:   round2(totalAvailableDays),
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

func selectMetrics(selected []string, revenue float64, bookingsCount uint64, bookedDays, availableDays float64) response.SalesReportMetrics {
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
			if availableDays > 0 {
				occupancy = (bookedDays / availableDays) * 100
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

	if cols >= 6 {
		base := 12 / cols
		rem := 12 % cols
		grid := make([]uint, 0, cols)
		for i := 0; i < cols; i++ {
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

	first := uint(4)
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
		return datetime.Format(bucketStart)
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

func truncateText(value string, maxRunes int) string {
	if maxRunes <= 0 {
		return ""
	}
	r := []rune(value)
	if len(r) <= maxRunes {
		return value
	}
	if maxRunes <= 1 {
		return "…"
	}
	return string(r[:maxRunes-1]) + "…"
}

func formatDateOnly(value string) string {
	if value == "" {
		return "-"
	}
	if t, err := datetime.Parse(value); err == nil {
		return datetime.Format(t)
	}
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return truncateText(value, 10)
	}
	return datetime.Format(t)
}

func safeAvg(sum float64, count uint64) float64 {
	if count == 0 {
		return 0
	}
	return round2(sum / float64(count))
}

func safePercent(part, total float64) float64 {
	if total <= 0 {
		return 0
	}
	return round2((part / total) * 100)
}

func countOverlappingBookings(bookings []model.Booking, from, to time.Time) uint64 {
	var total uint64
	for _, booking := range bookings {
		if maxTime(booking.StartDateTime, from).Before(minTime(booking.EndDateTime, to)) {
			total++
		}
	}
	return total
}

func billableDaysBetween(start, end time.Time) float64 {
	if !start.Before(end) {
		return 0
	}
	loc := bookingCalendarLocation()
	normalizedStart := startOfDayInLocation(start, loc)
	normalizedEnd := startOfDayInLocation(end, loc)
	if !isStartOfDayInLocation(end, loc) {
		normalizedEnd = normalizedEnd.AddDate(0, 0, 1)
	}
	if !normalizedStart.Before(normalizedEnd) {
		return 0
	}

	var days float64
	for day := normalizedStart; day.Before(normalizedEnd); day = day.AddDate(0, 0, 1) {
		days++
	}
	return days
}

func formatPeriod(fromValue, toValue string) string {
	from := formatDateOnly(fromValue)
	to := formatDateOnly(toValue)
	return from + " - " + to
}

func (s *reportService) ensureAdmin(userID uint64) error {
	user, err := s.userService.FindByID(userID)
	if err != nil {
		return err
	}
	if user.Role != enum.RoleAdmin {
		return errs.Forbidden("Cannot get report", "You do not have permission to access admin reports")
	}
	return nil
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

func buildSimpleReportPDF(studio, title, generatedAt string, meta []string, summaryRows [][]string, headers []string, rows [][]string) ([]byte, error) {
	m := pdf.NewMaroto(consts.Portrait, consts.A4)
	if err := configureUTF8Font(m); err != nil {
		return nil, errs.InternalServerError("Cannot generate report PDF", err.Error())
	}
	m.SetPageMargins(10, 12, 10)

	dark := color.Color{Red: 28, Green: 45, Blue: 68}
	light := color.Color{Red: 244, Green: 247, Blue: 252}
	white := color.Color{Red: 255, Green: 255, Blue: 255}

	m.SetBackgroundColor(dark)
	m.Row(9, func() {
		m.Col(12, func() {
			m.Text(studio, props.Text{Top: 2, Style: consts.Bold, Size: 15, Align: consts.Center, Color: white})
		})
	})
	m.SetBackgroundColor(white)

	m.Row(8, func() {
		m.Col(12, func() {
			m.Text(title, props.Text{Top: 1, Style: consts.Bold, Size: 12, Align: consts.Center, Color: dark})
		})
	})

	for _, line := range meta {
		text := line
		m.Row(6, func() {
			m.Col(12, func() {
				m.Text(text, props.Text{Size: 9, Color: dark})
			})
		})
	}
	m.Row(6, func() {
		m.Col(12, func() {
			m.Text("Сформирован: "+generatedAt, props.Text{Size: 9, Color: dark})
		})
	})
	m.Line(1, props.Line{Color: dark, Width: 0.2})

	if len(summaryRows) > 0 {
		m.SetBackgroundColor(light)
		m.Row(8, func() {
			m.Col(12, func() {
				m.Text("Итоги периода", props.Text{Top: 1.5, Style: consts.Bold, Size: 11, Align: consts.Center, Color: dark})
			})
		})
		m.SetBackgroundColor(white)

		for _, summaryRow := range summaryRows {
			values := summaryRow
			cols := len(values)
			if cols == 0 {
				continue
			}
			colWidth := uint(12 / cols)
			m.Row(6, func() {
				for _, value := range values {
					text := value
					m.Col(colWidth, func() {
						m.Text(text, props.Text{Size: 8.7, Color: dark})
					})
				}
			})
		}
		m.Line(1, props.Line{Color: dark, Width: 0.2})
	}

	if len(rows) == 0 {
		m.Row(7, func() {
			m.Col(12, func() {
				m.Text("Нет данных за выбранный период", props.Text{Style: consts.Italic, Size: 9, Color: dark})
			})
		})
	} else {
		grid := buildGridSizes(len(headers))
		m.SetBackgroundColor(dark)
		m.TableList(headers, rows, props.TableList{
			HeaderProp: props.TableListContent{
				Style:     consts.Bold,
				Size:      9,
				Color:     white,
				GridSizes: grid,
			},
			ContentProp: props.TableListContent{
				Size:      8,
				Color:     dark,
				GridSizes: grid,
			},
			Align:                  consts.Center,
			AlternatedBackground:   &light,
			HeaderContentSpace:     2,
			VerticalContentPadding: 1,
			Line:                   true,
			LineProp:               props.Line{Color: color.Color{Red: 215, Green: 223, Blue: 236}, Width: 0.15},
		})
		m.SetBackgroundColor(white)
	}

	out, err := m.Output()
	if err != nil {
		return nil, errs.InternalServerError("Cannot generate report PDF", err.Error())
	}
	return out.Bytes(), nil
}

func buildBookingsDynamicsReportPDF(report *response.BookingsDynamicsReport) ([]byte, error) {
	m := pdf.NewMaroto(consts.Portrait, consts.A4)
	if err := configureUTF8Font(m); err != nil {
		return nil, errs.InternalServerError("Cannot generate report PDF", err.Error())
	}
	m.SetPageMargins(10, 12, 10)

	dark := color.Color{Red: 28, Green: 45, Blue: 68}
	light := color.Color{Red: 244, Green: 247, Blue: 252}
	white := color.Color{Red: 255, Green: 255, Blue: 255}

	m.SetBackgroundColor(dark)
	m.Row(9, func() {
		m.Col(12, func() {
			m.Text(report.StudioName, props.Text{Top: 2, Style: consts.Bold, Size: 15, Align: consts.Center, Color: white})
		})
	})
	m.SetBackgroundColor(white)

	m.Row(8, func() {
		m.Col(12, func() {
			m.Text(report.Title+" (диаграммы)", props.Text{Top: 1, Style: consts.Bold, Size: 12, Align: consts.Center, Color: dark})
		})
	})

	m.Row(6, func() {
		m.Col(12, func() {
			m.Text(fmt.Sprintf("Период: %s", formatPeriod(report.From, report.To)), props.Text{Size: 9, Color: dark})
		})
	})
	m.Row(6, func() {
		m.Col(12, func() {
			m.Text(fmt.Sprintf("Группировка: %s | Сегментов: %d", report.GroupBy, report.Totals.RowsCount), props.Text{Size: 9, Color: dark})
		})
	})
	m.Row(6, func() {
		m.Col(12, func() {
			m.Text("Сформирован: "+report.GeneratedAt, props.Text{Size: 9, Color: dark})
		})
	})
	m.Line(1, props.Line{Color: dark, Width: 0.2})

	summaryRows := [][]string{
		{
			fmt.Sprintf("Бронирований: %d", report.Totals.BookingsCount),
			fmt.Sprintf("Выручка: %.2f", report.Totals.Revenue),
			fmt.Sprintf("Строк динамики: %d", report.Totals.RowsCount),
		},
		{
			fmt.Sprintf("Занято дней: %.2f", report.Totals.BookedDays),
			fmt.Sprintf("Средний чек: %.2f", report.Totals.AvgCheck),
			fmt.Sprintf("Средняя загрузка: %.2f%%", report.Totals.AverageOccupancy),
		},
	}

	m.SetBackgroundColor(light)
	m.Row(8, func() {
		m.Col(12, func() {
			m.Text("Итоги периода", props.Text{Top: 1.5, Style: consts.Bold, Size: 11, Align: consts.Center, Color: dark})
		})
	})
	m.SetBackgroundColor(white)
	for _, summaryRow := range summaryRows {
		values := summaryRow
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

	if len(report.Rows) == 0 {
		m.Row(7, func() {
			m.Col(12, func() {
				m.Text("Нет данных за выбранный период", props.Text{Style: consts.Italic, Size: 9, Color: dark})
			})
		})
	} else {
		bookingsValues := make([]float64, 0, len(report.Rows))
		revenueValues := make([]float64, 0, len(report.Rows))
		occupancyValues := make([]float64, 0, len(report.Rows))
		avgCheckValues := make([]float64, 0, len(report.Rows))
		for _, row := range report.Rows {
			bookingsValues = append(bookingsValues, float64(row.BookingsCount))
			revenueValues = append(revenueValues, row.Revenue)
			occupancyValues = append(occupancyValues, row.Occupancy)
			avgCheckValues = append(avgCheckValues, safeAvg(row.Revenue, row.BookingsCount))
		}

		kpis := []struct {
			title string
			value string
			delta string
		}{
			{
				title: "Бронирования",
				value: fmt.Sprintf("%d", report.Totals.BookingsCount),
				delta: formatDeltaToPrevious(bookingsValues),
			},
			{
				title: "Выручка",
				value: fmt.Sprintf("%.2f", report.Totals.Revenue),
				delta: formatDeltaToPrevious(revenueValues),
			},
			{
				title: "Средний чек",
				value: fmt.Sprintf("%.2f", report.Totals.AvgCheck),
				delta: formatDeltaToPrevious(avgCheckValues),
			},
			{
				title: "Загрузка",
				value: fmt.Sprintf("%.2f%%", report.Totals.AverageOccupancy),
				delta: formatDeltaToPrevious(occupancyValues),
			},
		}

		m.SetBackgroundColor(light)
		m.Row(7, func() {
			m.Col(12, func() {
				m.Text("Ключевые показатели (сравнение с предыдущим сегментом)", props.Text{
					Top:   1.2,
					Style: consts.Bold,
					Size:  10,
					Align: consts.Center,
					Color: dark,
				})
			})
		})
		m.SetBackgroundColor(white)

		for i := 0; i < len(kpis); i += 2 {
			leftKPI := kpis[i]
			rightKPI := kpis[i+1]
			m.Row(16, func() {
				m.Col(6, func() {
					m.Text(leftKPI.title, props.Text{Top: 1.3, Size: 8, Style: consts.Bold, Color: dark})
					m.Text(leftKPI.value, props.Text{Top: 6, Size: 11, Style: consts.Bold, Color: dark})
					m.Text(leftKPI.delta, props.Text{Top: 11, Size: 7.8, Style: consts.Italic, Color: dark})
				})
				m.Col(6, func() {
					m.Text(rightKPI.title, props.Text{Top: 1.3, Size: 8, Style: consts.Bold, Color: dark})
					m.Text(rightKPI.value, props.Text{Top: 6, Size: 11, Style: consts.Bold, Color: dark})
					m.Text(rightKPI.delta, props.Text{Top: 11, Size: 7.8, Style: consts.Italic, Color: dark})
				})
			})
		}
		m.Line(1, props.Line{Color: dark, Width: 0.15})

		if err := drawSparklineCard(m, dark, light, "Тренд бронирований", bookingsValues, func(v float64) string {
			return fmt.Sprintf("%.0f", v)
		}); err != nil {
			return nil, errs.InternalServerError("Cannot generate report PDF", err.Error())
		}
		if err := drawSparklineCard(m, dark, light, "Тренд выручки", revenueValues, func(v float64) string {
			return fmt.Sprintf("%.2f", v)
		}); err != nil {
			return nil, errs.InternalServerError("Cannot generate report PDF", err.Error())
		}
		if err := drawSparklineCard(m, dark, light, "Тренд среднего чека", avgCheckValues, func(v float64) string {
			return fmt.Sprintf("%.2f", v)
		}); err != nil {
			return nil, errs.InternalServerError("Cannot generate report PDF", err.Error())
		}
		if err := drawSparklineCard(m, dark, light, "Тренд загрузки (%)", occupancyValues, func(v float64) string {
			return fmt.Sprintf("%.2f%%", v)
		}); err != nil {
			return nil, errs.InternalServerError("Cannot generate report PDF", err.Error())
		}
	}

	out, err := m.Output()
	if err != nil {
		return nil, errs.InternalServerError("Cannot generate report PDF", err.Error())
	}
	return out.Bytes(), nil
}

func drawSparklineCard(
	m pdf.Maroto,
	dark color.Color,
	light color.Color,
	title string,
	values []float64,
	formatValue func(v float64) string,
) error {
	if len(values) == 0 {
		return nil
	}

	sparklineBase64, err := buildSparklinePNGBase64(values, 920, 180)
	if err != nil {
		return err
	}

	minValue := values[0]
	maxValue := values[0]
	for _, value := range values[1:] {
		if value < minValue {
			minValue = value
		}
		if value > maxValue {
			maxValue = value
		}
	}
	lastValue := values[len(values)-1]

	m.SetBackgroundColor(light)
	m.Row(7, func() {
		m.Col(12, func() {
			m.Text(title, props.Text{Top: 1.2, Style: consts.Bold, Size: 10, Align: consts.Center, Color: dark})
		})
	})
	m.SetBackgroundColor(color.Color{Red: 255, Green: 255, Blue: 255})
	m.Row(5, func() {
		m.Col(12, func() {
			m.Text(fmt.Sprintf("min: %s | max: %s | last: %s", formatValue(minValue), formatValue(maxValue), formatValue(lastValue)), props.Text{
				Size:  8,
				Style: consts.Italic,
				Color: dark,
			})
		})
	})
	m.Row(30, func() {
		m.Col(12, func() {
			if imageErr := m.Base64Image(sparklineBase64, consts.Extension("png"), props.Rect{
				Center:  true,
				Percent: 100,
			}); imageErr != nil && err == nil {
				err = imageErr
			}
		})
	})
	if err != nil {
		return err
	}
	m.Line(1, props.Line{Color: dark, Width: 0.15})
	return nil
}

func buildSparklinePNGBase64(values []float64, width, height int) (string, error) {
	if len(values) == 0 {
		values = []float64{0}
	}
	if width < 120 {
		width = 120
	}
	if height < 60 {
		height = 60
	}

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	bgColor := stdcolor.RGBA{R: 247, G: 250, B: 255, A: 255}
	gridColor := stdcolor.RGBA{R: 223, G: 231, B: 243, A: 255}
	lineColor := stdcolor.RGBA{R: 36, G: 78, B: 145, A: 255}
	pointColor := stdcolor.RGBA{R: 16, G: 44, B: 89, A: 255}

	imagedraw.Draw(img, img.Bounds(), &image.Uniform{C: bgColor}, image.Point{}, imagedraw.Src)

	left := 28
	right := width - 16
	top := 12
	bottom := height - 16
	if right <= left {
		right = left + 1
	}
	if bottom <= top {
		bottom = top + 1
	}

	for i := 1; i <= 3; i++ {
		y := top + ((bottom-top)*i)/4
		drawLineRGBA(img, left, y, right, y, gridColor)
	}

	minValue := values[0]
	maxValue := values[0]
	for _, v := range values[1:] {
		if v < minValue {
			minValue = v
		}
		if v > maxValue {
			maxValue = v
		}
	}

	denominator := maxValue - minValue
	xStep := 0.0
	if len(values) > 1 {
		xStep = float64(right-left) / float64(len(values)-1)
	}

	points := make([]image.Point, 0, len(values))
	for i, v := range values {
		x := left + int(math.Round(float64(i)*xStep))
		ratio := 0.5
		if denominator > 0 {
			ratio = (v - minValue) / denominator
		}
		y := bottom - int(math.Round(ratio*float64(bottom-top)))
		points = append(points, image.Point{X: x, Y: y})
	}

	for i := 0; i < len(points)-1; i++ {
		drawLineRGBA(img, points[i].X, points[i].Y, points[i+1].X, points[i+1].Y, lineColor)
	}
	for _, p := range points {
		drawFilledCircleRGBA(img, p.X, p.Y, 3, pointColor)
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func drawLineRGBA(img *image.RGBA, x0, y0, x1, y1 int, clr stdcolor.RGBA) {
	dx := int(math.Abs(float64(x1 - x0)))
	dy := -int(math.Abs(float64(y1 - y0)))
	sx := -1
	if x0 < x1 {
		sx = 1
	}
	sy := -1
	if y0 < y1 {
		sy = 1
	}
	err := dx + dy
	for {
		setPixelSafeRGBA(img, x0, y0, clr)
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x0 += sx
		}
		if e2 <= dx {
			err += dx
			y0 += sy
		}
	}
}

func drawFilledCircleRGBA(img *image.RGBA, cx, cy, r int, clr stdcolor.RGBA) {
	for y := -r; y <= r; y++ {
		for x := -r; x <= r; x++ {
			if x*x+y*y <= r*r {
				setPixelSafeRGBA(img, cx+x, cy+y, clr)
			}
		}
	}
}

func setPixelSafeRGBA(img *image.RGBA, x, y int, clr stdcolor.RGBA) {
	bounds := img.Bounds()
	if x < bounds.Min.X || x >= bounds.Max.X || y < bounds.Min.Y || y >= bounds.Max.Y {
		return
	}
	img.SetRGBA(x, y, clr)
}

func formatDeltaToPrevious(values []float64) string {
	if len(values) < 2 {
		return "Δ к пред. сегменту: n/a"
	}
	prev := values[len(values)-2]
	curr := values[len(values)-1]
	if prev == 0 {
		if curr == 0 {
			return "Δ к пред. сегменту: 0.00%"
		}
		return "Δ к пред. сегменту: n/a"
	}
	delta := ((curr - prev) / prev) * 100
	return fmt.Sprintf("Δ к пред. сегменту: %+0.2f%%", round2(delta))
}
