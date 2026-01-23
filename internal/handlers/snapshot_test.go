package handlers

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/ptdewey/pendulum-server/internal/handlers/data"
	"github.com/ptdewey/pendulum-server/internal/handlers/prettify"
	"github.com/ptdewey/shutter"
)

// Snapshot Testing Documentation:
//
// This file contains snapshot tests for the metrics report generation pipeline.
// Snapshot testing uses shutter (https://github.com/ptdewey/shutter) to capture
// and validate the full formatted output of reports.
//
// Key Features:
// - Deterministic Test Data: Uses testdata/synthetic_log.csv with fixed, unique
//   active time values to ensure consistent, non-flaky test results
// - Scrubbers: Dynamic content (timestamps, file paths) is scrubbed to a placeholder
//   before comparison to avoid false positives when environments differ
// - Full Pipeline Testing: Tests verify the complete flow: CSV read → Aggregation →
//   Formatting → Report generation
// - Isolated Formatter Testing: Direct formatter tests validate output formatting
//   without reliance on CSV data
//
// Test Data Details (testdata/synthetic_log.csv):
// - 106 lines (header + 105 data rows)
// - All entries dated 2024-06-15 with 1-minute intervals
// - Unique active time durations per metric to prevent non-deterministic sorting
// - Branches: main (37m), develop (22m), feature/auth (13m), bugfix/login (7m)
// - Projects: myproject (32m), webapp (22m), api-server (22m), notes (3m)
// - Files with unique durations: auth.go (16m), server.py (11m), app.js (9m), etc.
//
// Maintaining Tests:
// 1. If test output changes unexpectedly, review the diff carefully
// 2. If changes are intentional, update snapshots by running:
//    go test ./internal/handlers -run Snapshot
// 3. Verify the updated snapshots in __snapshots__/ directory
// 4. Ensure all unique time values in test data remain unique to prevent
//    non-deterministic sorting issues
//
// Scrubbers Applied:
// - Timestamp pattern: \d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2} → <TIMESTAMP>
// - File path: .+synthetic_log\.csv → <LOG_FILE_PATH>

// getTestdataPath returns the absolute path to the testdata directory
func getTestdataPath() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "..", "..", "testdata")
}

// TestMetricsReportSnapshot tests the full metrics report output using snapshot testing
func TestMetricsReportSnapshot(t *testing.T) {
	logFile := filepath.Join(getTestdataPath(), "synthetic_log.csv")

	// Create params for the report
	params := &data.MetricsParams{
		LogFile:               logFile,
		TopN:                  5,
		TimeRange:             "all",
		TimeFormat:            "24h",
		TimeZone:              "UTC",
		TimeoutLen:            180, // 3 minutes - entries in test data are 1 min apart
		ReportExcludes:        map[string][]string{},
		ReportSectionExcludes: []string{},
	}

	// Read CSV data
	csvReader := data.NewCSVReader(logFile)
	csvData, err := csvReader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to read CSV: %v", err)
	}

	// Aggregate metrics
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	aggregator := data.NewMetricsAggregator(params)
	result, err := aggregator.AggregatePendulumMetrics(ctx, csvData)
	if err != nil {
		t.Fatalf("Failed to aggregate metrics: %v", err)
	}

	// Format the report
	formatter := prettify.NewMetricsFormatter(params)
	formattedLines := formatter.FormatMetrics(result.Metrics)

	// Join lines into final report
	report := ""
	for _, line := range formattedLines {
		report += line + "\n"
	}

	// Snapshot with scrubbers for dynamic content
	shutter.SnapString(t, "metrics_report_all", report,
		// Scrub the generated timestamp (format: 2006-01-02 15:04:05)
		shutter.ScrubRegex(`\*\*Generated:\*\* \d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}`, "**Generated:** <TIMESTAMP>"),
		// Scrub the log file path which varies by environment
		shutter.ScrubRegex(`\*\*Log File:\*\* .+synthetic_log\.csv`, "**Log File:** <LOG_FILE_PATH>"),
	)
}

