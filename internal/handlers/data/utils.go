package data

import (
	"fmt"
	"regexp"
	"time"
)

const timestampLayout = "2006-01-02 15:04:05"

func ParseTimestamp(value string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t.UTC(), nil
	}
	t, err := time.ParseInLocation(timestampLayout, value, time.UTC)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse timestamp %q: %w", value, err)
	}
	return t.UTC(), nil
}
func TimeDiff(timestamps []string, timeoutLen float64, _ bool) (time.Duration, error) {
	if len(timestamps) < 2 {
		return 0, nil
	}
	current, err := ParseTimestamp(timestamps[len(timestamps)-1])
	if err != nil {
		return 0, err
	}
	previous, err := ParseTimestamp(timestamps[len(timestamps)-2])
	if err != nil {
		return 0, err
	}
	d := current.Sub(previous)
	if d < 0 {
		return 0, fmt.Errorf("timestamps out of order")
	}
	if d.Seconds() > timeoutLen {
		return 0, nil
	}
	return d, nil
}
func CompileRegexPatterns(filters []string) ([]*regexp.Regexp, error) {
	patterns := make([]*regexp.Regexp, 0, len(filters))
	for _, expr := range filters {
		r, err := regexp.Compile(expr)
		if err != nil {
			return nil, &MetricsError{Type: ErrInvalidParameters, Message: fmt.Sprintf("failed to compile regex pattern: %s", expr), Cause: err}
		}
		patterns = append(patterns, r)
	}
	return patterns, nil
}
func IsExcluded(val string, patterns []*regexp.Regexp) bool {
	for _, r := range patterns {
		if r.MatchString(val) {
			return true
		}
	}
	return false
}

type TimeRangeFilter struct {
	startOfRange, endOfRange time.Time
	loc                      *time.Location
	isAll                    bool
}

func NewTimeRangeFilter(rangeType, timeZone string) (*TimeRangeFilter, error) {
	if rangeType == "all" {
		return &TimeRangeFilter{isAll: true}, nil
	}
	loc, err := time.LoadLocation(timeZone)
	if err != nil {
		return nil, &MetricsError{Type: ErrInvalidParameters, Message: "load timezone", Cause: err}
	}
	now := time.Now().In(loc)
	day := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	var start, end time.Time
	switch rangeType {
	case "today", "day":
		start, end = day, day.AddDate(0, 0, 1)
	case "week":
		start, end = day.AddDate(0, 0, -6), day.AddDate(0, 0, 1)
	case "month":
		start, end = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc), time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, loc)
	case "year":
		start, end = time.Date(now.Year(), time.January, 1, 0, 0, 0, 0, loc), time.Date(now.Year()+1, time.January, 1, 0, 0, 0, 0, loc)
	case "hour":
		start, end = now.Truncate(time.Hour), now.Truncate(time.Hour).Add(time.Hour)
	default:
		return nil, &MetricsError{Type: ErrInvalidParameters, Message: fmt.Sprintf("unsupported time range: %s", rangeType)}
	}
	return &TimeRangeFilter{startOfRange: start.UTC(), endOfRange: end.UTC(), loc: loc}, nil
}
func (f *TimeRangeFilter) InRange(timestamp string) (bool, error) {
	t, err := ParseTimestamp(timestamp)
	if err != nil {
		return false, err
	}
	return f.isAll || (!t.Before(f.startOfRange) && t.Before(f.endOfRange)), nil
}
func (f *TimeRangeFilter) Clip(start, end time.Time) (interval, bool) {
	if f.isAll {
		return interval{start: start, end: end}, start.Before(end)
	}
	if start.Before(f.startOfRange) {
		start = f.startOfRange
	}
	if end.After(f.endOfRange) {
		end = f.endOfRange
	}
	return interval{start: start, end: end}, start.Before(end)
}
func IsTimestampInRange(timestamp, rangeType, timeZone string) (bool, error) {
	f, err := NewTimeRangeFilter(rangeType, timeZone)
	if err != nil {
		return false, err
	}
	return f.InRange(timestamp)
}
