package handlers

import (
	"os"
	"strings"
	"testing"

	"github.com/ptdewey/pendulum-server/internal/config"
	"github.com/tliron/glsp"
)

// TestGenerateMetricsReport_Integration tests the complete metrics generation flow
func TestGenerateMetricsReport_Integration(t *testing.T) {
	// Create a temporary CSV file with test data
	tmpFile, err := os.CreateTemp("", "pendulum_test*.csv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write comprehensive test data
	testData := `active,branch,directory,file,filetype,project,time
true,main,/home/user/project,main.go,go,myproject,2024-01-01 10:00:00
false,main,/home/user/project,main.go,go,myproject,2024-01-01 10:01:00
true,main,/home/user/project,test.go,go,myproject,2024-01-01 10:02:00
true,feature,/home/user/project,api.go,go,myproject,2024-01-01 10:03:00
false,feature,/home/user/docs,README.md,markdown,myproject,2024-01-01 10:04:00
true,main,/home/user/project,utils.go,go,myproject,2024-01-01 10:05:00`

	if _, err := tmpFile.WriteString(testData); err != nil {
		t.Fatalf("Failed to write test data: %v", err)
	}
	tmpFile.Close()

	tests := []struct {
		name         string
		args         []any
		expectError  bool
		validateFunc func(string) error
	}{
		{
			name: "basic metrics report",
			args: []any{
				map[string]any{
					"log_file": tmpFile.Name(),
					"top_n":    3,
				},
			},
			expectError: false,
			validateFunc: func(result string) error {
				// Check that we got a proper markdown report
				if !strings.Contains(result, "# Pendulum Metrics Report") {
					t.Error("Expected report header")
				}
				if !strings.Contains(result, "## Top") {
					t.Error("Expected metrics sections")
				}
				if !strings.Contains(result, "Processing Summary") {
					t.Error("Expected processing summary")
				}
				return nil
			},
		},
		{
			name: "metrics with exclusions",
			args: []any{
				map[string]any{
					"log_file": tmpFile.Name(),
					"top_n":    2,
					"report_excludes": map[string]any{
						"file": []any{"test.*"},
					},
					"report_section_excludes": []any{"branch"},
				},
			},
			expectError: false,
			validateFunc: func(result string) error {
				// Should not contain branch metrics
				if strings.Contains(result, "Branches") {
					t.Error("Expected branch section to be excluded")
				}
				// Should not contain test.go (excluded by pattern)
				if strings.Contains(result, "test.go") {
					t.Error("Expected test.go to be excluded")
				}
				return nil
			},
		},
		{
			name: "invalid log file",
			args: []any{
				map[string]any{
					"log_file": "/nonexistent/file.csv",
				},
			},
			expectError: true,
		},
		{
			name: "invalid parameters",
			args: []any{
				map[string]any{
					"log_file": tmpFile.Name(),
					"top_n":    -1, // Invalid
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock LSP context (we don't use it in our implementation)
			ctx := (*glsp.Context)(nil)

			result, err := GenerateMetricsReport(ctx, tt.args)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if tt.validateFunc != nil {
				if err := tt.validateFunc(result); err != nil {
					t.Error(err)
				}
			}
		})
	}
}

func TestGenerateMetricsReport_WithConfig(t *testing.T) {
	// Test using config default log file
	tmpFile, err := os.CreateTemp("", "pendulum_config_test*.csv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write minimal test data
	testData := `active,branch,directory,file,filetype,project,time
true,main,/home,test.go,go,project,2024-01-01 10:00:00`

	if _, err := tmpFile.WriteString(testData); err != nil {
		t.Fatalf("Failed to write test data: %v", err)
	}
	tmpFile.Close()

	// Set up config with the test file
	err = config.Setup(
		config.WithActivityFile(tmpFile.Name()),
		config.WithLogFile("/tmp/lsp.log"),
		config.WithDebug(false),
	)
	if err != nil {
		t.Fatalf("Failed to setup config: %v", err)
	}

	// Test with empty log_file (should use config default)
	args := []any{
		map[string]any{
			"top_n": 5,
		},
	}

	ctx := (*glsp.Context)(nil)
	result, err := GenerateMetricsReport(ctx, args)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !strings.Contains(result, "# Pendulum Metrics Report") {
		t.Error("Expected valid report when using config default")
	}
}

func TestGenerateMetricsReport_EmptyFile(t *testing.T) {
	// Test with empty CSV file
	tmpFile, err := os.CreateTemp("", "pendulum_empty*.csv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write only header
	testData := `active,branch,directory,file,filetype,project,time`
	if _, err := tmpFile.WriteString(testData); err != nil {
		t.Fatalf("Failed to write test data: %v", err)
	}
	tmpFile.Close()

	args := []any{
		map[string]any{
			"log_file": tmpFile.Name(),
		},
	}

	ctx := (*glsp.Context)(nil)
	result, err := GenerateMetricsReport(ctx, args)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !strings.Contains(result, "No data available") {
		t.Error("Expected 'No data available' message for empty file")
	}
}

func TestGenerateHourlyReport(t *testing.T) {
	// Create a temporary CSV file with test data
	tmpFile, err := os.CreateTemp("", "pendulum_hours_test*.csv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write test data with different hours
	testData := `active,branch,directory,file,filetype,project,time
true,main,/home/user/project,main.go,go,myproject,2024-01-01 10:00:00
false,main,/home/user/project,main.go,go,myproject,2024-01-01 10:01:00
true,main,/home/user/project,test.go,go,myproject,2024-01-01 10:02:00
true,feature,/home/user/project,api.go,go,myproject,2024-01-01 14:00:00
false,feature,/home/user/docs,README.md,markdown,myproject,2024-01-01 14:01:00
true,main,/home/user/project,utils.go,go,myproject,2024-01-01 14:02:00`

	if _, err := tmpFile.WriteString(testData); err != nil {
		t.Fatalf("Failed to write test data: %v", err)
	}
	tmpFile.Close()

	tests := []struct {
		name         string
		args         []any
		expectError  bool
		validateFunc func(string) error
	}{
		{
			name: "basic hourly report",
			args: []any{
				map[string]any{
					"log_file": tmpFile.Name(),
					"top_n":    5,
				},
			},
			expectError: false,
			validateFunc: func(result string) error {
				if !strings.Contains(result, "# Pendulum Hours Report") {
					t.Error("Expected hours report header")
				}
				if !strings.Contains(result, "Times Most Active") {
					t.Error("Expected 'Times Most Active' section")
				}
				if !strings.Contains(result, "Processing Summary") {
					t.Error("Expected processing summary")
				}
				return nil
			},
		},
		{
			name: "invalid log file",
			args: []any{
				map[string]any{
					"log_file": "/nonexistent/file.csv",
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := (*glsp.Context)(nil)

			result, err := GenerateHourlyReport(ctx, tt.args)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if tt.validateFunc != nil {
				if err := tt.validateFunc(result); err != nil {
					t.Error(err)
				}
			}
		})
	}
}

func BenchmarkGenerateMetricsReport(b *testing.B) {
	// Create a larger test file for benchmarking
	tmpFile, err := os.CreateTemp("", "pendulum_bench*.csv")
	if err != nil {
		b.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Generate more test data
	var lines []string
	lines = append(lines, "active,branch,directory,file,filetype,project,time")

	for i := 0; i < 1000; i++ {
		active := "true"
		if i%2 == 0 {
			active = "false"
		}
		line := strings.Join([]string{
			active,
			"main",
			"/home/user/project",
			"file.go",
			"go",
			"myproject",
			"2024-01-01 10:00:00",
		}, ",")
		lines = append(lines, line)
	}

	if _, err := tmpFile.WriteString(strings.Join(lines, "\n")); err != nil {
		b.Fatalf("Failed to write test data: %v", err)
	}
	tmpFile.Close()

	args := []any{
		map[string]any{
			"log_file": tmpFile.Name(),
			"top_n":    10,
		},
	}

	ctx := (*glsp.Context)(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GenerateMetricsReport(ctx, args)
		if err != nil {
			b.Fatalf("Generate metrics failed: %v", err)
		}
	}
}
