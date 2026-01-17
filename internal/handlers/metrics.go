package handlers

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/ptdewey/pendulum-server/internal/config"
	"github.com/ptdewey/pendulum-server/internal/handlers/data"
	"github.com/ptdewey/pendulum-server/internal/handlers/params"
	"github.com/ptdewey/pendulum-server/internal/handlers/prettify"
	"github.com/tliron/glsp"
)

// GenerateMetricsReport generates a formatted metrics report via LSP
func GenerateMetricsReport(ctx *glsp.Context, args []any) (string, error) {
	log.Printf("executing command 'pendulum.generateMetricsReport' with args %v", args)

	// Parse and validate parameters
	metricsParams, err := params.ParseMetricsParams(args)
	if err != nil {
		log.Printf("Error parsing metrics parameters: %v", err)
		return "", err
	}

	// Use config log file if not provided in params
	if metricsParams.LogFile == "" {
		cfg := config.Config()
		if cfg != nil && cfg.LogFile != "" {
			metricsParams.LogFile = cfg.LogFile
		} else {
			return "", &data.MetricsError{
				Type:    data.ErrFileNotFound,
				Message: "no log file specified and no default configured",
			}
		}
	}

	// Create processing context with timeout
	processingCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Read CSV data
	csvReader := data.NewCSVReader(metricsParams.LogFile)
	csvData, err := csvReader.ReadAll()
	if err != nil {
		log.Printf("Error reading CSV file: %v", err)
		return "", err
	}

	if len(csvData) <= 1 {
		return "# No data available\n\nThe log file contains no activity data.", nil
	}

	// Create aggregator and process metrics
	aggregator := data.NewMetricsAggregator(metricsParams)
	result, err := aggregator.AggregatePendulumMetrics(processingCtx, csvData)
	if err != nil {
		log.Printf("Error aggregating metrics: %v", err)
		return "", err
	}

	// Format results
	formatter := prettify.NewMetricsFormatter(metricsParams)
	formattedLines := formatter.FormatMetrics(result.Metrics)

	// Add processing metadata
	metadata := []string{
		"",
		"---",
		"",
		"**Processing Summary:**",
		"- Rows processed: " + formatNumber(result.Processed),
		"- Processing time: " + result.Duration.String(),
	}

	formattedLines = append(formattedLines, metadata...)

	// Join all lines into a single string
	output := strings.Join(formattedLines, "\n")

	log.Printf("Generated metrics report: %d lines, %d metrics, processed in %v",
		len(formattedLines), len(result.Metrics), result.Duration)

	return output, nil
}

// GenerateHourlyReport generates an hourly activity report via LSP
func GenerateHourlyReport(ctx *glsp.Context, args []any) (string, error) {
	log.Printf("executing command 'pendulum.generateHourlyReport' with args %v", args)

	// Parse and validate parameters (reuse metrics params)
	metricsParams, err := params.ParseMetricsParams(args)
	if err != nil {
		log.Printf("Error parsing metrics parameters: %v", err)
		return "", err
	}

	// Use config log file if not provided in params
	if metricsParams.LogFile == "" {
		cfg := config.Config()
		if cfg != nil && cfg.LogFile != "" {
			metricsParams.LogFile = cfg.LogFile
		} else {
			return "", &data.MetricsError{
				Type:    data.ErrFileNotFound,
				Message: "no log file specified and no default configured",
			}
		}
	}

	// Create processing context with timeout
	processingCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Read CSV data
	csvReader := data.NewCSVReader(metricsParams.LogFile)
	csvData, err := csvReader.ReadAll()
	if err != nil {
		log.Printf("Error reading CSV file: %v", err)
		return "", err
	}

	if len(csvData) <= 1 {
		return "# No data available\n\nThe log file contains no activity data.", nil
	}

	// Create aggregator and process hours
	aggregator := data.NewMetricsAggregator(metricsParams)
	result, err := aggregator.AggregatePendulumHours(processingCtx, csvData)
	if err != nil {
		log.Printf("Error aggregating hours: %v", err)
		return "", err
	}

	// Format results
	formatter := prettify.NewMetricsFormatter(metricsParams)
	formattedLines := formatter.FormatHours(result.Hours, metricsParams.TopN)

	// Add processing metadata
	metadata := []string{
		"",
		"---",
		"",
		"**Processing Summary:**",
		"- Rows processed: " + formatNumber(result.Processed),
		"- Processing time: " + result.Duration.String(),
	}

	formattedLines = append(formattedLines, metadata...)

	// Join all lines into a single string
	output := strings.Join(formattedLines, "\n")

	log.Printf("Generated hourly report: %d lines, processed in %v",
		len(formattedLines), result.Duration)

	return output, nil
}

// formatNumber formats a number for display with commas
func formatNumber(n int) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	if n < 1000000 {
		return fmt.Sprintf("%d,%03d", n/1000, n%1000)
	}
	return fmt.Sprintf("%.1fM", float64(n)/1000000)
}