// TestMetricsReportSnapshotWithExclusions tests report with section exclusions
func TestMetricsReportSnapshotWithExclusions(t *testing.T) {
	logFile := filepath.Join(getTestdataPath(), "synthetic_log.csv")

	// Create params with exclusions
	params := &data.MetricsParams{
		LogFile:               logFile,
		TopN:                  3,
		TimeRange:             "all",
		TimeFormat:            "24h",
		TimeZone:              "UTC",
		TimeoutLen:            180,
		ReportExcludes:        map[string][]string{},
		ReportSectionExcludes: []string{"branch", "directory"},
	}

	// Read CSV data
	csvReader := data.NewCSVReader(logFile)
	csvData, err := csvReader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to read CSV: %v", err)
	}

	// Aggregate metrics
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	aggregator := data.NewMetricsAggregator(params)
	result, err := aggregator.AggregatePendulumMetrics(ctx, csvData)
	if err != nil {
		t.Fatalf("Failed to aggregate metrics: %v", err)
	}

	// Format the report
	formatter := prettify.NewMetricsFormatter(params)
	formattedLines := formatter.FormatMetrics(result.Metrics)

	// Join lines into final report
	report := ""
	for _, line := range formattedLines {
		report += line + "\n"
	}

	// Snapshot with scrubbers
	shutter.SnapString(t, "metrics_report_exclusions", report,
		shutter.ScrubRegex(`\*\*Generated:\*\* \d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}`, "**Generated:** <TIMESTAMP>"),
		shutter.ScrubRegex(`\*\*Log File:\*\* .+synthetic_log\.csv`, "**Log File:** <LOG_FILE_PATH>"),
	)
}

// TestMetricsReportSnapshotTopN tests report with different TopN values
func TestMetricsReportSnapshotTopN(t *testing.T) {
	logFile := filepath.Join(getTestdataPath(), "synthetic_log.csv")

	// Create params with TopN = 2
	params := &data.MetricsParams{
		LogFile:               logFile,
		TopN:                  2,
		TimeRange:             "all",
		TimeFormat:            "24h",
		TimeZone:              "UTC",
		TimeoutLen:            180,
		ReportExcludes:        map[string][]string{},
		ReportSectionExcludes: []string{},
	}

	// Read CSV data
	csvReader := data.NewCSVReader(logFile)
	csvData, err := csvReader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to read CSV: %v", err)
	}

	// Aggregate metrics
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	aggregator := data.NewMetricsAggregator(params)
	result, err := aggregator.AggregatePendulumMetrics(ctx, csvData)
	if err != nil {
		t.Fatalf("Failed to aggregate metrics: %v", err)
	}

	// Format the report
	formatter := prettify.NewMetricsFormatter(params)
	formattedLines := formatter.FormatMetrics(result.Metrics)

	// Join lines into final report
	report := ""
	for _, line := range formattedLines {
		report += line + "\n"
	}

	// Snapshot with scrubbers
	shutter.SnapString(t, "metrics_report_top2", report,
		shutter.ScrubRegex(`\*\*Generated:\*\* \d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}`, "**Generated:** <TIMESTAMP>"),
		shutter.ScrubRegex(`\*\*Log File:\*\* .+synthetic_log\.csv`, "**Log File:** <LOG_FILE_PATH>"),
	)
}

