package lib

import (
	"time"
)

// StartOfDay returns midnight for the calendar date in value's location.
func StartOfDay(value time.Time) time.Time {
	year, month, day := value.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, value.Location())
}
