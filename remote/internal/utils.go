package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

// reformat time.Duration values into a more readable format
func formatDuration(d time.Duration) string {
	if d >= 24*time.Hour {
		days := float32(d) / (24 * float32(time.Hour))
		return fmt.Sprintf("%.2fd", days)
	} else if d >= time.Hour {
		hours := float32(d) / float32(time.Hour)
		return fmt.Sprintf("%.2fh", hours)
	} else if d >= time.Minute {
		minutes := float32(d) / float32(time.Minute)
		return fmt.Sprintf("%.2fm", minutes)
	} else {
		seconds := float32(d) / float32(time.Second)
		return fmt.Sprintf("%.2fs", seconds)
	}
}

// Truncate long path strings and replace /home/user with ~/
func truncatePath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return "<No Name>"
	}

	home, err := os.UserHomeDir()
	if err != nil {
		home = ""
	}

	if strings.HasPrefix(path, home) {
		rpath, err := filepath.Rel(home, path)
		if err == nil {
			path = "~" + string(filepath.Separator) + rpath
		}
	}

	parts := strings.Split(path, string(filepath.Separator))
	if len(parts) > 7 {
		path = strings.Join(parts[len(parts)-7:], string(filepath.Separator))
	}

	return path
}

// DOC:
// FIX: potential time zone issue (hour timeframe is empty)
func isTimestampInRange(timestampStr, rangeType string) (bool, error) {
	layout := "2006-01-02 15:04:05"

	timestamp, err := time.Parse(layout, timestampStr)
	if err != nil {
		return false, fmt.Errorf("error parsing timestamp: %v", err)
	}

	now := time.Now()

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
		startOfRange = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, now.Location())
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