// TestFormatterSnapshotDirectly tests the formatter output directly without full pipeline
func TestFormatterSnapshotDirectly(t *testing.T) {
	// Create synthetic metrics data directly
	metrics := []data.PendulumMetric{
		{
			Name: "branch",
			Value: map[string]*data.PendulumEntry{
				"main": {
					ID:         "main",
					TotalTime:  30 * time.Minute,
					ActiveTime: 25 * time.Minute,
					ActivePct:  0.833,
				},
				"feature/auth": {
					ID:         "feature/auth",
					TotalTime:  20 * time.Minute,
					ActiveTime: 18 * time.Minute,
					ActivePct:  0.90,
				},
			},
		},
		{
			Name: "project",
			Value: map[string]*data.PendulumEntry{
				"myproject": {
					ID:         "myproject",
					TotalTime:  45 * time.Minute,
					ActiveTime: 40 * time.Minute,
					ActivePct:  0.889,
				},
				"webapp": {
					ID:         "webapp",
					TotalTime:  15 * time.Minute,
					ActiveTime: 12 * time.Minute,
					ActivePct:  0.80,
				},
			},
		},
	}

	params := &data.MetricsParams{
		LogFile:    "/test/pendulum.csv",
		TopN:       5,
		TimeRange:  "all",
		TimeFormat: "24h",
		TimeZone:   "UTC",
	}

	formatter := prettify.NewMetricsFormatter(params)
	formattedLines := formatter.FormatMetrics(metrics)

	report := ""
	for _, line := range formattedLines {
		report += line + "\n"
	}

	shutter.SnapString(t, "formatter_direct", report,
		shutter.ScrubRegex(`\*\*Generated:\*\* \d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}`, "**Generated:** <TIMESTAMP>"),
	)
}

// TestHoursReportSnapshot tests the full hours report output using snapshot testing
func TestHoursReportSnapshot(t *testing.T) {
	logFile := filepath.Join(getTestdataPath(), "synthetic_log.csv")

	// Create params for the report
	params := &data.MetricsParams{
		LogFile:               logFile,
		TopN:                  5,
		TimeRange:             "all",
		TimeFormat:            "24h",
		TimeZone:              "UTC",
		TimeoutLen:            180, // 3 minutes - entries in test data are 1 min apart
		ReportExcludes:        map[string][]string{},
		ReportSectionExcludes: []string{},
	}

	// Read CSV data
	csvReader := data.NewCSVReader(logFile)
	csvData, err := csvReader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to read CSV: %v", err)
	}

	// Aggregate hours
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	aggregator := data.NewMetricsAggregator(params)
	result, err := aggregator.AggregatePendulumHours(ctx, csvData)
	if err != nil {
		t.Fatalf("Failed to aggregate hours: %v", err)
	}

	// Format the report
	formatter := prettify.NewMetricsFormatter(params)
	formattedLines := formatter.FormatHours(result.Hours, params.TopN)

	// Join lines into final report
	report := ""
	for _, line := range formattedLines {
		report += line + "\n"
	}

	// Snapshot with scrubbers for dynamic content
	shutter.SnapString(t, "hours_report_all", report,
		// Scrub the generated timestamp (format: 2006-01-02 15:04:05)
		shutter.ScrubRegex(`\*\*Generated:\*\* \d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}`, "**Generated:** <TIMESTAMP>"),
		// Scrub the log file path which varies by environment
		shutter.ScrubRegex(`\*\*Log File:\*\* .+synthetic_log\.csv`, "**Log File:** <LOG_FILE_PATH>"),
	)
}

// TestHoursReportSnapshot12h tests hours report with 12-hour time format
func TestHoursReportSnapshot12h(t *testing.T) {
	logFile := filepath.Join(getTestdataPath(), "synthetic_log.csv")

	// Create params with 12h format
	params := &data.MetricsParams{
		LogFile:               logFile,
		TopN:                  5,
		TimeRange:             "all",
		TimeFormat:            "12h",
		TimeZone:              "UTC",
		TimeoutLen:            180,
		ReportExcludes:        map[string][]string{},
		ReportSectionExcludes: []string{},
	}

	// Read CSV data
	csvReader := data.NewCSVReader(logFile)
	csvData, err := csvReader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to read CSV: %v", err)
	}

	// Aggregate hours
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	aggregator := data.NewMetricsAggregator(params)
	result, err := aggregator.AggregatePendulumHours(ctx, csvData)
	if err != nil {
		t.Fatalf("Failed to aggregate hours: %v", err)
	}

	// Format the report
	formatter := prettify.NewMetricsFormatter(params)
	formattedLines := formatter.FormatHours(result.Hours, params.TopN)

	// Join lines into final report
	report := ""
	for _, line := range formattedLines {
		report += line + "\n"
	}

	// Snapshot with scrubbers
	shutter.SnapString(t, "hours_report_12h", report,
		shutter.ScrubRegex(`\*\*Generated:\*\* \d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}`, "**Generated:** <TIMESTAMP>"),
		shutter.ScrubRegex(`\*\*Log File:\*\* .+synthetic_log\.csv`, "**Log File:** <LOG_FILE_PATH>"),
	)
}

