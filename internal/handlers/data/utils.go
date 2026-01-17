package data

import (
	"fmt"
	"regexp"
	"time"
)

// timeDiff calculates the time difference between the last two timestamps
func TimeDiff(timestamps []string, timeoutLen float64, clamp bool) (time.Duration, error) {
	n := len(timestamps)
	if n < 2 {
		return time.Duration(0), nil
	}

	curr, prev := timestamps[n-1], timestamps[n-2]
	var d time.Duration
	var err error

	if !clamp {
		d, err = calcDuration(curr, prev)
	} else {
		d, err = calcDurationWithinHour(curr, prev)
	}

	if err != nil {
		return time.Duration(0), &MetricsError{
			Type:    ErrParsingFailed,
			Message: "failed to calculate time difference",
			Cause:   err,
		}
	}

	// If difference exceeds timeout, editor was closed between sessions
	if d.Seconds() > timeoutLen {
		return time.Duration(0), nil
	}

	return d, nil
}

// calcDuration calculates the duration between two string timestamps (curr - prev)
func calcDuration(curr string, prev string) (time.Duration, error) {
	layout := "2006-01-02 15:04:05"

	currT, err := time.Parse(layout, curr)
	if err != nil {
		return time.Duration(0), fmt.Errorf("failed to parse current timestamp %s: %w", curr, err)
	}

	prevT, err := time.Parse(layout, prev)
	if err != nil {
		return time.Duration(0), fmt.Errorf("failed to parse previous timestamp %s: %w", prev, err)
	}

	return currT.Sub(prevT), nil
}

// calcDurationWithinHour calculates duration clamped to the hour boundary
func calcDurationWithinHour(curr string, prev string) (time.Duration, error) {
	layout := "2006-01-02 15:04:05"

	currT, err := time.Parse(layout, curr)
	if err != nil {
		return 0, fmt.Errorf("failed to parse current timestamp %s: %w", curr, err)
	}

	prevT, err := time.Parse(layout, prev)
	if err != nil {
		return 0, fmt.Errorf("failed to parse previous timestamp %s: %w", prev, err)
	}

	// If prev_t is within the same hour as curr_t, return the direct difference
	if prevT.Hour() == currT.Hour() && prevT.Day() == currT.Day() {
		return currT.Sub(prevT), nil
	}

	// Otherwise, clamp prev_t to the start of curr_t's hour
	clampedPrevT := time.Date(currT.Year(), currT.Month(), currT.Day(), currT.Hour(),
		0, 0, 0, currT.Location())

	return currT.Sub(clampedPrevT), nil
}

// CompileRegexPatterns compiles regex patterns for exclusion filtering
func CompileRegexPatterns(filters []string) ([]*regexp.Regexp, error) {
	if len(filters) == 0 {
		return nil, nil
	}

	patterns := make([]*regexp.Regexp, 0, len(filters))
	for _, expr := range filters {
		r, err := regexp.Compile(expr)
		if err != nil {
			return nil, &MetricsError{
				Type:    ErrInvalidParameters,
				Message: fmt.Sprintf("failed to compile regex pattern: %s", expr),
				Cause:   err,
			}
		}
		patterns = append(patterns, r)
	}

	return patterns, nil
}

// IsExcluded checks if a value matches any exclusion pattern
func IsExcluded(val string, patterns []*regexp.Regexp) bool {
	for _, r := range patterns {
		if r.MatchString(val) {
			return true
		}
	}
	return false
}

// TimeRangeFilter provides efficient time range filtering with pre-computed boundaries
type TimeRangeFilter struct {
	startOfRange time.Time
	endOfRange   time.Time
	loc          *time.Location
	layout       string
	isAll        bool
}

// NewTimeRangeFilter creates a filter with pre-computed time boundaries
func NewTimeRangeFilter(rangeType, timeZone string) (*TimeRangeFilter, error) {
	if rangeType == "all" {
		return &TimeRangeFilter{isAll: true}, nil
	}

	loc, err := time.LoadLocation(timeZone)
	if err != nil {
		loc = time.UTC
	}

	now := time.Now().In(loc)
	var startOfRange, endOfRange time.Time

	switch rangeType {
	case "today", "day":
		startOfRange = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
		endOfRange = startOfRange.Add(24 * time.Hour)
	case "year":
		startOfRange = time.Date(now.Year(), time.January, 1, 0, 0, 0, 0, loc)
		endOfRange = startOfRange.AddDate(1, 0, 0)
	case "month":
		startOfRange = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
		endOfRange = startOfRange.AddDate(0, 1, 0)
	case "week":
		startOfRange = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc).AddDate(0, 0, -6)
		endOfRange = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc).Add(24 * time.Hour)
	case "hour":
		startOfRange = now.Truncate(time.Hour)
		endOfRange = startOfRange.Add(time.Hour)
	default:
		return nil, &MetricsError{
			Type:    ErrInvalidParameters,
			Message: fmt.Sprintf("unsupported time range: %s", rangeType),
		}
	}

	return &TimeRangeFilter{
		startOfRange: startOfRange,
		endOfRange:   endOfRange,
		loc:          loc,
		layout:       "2006-01-02 15:04:05",
		isAll:        false,
	}, nil
}

// InRange checks if a timestamp string falls within the pre-computed range
// Note: timestamps in the CSV are stored in UTC, so we parse them as UTC
// and compare against the range boundaries (which are also in UTC internally)
func (f *TimeRangeFilter) InRange(timestampStr string) (bool, error) {
	if f.isAll {
		return true, nil
	}

	// Timestamps in CSV are always in UTC
	timestamp, err := time.Parse(f.layout, timestampStr)
	if err != nil {
		return false, err
	}

	return !timestamp.Before(f.startOfRange) && timestamp.Before(f.endOfRange), nil
}

// IsTimestampInRange checks if a timestamp falls within the specified time range
// Deprecated: Use TimeRangeFilter for better performance when checking multiple timestamps
func IsTimestampInRange(timestampStr, rangeType, timeZone string) (bool, error) {
	filter, err := NewTimeRangeFilter(rangeType, timeZone)
	if err != nil {
		return false, err
	}
	return filter.InRange(timestampStr)
}
