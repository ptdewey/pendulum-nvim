package data

import (
	"context"
	"os"
	"reflect"
	"testing"
)

func TestNewCSVReader(t *testing.T) {
	reader := NewCSVReader("test.csv")
	if reader.filepath != "test.csv" {
		t.Errorf("Expected filepath 'test.csv', got '%s'", reader.filepath)
	}
}

func TestCSVReader_ReadAll(t *testing.T) {
	// Create a temporary CSV file for testing
	tmpFile, err := os.CreateTemp("", "test*.csv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write test data
	testData := `active,branch,cwd,file,filetype,project,time
true,main,/home/user,test.go,go,myproject,2024-01-01 10:00:00
false,main,/home/user,test.go,go,myproject,2024-01-01 10:01:00
true,feature,/home/user,main.go,go,myproject,2024-01-01 10:02:00`

	if _, err := tmpFile.WriteString(testData); err != nil {
		t.Fatalf("Failed to write test data: %v", err)
	}
	tmpFile.Close()

	// Test ReadAll
	reader := NewCSVReader(tmpFile.Name())
	data, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}

	expectedRows := 4 // header + 3 data rows
	if len(data) != expectedRows {
		t.Errorf("Expected %d rows, got %d", expectedRows, len(data))
	}

	// Check header
	expectedHeader := []string{"active", "branch", "cwd", "file", "filetype", "project", "time"}
	if !reflect.DeepEqual(data[0], expectedHeader) {
		t.Errorf("Expected header %v, got %v", expectedHeader, data[0])
	}

	// Check first data row
	expectedFirstRow := []string{"true", "main", "/home/user", "test.go", "go", "myproject", "2024-01-01 10:00:00"}
	if !reflect.DeepEqual(data[1], expectedFirstRow) {
		t.Errorf("Expected first row %v, got %v", expectedFirstRow, data[1])
	}
}

func TestCSVReader_StreamRows(t *testing.T) {
	// Create a temporary CSV file
	tmpFile, err := os.CreateTemp("", "test*.csv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	testData := `active,branch,cwd
true,main,/home
false,dev,/tmp`

	if _, err := tmpFile.WriteString(testData); err != nil {
		t.Fatalf("Failed to write test data: %v", err)
	}
	tmpFile.Close()

	// Test streaming
	reader := NewCSVReader(tmpFile.Name())
	var rows [][]string

	ctx := context.Background()
	err = reader.StreamRows(ctx, func(row []string) error {
		rows = append(rows, row)
		return nil
	})

	if err != nil {
		t.Fatalf("StreamRows failed: %v", err)
	}

	if len(rows) != 3 { // header + 2 data rows
		t.Errorf("Expected 3 rows, got %d", len(rows))
	}
}

func TestCSVReader_FileNotFound(t *testing.T) {
	reader := NewCSVReader("nonexistent.csv")
	_, err := reader.ReadAll()

	if err == nil {
		t.Error("Expected error for nonexistent file")
	}

	if metricsErr, ok := err.(*MetricsError); ok {
		if metricsErr.Type != ErrFileNotFound {
			t.Errorf("Expected ErrFileNotFound, got %v", metricsErr.Type)
		}
	} else {
		t.Error("Expected MetricsError type")
	}
}

func TestCSVReader_ContextCancellation(t *testing.T) {
	// Create a temporary CSV file
	tmpFile, err := os.CreateTemp("", "test*.csv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write some test data
	testData := `active,branch
true,main
false,dev`

	if _, err := tmpFile.WriteString(testData); err != nil {
		t.Fatalf("Failed to write test data: %v", err)
	}
	tmpFile.Close()

	// Test with cancelled context
	reader := NewCSVReader(tmpFile.Name())
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err = reader.StreamRows(ctx, func(row []string) error {
		return nil
	})

	if err == nil {
		t.Error("Expected error for cancelled context")
	}

	if metricsErr, ok := err.(*MetricsError); ok {
		if metricsErr.Type != ErrProcessingTimeout {
			t.Errorf("Expected ErrProcessingTimeout, got %v", metricsErr.Type)
		}
	}
}
