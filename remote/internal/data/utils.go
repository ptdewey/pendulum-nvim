package data

import (
	"fmt"
	"time"
)

// calculate the duration between two string timestamps (curr - prev)
func calcDuration(curr string, prev string) (time.Duration, error) {
	layout := "2006-01-02 15:04:05"

	curr_t, err := time.Parse(layout, curr)
	if err != nil {
		return time.Duration(0), err
	}

	prev_t, err := time.Parse(layout, prev)
	if err != nil {
		return time.Duration(0), err
	}

	return curr_t.Sub(prev_t), nil
}

func isTimestampInRange(timestampStr, rangeType string) (bool, error) {
	layout := "2006-01-02 15:04:05"

	// WARN: input timestamp format has to be in UTC for hour filtering (or allow a TZ config option)
	// TODO: use new TimeZone option from pendulumArgs
	timestamp, err := time.Parse(layout, timestampStr)
	if err != nil {
		return false, fmt.Errorf("error parsing timestamp: %v", err)
	}

	now := time.Now().UTC()

	var startOfRange, endOfRange time.Time

	switch rangeType {
	case "today", "day":
		startOfRange = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		endOfRange = startOfRange.Add(24 * time.Hour).Add(-time.Nanosecond)
	case "year":
		startOfRange = time.Date(now.Year(), time.January, 1, 0, 0, 0, 0, now.Location())
		endOfRange = startOfRange.Add(365 * 24 * time.Hour).Add(-time.Nanosecond)
	case "month":
		startOfRange = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		endOfRange = startOfRange.AddDate(0, 1, 0).Add(-time.Nanosecond)
	case "week":
		year, week := now.ISOWeek()
		// calculate the start of the week based on the current week and year
		startOfWeek := time.Date(year, time.January, 1, 0, 0, 0, 0, now.Location())
		for startOfWeek.Weekday() != time.Monday {
			startOfWeek = startOfWeek.AddDate(0, 0, 1)
		}
		startOfRange = startOfWeek.AddDate(0, 0, (week-1)*7)
		endOfRange = startOfRange.Add(7 * 24 * time.Hour).Add(-time.Nanosecond)
	case "hour":
		startOfRange = time.Date(now.Year(), now.Month(), now.Day(), now.Hour()-1, now.Minute(), now.Second(), 0, now.Location())
		endOfRange = startOfRange.Add(1 * time.Hour).Add(-time.Nanosecond)
	default:
		// default to "all" range if input is invalid or not provided
		startOfRange = time.Time{}
		endOfRange = time.Time{}
	}

	// handle the default "all" range case (from the earliest possible time to the latest possible time)
	if rangeType == "all" || startOfRange.IsZero() || endOfRange.IsZero() {
		startOfRange = time.Time{}
		endOfRange = time.Now().Add(1 * time.Second)
	}

	// Check if the timestamp is within the range
	return timestamp.After(startOfRange) && timestamp.Before(endOfRange), nil
}
