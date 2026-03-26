package datetime

import (
	"college-graduation-project-backend/internal/config"
	"strings"
	"time"
)

const Layout = "2006-01-02"

func Location() *time.Location {
	if config.AppLocation != nil {
		return config.AppLocation
	}
	return time.UTC
}

func Parse(value string) (time.Time, error) {
	return time.ParseInLocation(Layout, strings.TrimSpace(value), Location())
}

func StartOfDay(t time.Time) time.Time {
	local := t.In(Location())
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, Location())
}

func EndOfDay(t time.Time) time.Time {
	local := t.In(Location())
	return time.Date(local.Year(), local.Month(), local.Day(), 23, 59, 59, 0, Location())
}

func Format(t time.Time) string {
	return t.In(Location()).Format(Layout)
}
