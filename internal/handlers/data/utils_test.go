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
