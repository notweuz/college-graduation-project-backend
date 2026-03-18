package response

type SalesReport struct {
	StudioName  string             `json:"studio_name"`
	Title       string             `json:"title"`
	GeneratedAt string             `json:"generated_at"`
	From        string             `json:"from"`
	To          string             `json:"to"`
	GroupBy     string             `json:"group_by"`
	Metrics     []string           `json:"metrics"`
	Filters     SalesReportFilters `json:"filters"`
	Rows        []SalesReportRow   `json:"rows"`
	Totals      SalesReportMetrics `json:"totals"`
	Summary     SalesReportSummary `json:"summary"`
}

type SalesReportRow struct {
	Bucket      string             `json:"bucket"`
	HallID      *uint64            `json:"hall_id,omitempty"`
	HallName    *string            `json:"hall_name,omitempty"`
	DateFrom    string             `json:"date_from"`
	DateTo      string             `json:"date_to"`
	BookedHours float64            `json:"booked_hours"`
	Metrics     SalesReportMetrics `json:"metrics"`
}

type SalesReportFilters struct {
	From    string   `json:"from"`
	To      string   `json:"to"`
	HallID  *uint64  `json:"hall_id,omitempty"`
	GroupBy string   `json:"group_by"`
	Metrics []string `json:"metrics"`
}

type SalesReportMetrics struct {
	Revenue       *float64 `json:"revenue,omitempty"`
	BookingsCount *uint64  `json:"bookings_count,omitempty"`
	AvgCheck      *float64 `json:"avg_check,omitempty"`
	Occupancy     *float64 `json:"occupancy,omitempty"`
}

type SalesReportSummary struct {
	RowsCount        uint64  `json:"rows_count"`
	HallsCount       uint64  `json:"halls_count"`
	TotalBookedHours float64 `json:"total_booked_hours"`
	AvailableHours   float64 `json:"available_hours"`
}

type HallsLoadReport struct {
	StudioName  string           `json:"studio_name"`
	Title       string           `json:"title"`
	GeneratedAt string           `json:"generated_at"`
	From        string           `json:"from"`
	To          string           `json:"to"`
	Filters     HallsLoadFilters `json:"filters"`
	Rows        []HallsLoadRow   `json:"rows"`
	Totals      HallsLoadTotals  `json:"totals"`
}

type HallsLoadFilters struct {
	From   string  `json:"from"`
	To     string  `json:"to"`
	HallID *uint64 `json:"hall_id,omitempty"`
}

type HallsLoadRow struct {
	HallID        uint64  `json:"hall_id"`
	HallName      string  `json:"hall_name"`
	BookingsCount uint64  `json:"bookings_count"`
	Revenue       float64 `json:"revenue"`
	BookedHours   float64 `json:"booked_hours"`
	AvgCheck      float64 `json:"avg_check"`
	Occupancy     float64 `json:"occupancy"`
}

type HallsLoadTotals struct {
	HallsCount       uint64  `json:"halls_count"`
	BookingsCount    uint64  `json:"bookings_count"`
	Revenue          float64 `json:"revenue"`
	BookedHours      float64 `json:"booked_hours"`
	AvgCheck         float64 `json:"avg_check"`
	AverageOccupancy float64 `json:"average_occupancy"`
}

type ClientsReport struct {
	StudioName  string         `json:"studio_name"`
	Title       string         `json:"title"`
	GeneratedAt string         `json:"generated_at"`
	From        string         `json:"from"`
	To          string         `json:"to"`
	Filters     ClientsFilters `json:"filters"`
	Rows        []ClientsRow   `json:"rows"`
	Summary     ClientsSummary `json:"summary"`
}

type ClientsFilters struct {
	From   string  `json:"from"`
	To     string  `json:"to"`
	HallID *uint64 `json:"hall_id,omitempty"`
	Limit  uint64  `json:"limit"`
}

type ClientsRow struct {
	UserID        uint64  `json:"user_id"`
	FullName      *string `json:"full_name,omitempty"`
	Email         *string `json:"email,omitempty"`
	BookingsCount uint64  `json:"bookings_count"`
	TotalSpent    float64 `json:"total_spent"`
	AvgCheck      float64 `json:"avg_check"`
	LastBookingAt string  `json:"last_booking_at"`
}

type ClientsSummary struct {
	UniqueClients uint64  `json:"unique_clients"`
	BookingsCount uint64  `json:"bookings_count"`
	Revenue       float64 `json:"revenue"`
	AvgCheck      float64 `json:"avg_check"`
}

type BookingsDynamicsReport struct {
	StudioName  string                  `json:"studio_name"`
	Title       string                  `json:"title"`
	GeneratedAt string                  `json:"generated_at"`
	From        string                  `json:"from"`
	To          string                  `json:"to"`
	GroupBy     string                  `json:"group_by"`
	Filters     BookingsDynamicsFilters `json:"filters"`
	Rows        []BookingsDynamicsRow   `json:"rows"`
	Totals      BookingsDynamicsTotals  `json:"totals"`
}

type BookingsDynamicsFilters struct {
	From    string  `json:"from"`
	To      string  `json:"to"`
	HallID  *uint64 `json:"hall_id,omitempty"`
	GroupBy string  `json:"group_by"`
}

type BookingsDynamicsRow struct {
	Bucket        string  `json:"bucket"`
	DateFrom      string  `json:"date_from"`
	DateTo        string  `json:"date_to"`
	BookingsCount uint64  `json:"bookings_count"`
	Revenue       float64 `json:"revenue"`
	BookedHours   float64 `json:"booked_hours"`
	Occupancy     float64 `json:"occupancy"`
}

type BookingsDynamicsTotals struct {
	RowsCount        uint64  `json:"rows_count"`
	BookingsCount    uint64  `json:"bookings_count"`
	Revenue          float64 `json:"revenue"`
	BookedHours      float64 `json:"booked_hours"`
	AvgCheck         float64 `json:"avg_check"`
	AverageOccupancy float64 `json:"average_occupancy"`
}
