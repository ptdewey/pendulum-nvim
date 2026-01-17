package data

import (
	"testing"
	"time"
)

func TestTimeDiff(t *testing.T) {
	tests := []struct {
		name        string
		timestamps  []string
		timeoutLen  float64
		clamp       bool
		expected    time.Duration
		expectError bool
	}{
		{
			name:       "empty timestamps",
			timestamps: []string{},
			timeoutLen: 180.0,
			clamp:      false,
			expected:   0,
		},
		{
			name:       "single timestamp",
			timestamps: []string{"2024-01-01 10:00:00"},
			timeoutLen: 180.0,
			clamp:      false,
			expected:   0,
		},
		{
			name:       "normal difference",
			timestamps: []string{"2024-01-01 10:00:00", "2024-01-01 10:01:00"},
			timeoutLen: 180.0,
			clamp:      false,
			expected:   time.Minute,
		},
		{
			name:       "exceeds timeout",
			timestamps: []string{"2024-01-01 10:00:00", "2024-01-01 12:00:00"},
			timeoutLen: 180.0,
			clamp:      false,
			expected:   0, // Should return 0 because it exceeds timeout
		},
		{
			name:        "invalid timestamp format",
			timestamps:  []string{"invalid", "2024-01-01 10:01:00"},
			timeoutLen:  180.0,
			clamp:       false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := TimeDiff(tt.timestamps, tt.timeoutLen, tt.clamp)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestCompileRegexPatterns(t *testing.T) {
	tests := []struct {
		name        string
		filters     []string
		expectError bool
		expectCount int
	}{
		{
			name:        "empty filters",
			filters:     []string{},
			expectError: false,
			expectCount: 0,
		},
		{
			name:        "valid patterns",
			filters:     []string{"test.*", "^main$"},
			expectError: false,
			expectCount: 2,
		},
		{
			name:        "invalid pattern",
			filters:     []string{"[invalid"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patterns, err := CompileRegexPatterns(tt.filters)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(patterns) != tt.expectCount {
				t.Errorf("Expected %d patterns, got %d", tt.expectCount, len(patterns))
			}
		})
	}
}

func TestIsExcluded(t *testing.T) {
	patterns, err := CompileRegexPatterns([]string{"test.*", "^main$"})
	if err != nil {
		t.Fatalf("Failed to compile patterns: %v", err)
	}

	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{"matches test pattern", "test123", true},
		{"matches main pattern", "main", true},
		{"no match", "other", false},
		{"partial main match", "mainline", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsExcluded(tt.value, patterns)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestIsTimestampInRange(t *testing.T) {
	tests := []struct {
		name        string
		timestamp   string
		rangeType   string
		timeZone    string
		expectError bool
		// Note: We can't easily test the actual range logic without mocking time.Now()
		// so we focus on error conditions and "all" range
	}{
		{
			name:      "all range always true",
			timestamp: "2024-01-01 10:00:00",
			rangeType: "all",
			timeZone:  "UTC",
		},
		{
			name:        "invalid timestamp",
			timestamp:   "invalid",
			rangeType:   "day",
			timeZone:    "UTC",
			expectError: true,
		},
		{
			name:        "invalid range type",
			timestamp:   "2024-01-01 10:00:00",
			rangeType:   "invalid",
			timeZone:    "UTC",
			expectError: true,
		},
		{
			name:      "valid day range",
			timestamp: "2024-01-01 10:00:00",
			rangeType: "day",
			timeZone:  "UTC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := IsTimestampInRange(tt.timestamp, tt.rangeType, tt.timeZone)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if tt.rangeType == "all" && !result {
				t.Error("Expected 'all' range to return true")
			}
		})
	}
}

func TestIsTimestampInRange_Boundaries(t *testing.T) {
	// Test boundary conditions using fixed times relative to a known "now"
	// We'll test the logic by creating timestamps that should definitely be in or out of range

	now := time.Now()
	loc := time.UTC
	layout := "2006-01-02 15:04:05"

	// Generate test timestamps
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	midToday := time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, loc)
	yesterday := startOfToday.AddDate(0, 0, -1)
	tomorrow := startOfToday.AddDate(0, 0, 1)

	tests := []struct {
		name      string
		timestamp time.Time
		rangeType string
		expected  bool
	}{
		// Day tests
		{"start of today is in day range", startOfToday, "day", true},
		{"mid today is in day range", midToday, "day", true},
		{"yesterday is not in day range", yesterday, "day", false},
		{"tomorrow is not in day range", tomorrow, "day", false},

		// Week tests (last 7 days including today)
		{"today is in week range", midToday, "week", true},
		{"6 days ago is in week range", startOfToday.AddDate(0, 0, -6).Add(time.Hour), "week", true},
		{"8 days ago is not in week range", startOfToday.AddDate(0, 0, -8), "week", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timestampStr := tt.timestamp.In(loc).Format(layout)
			result, err := IsTimestampInRange(timestampStr, tt.rangeType, "UTC")

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if result != tt.expected {
				t.Errorf("For timestamp %s with range %s: expected %v, got %v",
					timestampStr, tt.rangeType, tt.expected, result)
			}
		})
	}
}

func BenchmarkTimeDiff(b *testing.B) {
	timestamps := []string{"2024-01-01 10:00:00", "2024-01-01 10:01:00"}

	b.ResetTimer()
	for b.Loop() {
		_, _ = TimeDiff(timestamps, 180.0, false)
	}
}

func BenchmarkCompileRegexPatterns(b *testing.B) {
	patterns := []string{"test.*", "^main$", ".*\\.go$"}

	b.ResetTimer()
	for b.Loop() {
		_, _ = CompileRegexPatterns(patterns)
	}
}

func TestTimeRangeFilter(t *testing.T) {
	t.Run("all range", func(t *testing.T) {
		filter, err := NewTimeRangeFilter("all", "UTC")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Any timestamp should be in range for "all"
		inRange, err := filter.InRange("2020-01-01 00:00:00")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if !inRange {
			t.Error("Expected 'all' filter to include any timestamp")
		}
	})

	t.Run("invalid range type", func(t *testing.T) {
		_, err := NewTimeRangeFilter("invalid", "UTC")
		if err == nil {
			t.Error("Expected error for invalid range type")
		}
	})

	t.Run("invalid timezone falls back to UTC", func(t *testing.T) {
		filter, err := NewTimeRangeFilter("day", "Invalid/Timezone")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		// Should not error, just use UTC
		if filter.loc.String() != "UTC" {
			t.Errorf("Expected UTC fallback, got %s", filter.loc.String())
		}
	})
}

func TestTimeRangeFilter_Boundaries(t *testing.T) {
	loc := time.UTC
	layout := "2006-01-02 15:04:05"
	now := time.Now().In(loc)

	// Generate boundary timestamps
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	endOfToday := startOfToday.Add(24*time.Hour - time.Nanosecond)
	midToday := startOfToday.Add(12 * time.Hour)
	yesterday := startOfToday.AddDate(0, 0, -1)
	tomorrow := startOfToday.AddDate(0, 0, 1)

	startOfWeek := startOfToday.AddDate(0, 0, -6)
	beforeWeek := startOfWeek.Add(-time.Hour)

	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
	lastMonth := startOfMonth.AddDate(0, 0, -1)

	startOfYear := time.Date(now.Year(), time.January, 1, 0, 0, 0, 0, loc)
	lastYear := startOfYear.AddDate(0, 0, -1)

	tests := []struct {
		name      string
		rangeType string
		timestamp time.Time
		expected  bool
	}{
		// Day range tests
		{"day: start of today (inclusive)", "day", startOfToday, true},
		{"day: mid today", "day", midToday, true},
		{"day: end of today", "day", endOfToday, true},
		{"day: yesterday excluded", "day", yesterday, false},
		{"day: tomorrow excluded", "day", tomorrow, false},

		// Week range tests (last 7 days)
		{"week: today included", "week", midToday, true},
		{"week: start of week (inclusive)", "week", startOfWeek, true},
		{"week: 6 days ago mid-day", "week", startOfWeek.Add(12 * time.Hour), true},
		{"week: before week excluded", "week", beforeWeek, false},

		// Month range tests
		{"month: today included", "month", midToday, true},
		{"month: start of month (inclusive)", "month", startOfMonth, true},
		{"month: last month excluded", "month", lastMonth, false},

		// Year range tests
		{"year: today included", "year", midToday, true},
		{"year: start of year (inclusive)", "year", startOfYear, true},
		{"year: last year excluded", "year", lastYear, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter, err := NewTimeRangeFilter(tt.rangeType, "UTC")
			if err != nil {
				t.Fatalf("Failed to create filter: %v", err)
			}

			timestampStr := tt.timestamp.Format(layout)
			result, err := filter.InRange(timestampStr)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if result != tt.expected {
				t.Errorf("For %s with range %s: expected %v, got %v (timestamp: %s, start: %s, end: %s)",
					tt.name, tt.rangeType, tt.expected, result,
					timestampStr,
					filter.startOfRange.Format(layout),
					filter.endOfRange.Format(layout))
			}
		})
	}
}

