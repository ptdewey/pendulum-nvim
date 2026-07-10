package data

import (
	"context"
	"testing"
	"time"
)

func intervalParams() *MetricsParams {
	return &MetricsParams{TimeRange: "all", TimeZone: "UTC", TimeoutLen: 180, ReportExcludes: map[string][]string{}}
}

func rows(values ...[]string) [][]string {
	return append([][]string{{"active", "branch", "directory", "file", "filetype", "project", "time"}}, values...)
}

func fileMetric(t *testing.T, result *ProcessingResult) PendulumMetric {
	t.Helper()
	for _, metric := range result.Metrics {
		if metric.Name == "file" {
			return metric
		}
	}
	t.Fatal("file metric missing")
	return PendulumMetric{}
}

func TestIntervalsAreGlobalAndOwnedByPrecedingSample(t *testing.T) {
	result, err := NewMetricsAggregator(intervalParams()).AggregatePendulumMetrics(context.Background(), rows(
		[]string{"true", "main", "/", "A", "go", "p", "2024-01-01 10:00:00"},
		[]string{"true", "main", "/", "B", "go", "p", "2024-01-01 10:01:00"},
		[]string{"true", "main", "/", "A", "go", "p", "2024-01-01 10:02:00"},
		[]string{"true", "main", "/", "B", "go", "p", "2024-01-01 10:03:00"},
	))
	if err != nil {
		t.Fatal(err)
	}
	m := fileMetric(t, result)
	if got := m.Value["A"].TotalTime; got != 2*time.Minute {
		t.Fatalf("A = %v, want 2m", got)
	}
	if got := m.Value["B"].TotalTime; got != time.Minute {
		t.Fatalf("B = %v, want 1m", got)
	}
}

func TestInactiveSampleDoesNotBridgeActiveTime(t *testing.T) {
	result, err := NewMetricsAggregator(intervalParams()).AggregatePendulumMetrics(context.Background(), rows(
		[]string{"true", "main", "/", "A", "go", "p", "2024-01-01 10:00:00"},
		[]string{"false", "main", "/", "A", "go", "p", "2024-01-01 10:01:00"},
		[]string{"true", "main", "/", "A", "go", "p", "2024-01-01 10:02:00"},
	))
	if err != nil {
		t.Fatal(err)
	}
	e := fileMetric(t, result).Value["A"]
	if e.TotalTime != 2*time.Minute || e.ActiveTime != time.Minute {
		t.Fatalf("total=%v active=%v", e.TotalTime, e.ActiveTime)
	}
}

func TestInvalidAndOutOfOrderRowsAreRejectedWithoutBridging(t *testing.T) {
	result, err := NewMetricsAggregator(intervalParams()).AggregatePendulumMetrics(context.Background(), rows(
		[]string{"true", "main", "/", "A", "go", "p", "2024-01-01 10:00:00"},
		[]string{"not-a-bool", "main", "/", "A", "go", "p", "2024-01-01 10:01:00"},
		[]string{"true", "main", "/", "A", "go", "p", "2024-01-01 10:02:00"},
		[]string{"true", "main", "/", "A", "go", "p", "2024-01-01 10:01:00"},
		[]string{"true", "main", "/", "A", "go", "p", "2024-01-01 10:03:00"},
	))
	if err != nil {
		t.Fatal(err)
	}
	if result.Processed != 3 || result.Rejected != 2 {
		t.Fatalf("accepted=%d rejected=%d, want 3/2", result.Processed, result.Rejected)
	}
	if entry := fileMetric(t, result).Value["A"]; entry != nil && entry.TotalTime != 0 {
		t.Fatalf("invalid rows must be barriers; got %v", entry.TotalTime)
	}
}

func TestDuplicateAndTimedOutIntervalsDoNotCreateDuration(t *testing.T) {
	result, err := NewMetricsAggregator(intervalParams()).AggregatePendulumMetrics(context.Background(), rows(
		[]string{"true", "main", "/", "A", "go", "p", "2024-01-01 10:00:00"},
		[]string{"true", "main", "/", "A", "go", "p", "2024-01-01 10:00:00"},
		[]string{"true", "main", "/", "A", "go", "p", "2024-01-01 10:10:00"},
	))
	if err != nil {
		t.Fatal(err)
	}
	if entry := fileMetric(t, result).Value["A"]; entry != nil && entry.TotalTime != 0 {
		t.Fatalf("duration = %v, want 0", entry.TotalTime)
	}
}

func TestEveryMetricPreservesActiveTimeInvariant(t *testing.T) {
	result, err := NewMetricsAggregator(intervalParams()).AggregatePendulumMetrics(context.Background(), rows(
		[]string{"true", "main", "/", "A", "go", "p", "2024-01-01 10:00:00"},
		[]string{"false", "main", "/", "B", "go", "p", "2024-01-01 10:01:00"},
		[]string{"true", "main", "/", "A", "go", "p", "2024-01-01 10:02:00"},
		[]string{"true", "main", "/", "B", "go", "p", "2024-01-01 10:03:00"},
	))
	if err != nil {
		t.Fatal(err)
	}
	for _, metric := range result.Metrics {
		for id, entry := range metric.Value {
			if entry.ActiveTime > entry.TotalTime {
				t.Fatalf("%s/%s: active %v > total %v", metric.Name, id, entry.ActiveTime, entry.TotalTime)
			}
		}
	}
}

func TestHoursSplitIntervalsAtHourBoundary(t *testing.T) {
	result, err := NewMetricsAggregator(intervalParams()).AggregatePendulumHours(context.Background(), rows(
		[]string{"true", "main", "/", "A", "go", "p", "2024-01-01 10:59:00"},
		[]string{"true", "main", "/", "A", "go", "p", "2024-01-01 11:01:00"},
	))
	if err != nil {
		t.Fatal(err)
	}
	if result.Hours.TotalTimeHours[10] != time.Minute || result.Hours.TotalTimeHours[11] != time.Minute {
		t.Fatalf("hours: %#v", result.Hours.TotalTimeHours)
	}
}
