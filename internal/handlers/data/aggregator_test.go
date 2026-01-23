package data

import (
	"context"
	"fmt"
	"testing"
)

func TestNewMetricsAggregator(t *testing.T) {
	params := &MetricsParams{
		LogFile:               "test.csv",
		TopN:                  5,
		TimeRange:             "all",
		ReportExcludes:        make(map[string][]string),
		ReportSectionExcludes: []string{},
		TimeZone:              "UTC",
		TimeoutLen:            180.0,
	}

	aggregator := NewMetricsAggregator(params)

	if aggregator.params != params {
		t.Error("Expected params to be set correctly")
	}

	if aggregator.workerCount <= 0 {
		t.Error("Expected positive worker count")
	}
}

func TestMetricsAggregator_AggregatePendulumMetrics(t *testing.T) {
	params := &MetricsParams{
		TopN:                  5,
		TimeRange:             "all",
		ReportExcludes:        make(map[string][]string),
		ReportSectionExcludes: []string{},
		TimeZone:              "UTC",
		TimeoutLen:            180.0,
	}

	aggregator := NewMetricsAggregator(params)

	tests := []struct {
		name        string
		data        [][]string
		expectError bool
		expectCount int
	}{
		{
			name:        "empty data",
			data:        [][]string{},
			expectError: false,
			expectCount: 0,
		},
		{
			name: "header only",
			data: [][]string{
				{"active", "branch", "directory", "file", "filetype", "project", "time"},
			},
			expectError: false,
			expectCount: 0,
		},
		{
			name: "normal data",
			data: [][]string{
				{"active", "branch", "directory", "file", "filetype", "project", "time"},
				{"true", "main", "/home/user", "test.go", "go", "myproject", "2024-01-01 10:00:00"},
				{"false", "main", "/home/user", "test.go", "go", "myproject", "2024-01-01 10:01:00"},
				{"true", "feature", "/home/user", "main.go", "go", "myproject", "2024-01-01 10:02:00"},
			},
			expectError: false,
			expectCount: 5, // branch, directory, file, filetype, project
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := aggregator.AggregatePendulumMetrics(ctx, tt.data)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(result.Metrics) != tt.expectCount {
				t.Errorf("Expected %d metrics, got %d", tt.expectCount, len(result.Metrics))
			}

			if len(tt.data) > 1 {
				expectedProcessed := len(tt.data) - 1 // Exclude header
				if result.Processed != expectedProcessed {
					t.Errorf("Expected %d processed rows, got %d", expectedProcessed, result.Processed)
				}
			}
		})
	}
}

func TestMetricsAggregator_WithExclusions(t *testing.T) {
	params := &MetricsParams{
		TopN:       5,
		TimeRange:  "all",
		TimeZone:   "UTC",
		TimeoutLen: 180.0,
		ReportExcludes: map[string][]string{
			"file": {"test.*"},
		},
		ReportSectionExcludes: []string{"branch"},
	}

	aggregator := NewMetricsAggregator(params)

	data := [][]string{
		{"active", "branch", "directory", "file", "filetype", "project", "time"},
		{"true", "main", "/home/user", "test.go", "go", "myproject", "2024-01-01 10:00:00"},
		{"true", "main", "/home/user", "main.go", "go", "myproject", "2024-01-01 10:01:00"},
	}

	ctx := context.Background()
	result, err := aggregator.AggregatePendulumMetrics(ctx, data)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should exclude branch section entirely
	for _, metric := range result.Metrics {
		if metric.Name == "branch" {
			t.Error("Expected branch metric to be excluded")
		}

		// For file metric, should exclude test.go but include main.go
		if metric.Name == "file" {
			if _, exists := metric.Value["test.go"]; exists {
				t.Error("Expected test.go to be excluded")
			}
			if _, exists := metric.Value["main.go"]; !exists {
				t.Error("Expected main.go to be included")
			}
		}
	}
}

func TestMetricsAggregator_ContextCancellation(t *testing.T) {
	params := &MetricsParams{
		TopN:                  5,
		TimeRange:             "all",
		ReportExcludes:        make(map[string][]string),
		ReportSectionExcludes: []string{},
		TimeZone:              "UTC",
		TimeoutLen:            180.0,
	}

	aggregator := NewMetricsAggregator(params)

	data := [][]string{
		{"active", "branch", "directory", "file", "filetype", "project", "time"},
		{"true", "main", "/home/user", "test.go", "go", "myproject", "2024-01-01 10:00:00"},
	}

	// Cancel context immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := aggregator.AggregatePendulumMetrics(ctx, data)

	if err == nil {
		t.Error("Expected error for cancelled context")
	}
}

func BenchmarkMetricsAggregation(b *testing.B) {
	params := &MetricsParams{
		TopN:                  5,
		TimeRange:             "all",
		ReportExcludes:        make(map[string][]string),
		ReportSectionExcludes: []string{},
		TimeZone:              "UTC",
		TimeoutLen:            180.0,
	}

	// Create test data with different sizes
	sizes := []int{100, 1000, 10000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("size-%d", size), func(b *testing.B) {
			data := generateTestData(size)
			aggregator := NewMetricsAggregator(params)
			ctx := context.Background()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := aggregator.AggregatePendulumMetrics(ctx, data)
				if err != nil {
					b.Fatalf("Aggregation failed: %v", err)
				}
			}
		})
	}
}

// generateTestData creates test CSV data for benchmarking
func generateTestData(size int) [][]string {
	data := [][]string{
		{"active", "branch", "directory", "file", "filetype", "project", "time"},
	}

	for i := 0; i < size; i++ {
		active := "true"
		if i%2 == 0 {
			active = "false"
		}

		row := []string{
			active,
			"main",
			"/home/user",
			"test.go",
			"go",
			"myproject",
			"2024-01-01 10:00:00",
		}
		data = append(data, row)
	}

	return data
}