func TestTimeRangeFilter_Timezones(t *testing.T) {
	// Test that timezone handling works correctly
	layout := "2006-01-02 15:04:05"

	// Create a filter for Eastern time
	filter, err := NewTimeRangeFilter("day", "America/New_York")
	if err != nil {
		t.Fatalf("Failed to create filter: %v", err)
	}

	// Get current time in Eastern
	eastern, _ := time.LoadLocation("America/New_York")
	now := time.Now().In(eastern)
	midToday := time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, eastern)

	// This timestamp should be in range
	timestampStr := midToday.Format(layout)
	inRange, err := filter.InRange(timestampStr)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !inRange {
		t.Errorf("Expected mid-day Eastern timestamp to be in range")
	}
}

func TestTimeRangeFilter_InvalidTimestamp(t *testing.T) {
	filter, err := NewTimeRangeFilter("day", "UTC")
	if err != nil {
		t.Fatalf("Failed to create filter: %v", err)
	}

	_, err = filter.InRange("not-a-timestamp")
	if err == nil {
		t.Error("Expected error for invalid timestamp")
	}
}

func BenchmarkTimeRangeFilter_InRange(b *testing.B) {
	filter, _ := NewTimeRangeFilter("week", "UTC")
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	b.ResetTimer()
	for b.Loop() {
		_, _ = filter.InRange(timestamp)
	}
}

func BenchmarkTimeRangeFilter_VsIsTimestampInRange(b *testing.B) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	b.Run("TimeRangeFilter (reused)", func(b *testing.B) {
		filter, _ := NewTimeRangeFilter("week", "America/New_York")
		b.ResetTimer()
		for b.Loop() {
			_, _ = filter.InRange(timestamp)
		}
	})

	b.Run("IsTimestampInRange (creates filter each time)", func(b *testing.B) {
		b.ResetTimer()
		for b.Loop() {
			_, _ = IsTimestampInRange(timestamp, "week", "America/New_York")
		}
	})
}
