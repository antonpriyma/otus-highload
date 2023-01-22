package utils

import (
	"time"
)

func DatesEqual(tm1, tm2 time.Time) bool {
	tm1Year, tm1Month, tm1Day := tm1.Date()
	tm2Year, tm2Month, tm2Day := tm2.Date()

	return tm2Year == tm1Year && tm2Month == tm1Month && tm2Day == tm1Day
}
