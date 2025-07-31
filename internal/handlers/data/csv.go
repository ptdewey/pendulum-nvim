package data

import (
	"bufio"
	"context"
	"encoding/csv"
	"os"
)

// CSVReader provides enhanced CSV reading capabilities
type CSVReader struct {
	filepath string
}

// NewCSVReader creates a new CSV reader for the given file
func NewCSVReader(filepath string) *CSVReader {
	return &CSVReader{filepath: filepath}
}

// ReadAll reads the entire CSV file into memory
func (r *CSVReader) ReadAll() ([][]string, error) {
	f, err := os.Open(r.filepath)
	if err != nil {
		return nil, &MetricsError{
			Type:    ErrFileNotFound,
			Message: "failed to open pendulum log file",
			Cause:   err,
		}
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	data, err := csvReader.ReadAll()
	if err != nil {
		return nil, &MetricsError{
			Type:    ErrParsingFailed,
			Message: "failed to parse CSV data",
			Cause:   err,
		}
	}

	return data, nil
}

// StreamRows streams CSV rows one by one with context cancellation support
func (r *CSVReader) StreamRows(ctx context.Context, processor func([]string) error) error {
	f, err := os.Open(r.filepath)
	if err != nil {
		return &MetricsError{
			Type:    ErrFileNotFound,
			Message: "failed to open pendulum log file",
			Cause:   err,
		}
	}
	defer f.Close()

	csvReader := csv.NewReader(bufio.NewReader(f))

	// Read header
	header, err := csvReader.Read()
	if err != nil {
		return &MetricsError{
			Type:    ErrParsingFailed,
			Message: "failed to read CSV header",
			Cause:   err,
		}
	}

	// Process header
	if err := processor(header); err != nil {
		return err
	}

	rowNum := 1
	for {
		select {
		case <-ctx.Done():
			return &MetricsError{
				Type:    ErrProcessingTimeout,
				Message: "CSV processing cancelled",
				Cause:   ctx.Err(),
			}
		default:
		}

		record, err := csvReader.Read()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return &MetricsError{
				Type:    ErrParsingFailed,
				Message: "failed to parse CSV row",
				Cause:   err,
			}
		}

		if err := processor(record); err != nil {
			return &MetricsError{
				Type:    ErrParsingFailed,
				Message: "failed to process CSV row",
				Cause:   err,
			}
		}

		rowNum++
	}

	return nil
}

// BatchStream processes CSV rows in batches for better performance
func (r *CSVReader) BatchStream(ctx context.Context, batchSize int, processor func([][]string) error) error {
	var batch [][]string

	err := r.StreamRows(ctx, func(row []string) error {
		batch = append(batch, row)

		if len(batch) >= batchSize {
			if err := processor(batch); err != nil {
				return err
			}
			batch = batch[:0] // Reset batch
		}

		return nil
	})

	// Process remaining rows in batch
	if err == nil && len(batch) > 0 {
		err = processor(batch)
	}

	return err
}
