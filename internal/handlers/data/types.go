package data

import (
	"time"
)

// PendulumMetric represents aggregated metrics for a specific column/field
type PendulumMetric struct {
	Name  string
	Index int
	Value map[string]*PendulumEntry
}

// PendulumEntry represents aggregated data for a specific value within a metric
type PendulumEntry struct {
	ID               string
	ActiveCount      uint
	TotalCount       uint
	ActiveTime       time.Duration
	TotalTime        time.Duration
	ActiveTimestamps []string
	Timestamps       []string
	ActivePct        float64
}

// MetricsParams holds parameters for metrics generation
type MetricsParams struct {
	LogFile               string
	TopN                  int
	TimeRange             string
	ReportExcludes        map[string][]string
	ReportSectionExcludes []string
	TimeFormat            string
	TimeZone              string
	TimeoutLen            float64
}

// CSVColumns maps column names to their indices in the CSV
var CSVColumns = map[string]int{
	"active":    0,
	"branch":    1,
	"directory": 2,
	"file":      3,
	"filetype":  4,
	"project":   5,
	"time":      6,
}

// MetricsError represents errors that can occur during metrics processing
type MetricsError struct {
	Type    ErrorType
	Message string
	Cause   error
}

func (e *MetricsError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

// ErrorType defines the category of metrics processing errors
type ErrorType int

const (
	ErrInvalidData ErrorType = iota
	ErrFileNotFound
	ErrParsingFailed
	ErrProcessingTimeout
	ErrInvalidParameters
)

// ProcessingResult holds the results of metrics processing
type ProcessingResult struct {
	Metrics   []PendulumMetric
	Processed int
	Duration  time.Duration
}
