package httpapp

import "time"

// Business dates are stored in SQL DATE columns and must not carry a timezone
// offset into the database driver. The database connection intentionally uses
// UTC for timestamps, so date-only values are represented as UTC midnight.
func parseBusinessDate(value string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02", value, time.UTC)
}

func businessToday(location *time.Location) time.Time {
	local := time.Now().In(location)
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, time.UTC)
}

func sameBusinessDate(a, b time.Time) bool {
	return a.Format("2006-01-02") == b.Format("2006-01-02")
}