// TestHoursReportSnapshotTopN tests hours report with different TopN values
func TestHoursReportSnapshotTopN(t *testing.T) {
	logFile := filepath.Join(getTestdataPath(), "synthetic_log.csv")

	// Create params with TopN = 3
	params := &data.MetricsParams{
		LogFile:               logFile,
		TopN:                  3,
		TimeRange:             "all",
		TimeFormat:            "24h",
		TimeZone:              "UTC",
		TimeoutLen:            180,
		ReportExcludes:        map[string][]string{},
		ReportSectionExcludes: []string{},
	}

	// Read CSV data
	csvReader := data.NewCSVReader(logFile)
	csvData, err := csvReader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to read CSV: %v", err)
	}

	// Aggregate hours
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	aggregator := data.NewMetricsAggregator(params)
	result, err := aggregator.AggregatePendulumHours(ctx, csvData)
	if err != nil {
		t.Fatalf("Failed to aggregate hours: %v", err)
	}

	// Format the report
	formatter := prettify.NewMetricsFormatter(params)
	formattedLines := formatter.FormatHours(result.Hours, params.TopN)

	// Join lines into final report
	report := ""
	for _, line := range formattedLines {
		report += line + "\n"
	}

	// Snapshot with scrubbers
	shutter.SnapString(t, "hours_report_top3", report,
		shutter.ScrubRegex(`\*\*Generated:\*\* \d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}`, "**Generated:** <TIMESTAMP>"),
		shutter.ScrubRegex(`\*\*Log File:\*\* .+synthetic_log\.csv`, "**Log File:** <LOG_FILE_PATH>"),
	)
}

// TestHoursFormatterSnapshotDirectly tests the hours formatter output directly without full pipeline
func TestHoursFormatterSnapshotDirectly(t *testing.T) {
	// Create synthetic hours data directly
	hours := &data.PendulumHours{
		ActiveTimestamps: []string{
			"2024-06-15 09:00:00",
			"2024-06-15 09:01:00",
			"2024-06-15 10:00:00",
			"2024-06-15 10:01:00",
			"2024-06-15 10:02:00",
			"2024-06-15 14:00:00",
		},
		Timestamps: []string{
			"2024-06-15 09:00:00",
			"2024-06-15 09:01:00",
			"2024-06-15 09:02:00",
			"2024-06-15 10:00:00",
			"2024-06-15 10:01:00",
			"2024-06-15 10:02:00",
			"2024-06-15 10:03:00",
			"2024-06-15 14:00:00",
			"2024-06-15 14:01:00",
		},
		ActiveTimeHours: map[int]time.Duration{
			9:  10 * time.Minute,
			10: 25 * time.Minute,
			14: 15 * time.Minute,
		},
		ActiveTimeHoursRecent: map[int]time.Duration{
			9:  5 * time.Minute,
			10: 12 * time.Minute,
			14: 8 * time.Minute,
		},
		TotalTimeHours: map[int]time.Duration{
			9:  15 * time.Minute,
			10: 30 * time.Minute,
			14: 20 * time.Minute,
		},
		TotalTimeHoursRecent: map[int]time.Duration{
			9:  8 * time.Minute,
			10: 15 * time.Minute,
			14: 10 * time.Minute,
		},
	}

	params := &data.MetricsParams{
		LogFile:    "/test/pendulum.csv",
		TopN:       5,
		TimeRange:  "all",
		TimeFormat: "24h",
		TimeZone:   "UTC",
	}

	formatter := prettify.NewMetricsFormatter(params)
	formattedLines := formatter.FormatHours(hours, params.TopN)

	report := ""
	for _, line := range formattedLines {
		report += line + "\n"
	}

	shutter.SnapString(t, "hours_formatter_direct", report,
		shutter.ScrubRegex(`\*\*Generated:\*\* \d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}`, "**Generated:** <TIMESTAMP>"),
	)
}
