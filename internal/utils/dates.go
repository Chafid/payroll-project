package utils

import "time"

// CountWorkingDays returns the number of weekdays between two dates inclusive.
func CountWorkingDays(start, end time.Time) int {
	count := 0
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		if d.Weekday() >= time.Monday && d.Weekday() <= time.Friday {
			count++
		}
	}
	return count
}
