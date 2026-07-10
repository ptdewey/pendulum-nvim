package params

import (
	"testing"

	"github.com/ptdewey/pendulum-server/internal/handlers/data"
)

func TestParseMetricsParams(t *testing.T) {
	tests := []struct {
		name        string
		args        []any
		expectError bool
		validate    func(*data.MetricsParams) error
	}{
		{
			name:        "empty args",
			args:        []any{},
			expectError: true,
		},
		{
			name:        "invalid arg type",
			args:        []any{"not a map"},
			expectError: true,
		},
		{
			name: "missing log_file is ok",
			args: []any{
				map[string]any{
					"top_n": 5,
				},
			},
			expectError: false,
			validate: func(p *data.MetricsParams) error {
				if p.LogFile != "" {
					t.Errorf("Expected empty log_file, got '%s'", p.LogFile)
				}
				return nil
			},
		},
		{
			name: "minimal valid args",
			args: []any{
				map[string]any{
					"log_file": "/path/to/log.csv",
				},
			},
			expectError: false,
			validate: func(p *data.MetricsParams) error {
				if p.LogFile != "/path/to/log.csv" {
					t.Errorf("Expected log_file '/path/to/log.csv', got '%s'", p.LogFile)
				}
				if p.TopN != 5 { // default
					t.Errorf("Expected default top_n 5, got %d", p.TopN)
				}
				if p.TimeRange != "all" { // default
					t.Errorf("Expected default time_range 'all', got '%s'", p.TimeRange)
				}
				return nil
			},
		},
		{
			name: "full valid args",
			args: []any{
				map[string]any{
					"log_file":    "/path/to/log.csv",
					"top_n":       10,
					"time_range":  "day",
					"timeout_len": 120.0,
					"time_format": "24h",
					"time_zone":   "America/New_York",
					"report_excludes": map[string]any{
						"file": []any{"test.*", ".*\\.tmp"},
					},
					"report_section_excludes": []any{"branch", "project"},
				},
			},
			expectError: false,
			validate: func(p *data.MetricsParams) error {
				if p.LogFile != "/path/to/log.csv" {
					t.Errorf("Expected log_file '/path/to/log.csv', got '%s'", p.LogFile)
				}
				if p.TopN != 10 {
					t.Errorf("Expected top_n 10, got %d", p.TopN)
				}
				if p.TimeRange != "day" {
					t.Errorf("Expected time_range 'day', got '%s'", p.TimeRange)
				}
				if p.TimeoutLen != 120.0 {
					t.Errorf("Expected timeout_len 120.0, got %f", p.TimeoutLen)
				}
				if p.TimeFormat != "24h" {
					t.Errorf("Expected time_format '24h', got '%s'", p.TimeFormat)
				}
				if p.TimeZone != "America/New_York" {
					t.Errorf("Expected time_zone 'America/New_York', got '%s'", p.TimeZone)
				}

				// Check report excludes
				fileExcludes := p.ReportExcludes["file"]
				if len(fileExcludes) != 2 {
					t.Errorf("Expected 2 file excludes, got %d", len(fileExcludes))
				}

				// Check section excludes
				if len(p.ReportSectionExcludes) != 2 {
					t.Errorf("Expected 2 section excludes, got %d", len(p.ReportSectionExcludes))
				}

				return nil
			},
		},
		{
			name: "invalid top_n type",
			args: []any{
				map[string]any{
					"log_file": "/path/to/log.csv",
					"top_n":    "not a number",
				},
			},
			expectError: true,
		},
		{
			name: "invalid timeout_len type",
			args: []any{
				map[string]any{
					"log_file":    "/path/to/log.csv",
					"timeout_len": "not a number",
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params, err := ParseMetricsParams(tt.args)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if tt.validate != nil {
				if err := tt.validate(params); err != nil {
					t.Error(err)
				}
			}
		})
	}
}

func TestValidateParams(t *testing.T) {
	tests := []struct {
		name        string
		params      *data.MetricsParams
		expectError bool
	}{
		{
			name: "valid params",
			params: &data.MetricsParams{
				LogFile:    "/path/to/log.csv",
				TopN:       5,
				TimeRange:  "all",
				TimeoutLen: 180.0,
			},
			expectError: false,
		},
		{
			name: "empty log file is ok",
			params: &data.MetricsParams{
				LogFile:    "",
				TopN:       5,
				TimeRange:  "all",
				TimeoutLen: 180.0,
			},
			expectError: false,
		},
		{
			name: "invalid top_n",
			params: &data.MetricsParams{
				LogFile:    "/path/to/log.csv",
				TopN:       0,
				TimeRange:  "all",
				TimeoutLen: 180.0,
			},
			expectError: true,
		},
		{
			name: "negative timeout",
			params: &data.MetricsParams{
				LogFile:    "/path/to/log.csv",
				TopN:       5,
				TimeRange:  "all",
				TimeoutLen: -1.0,
			},
			expectError: true,
		},
		{
			name: "invalid time range",
			params: &data.MetricsParams{
				LogFile:    "/path/to/log.csv",
				TopN:       5,
				TimeRange:  "invalid",
				TimeoutLen: 180.0,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateParams(tt.params)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestParseMetricsParams_NumberTypes(t *testing.T) {
	// Test different number types that might come from JSON/LSP
	tests := []struct {
		name     string
		value    any
		expected int
	}{
		{"int", 10, 10},
		{"int64", int64(15), 15},
		{"float64", 20.0, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := []any{
				map[string]any{
					"log_file": "/test.csv",
					"top_n":    tt.value,
				},
			}

			params, err := ParseMetricsParams(args)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if params.TopN != tt.expected {
				t.Errorf("Expected top_n %d, got %d", tt.expected, params.TopN)
			}
		})
	}
}
