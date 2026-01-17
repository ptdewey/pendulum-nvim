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
