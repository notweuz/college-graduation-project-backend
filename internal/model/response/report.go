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
