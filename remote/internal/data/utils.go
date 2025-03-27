package data

import (
	"fmt"
	"pendulum-nvim/pkg/args"
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

func calcDurationWithinHour(curr string, prev string) (time.Duration, error) {
	layout := "2006-01-02 15:04:05"

	curr_t, err := time.Parse(layout, curr)
	if err != nil {
		return 0, err
	}

	prev_t, err := time.Parse(layout, prev)
	if err != nil {
		return 0, err
	}

	// If prev_t is within the same hour as curr_t, return the direct difference
	if prev_t.Hour() == curr_t.Hour() && prev_t.Day() == curr_t.Day() {
		return curr_t.Sub(prev_t), nil
	}

	// Otherwise, clamp prev_t to the start of curr_t's hour
	clamped_prev_t := time.Date(curr_t.Year(), curr_t.Month(), curr_t.Day(), curr_t.Hour(),
		0, 0, 0, curr_t.Location())

	return curr_t.Sub(clamped_prev_t), nil
}

func isTimestampInRange(timestampStr, rangeType string) (bool, error) {
	layout := "2006-01-02 15:04:05"

	timestamp, err := time.Parse(layout, timestampStr)
	if err != nil {
		return false, fmt.Errorf("error parsing timestamp: %v", err)
	}

	// TEST: Add time-range option back, test tz change
	// TEST: ensure removing `now.loc()` calls from switch date calls doesn't break things
	tz := args.PendulumArgs().TimeZone
	loc, err := time.LoadLocation(tz)
	if err == nil {
		timestamp = timestamp.In(loc)
	}

	now := time.Now().UTC()

	var startOfRange, endOfRange time.Time

	switch rangeType {
	case "today", "day":
		startOfRange = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
		endOfRange = startOfRange.Add(24 * time.Hour).Add(-time.Nanosecond)
	case "year":
		startOfRange = time.Date(now.Year(), time.January, 1, 0, 0, 0, 0, loc)
		endOfRange = startOfRange.Add(365 * 24 * time.Hour).Add(-time.Nanosecond)
	case "month":
		startOfRange = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
		endOfRange = startOfRange.AddDate(0, 1, 0).Add(-time.Nanosecond)
	case "week":
		year, week := now.ISOWeek()
		// calculate the start of the week based on the current week and year
		startOfWeek := time.Date(year, time.January, 1, 0, 0, 0, 0, loc)
		for startOfWeek.Weekday() != time.Monday {
			startOfWeek = startOfWeek.AddDate(0, 0, 1)
		}
		startOfRange = startOfWeek.AddDate(0, 0, (week-1)*7)
		endOfRange = startOfRange.Add(7 * 24 * time.Hour).Add(-time.Nanosecond)
	case "hour":
		startOfRange = time.Date(now.Year(), now.Month(), now.Day(), now.Hour()-1, now.Minute(), now.Second(), 0, loc)
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
