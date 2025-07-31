package params

import (
	"fmt"

	"github.com/ptdewey/pendulum-server/internal/handlers/data"
)

// ParseMetricsParams parses and validates LSP command arguments for metrics generation
func ParseMetricsParams(args []any) (*data.MetricsParams, error) {
	if len(args) == 0 {
		return nil, &data.MetricsError{
			Type:    data.ErrInvalidParameters,
			Message: "no arguments provided for metrics command",
		}
	}

	// First argument should be a map containing the options
	argMap, ok := args[0].(map[string]any)
	if !ok {
		return nil, &data.MetricsError{
			Type:    data.ErrInvalidParameters,
			Message: "expected first argument to be an options map",
		}
	}

	params := &data.MetricsParams{
		TopN:                  5, // defaults
		TimeRange:             "all",
		TimeFormat:            "12h",
		TimeZone:              "UTC",
		TimeoutLen:            180.0,
		ReportExcludes:        make(map[string][]string),
		ReportSectionExcludes: make([]string, 0),
	}

	// Parse log_file (optional - can use config default)
	if logFile, ok := argMap["log_file"].(string); ok {
		params.LogFile = logFile
	}

	// Parse top_n (optional)
	if topN, ok := argMap["top_n"]; ok {
		switch v := topN.(type) {
		case int:
			params.TopN = v
		case int64:
			params.TopN = int(v)
		case float64:
			params.TopN = int(v)
		default:
			return nil, &data.MetricsError{
				Type:    data.ErrInvalidParameters,
				Message: "top_n must be a number",
			}
		}
	}

	// Parse time_range (optional)
	if timeRange, ok := argMap["time_range"].(string); ok {
		params.TimeRange = timeRange
	}

	// Parse timeout_len (optional)
	if timeoutLen, ok := argMap["timeout_len"]; ok {
		switch v := timeoutLen.(type) {
		case int:
			params.TimeoutLen = float64(v)
		case int64:
			params.TimeoutLen = float64(v)
		case float64:
			params.TimeoutLen = v
		default:
			return nil, &data.MetricsError{
				Type:    data.ErrInvalidParameters,
				Message: "timeout_len must be a number",
			}
		}
	}

	// Parse time_format (optional)
	if timeFormat, ok := argMap["time_format"].(string); ok {
		params.TimeFormat = timeFormat
	}

	// Parse time_zone (optional)
	if timeZone, ok := argMap["time_zone"].(string); ok {
		params.TimeZone = timeZone
	}

	// Parse report_excludes (optional)
	if reportExcludes, ok := argMap["report_excludes"].(map[string]any); ok {
		for key, value := range reportExcludes {
			if excludeList, ok := value.([]any); ok {
				stringList := make([]string, 0, len(excludeList))
				for _, item := range excludeList {
					if str, ok := item.(string); ok {
						stringList = append(stringList, str)
					}
				}
				params.ReportExcludes[key] = stringList
			}
		}
	}

	// Parse report_section_excludes (optional)
	if sectionExcludes, ok := argMap["report_section_excludes"].([]any); ok {
		for _, item := range sectionExcludes {
			if str, ok := item.(string); ok {
				params.ReportSectionExcludes = append(params.ReportSectionExcludes, str)
			}
		}
	}

	// Validate parameters
	if err := validateParams(params); err != nil {
		return nil, err
	}

	return params, nil
}

// validateParams validates the parsed parameters
func validateParams(params *data.MetricsParams) error {
	// Don't validate LogFile here - it can be empty and filled from config later

	if params.TopN < 1 {
		return &data.MetricsError{
			Type:    data.ErrInvalidParameters,
			Message: "top_n must be greater than 0",
		}
	}

	if params.TimeoutLen < 0 {
		return &data.MetricsError{
			Type:    data.ErrInvalidParameters,
			Message: "timeout_len must be non-negative",
		}
	}

	validTimeRanges := map[string]bool{
		"all":   true,
		"today": true,
		"day":   true,
		"week":  true,
		"month": true,
		"year":  true,
		"hour":  true,
	}

	if !validTimeRanges[params.TimeRange] {
		return &data.MetricsError{
			Type:    data.ErrInvalidParameters,
			Message: fmt.Sprintf("invalid time_range: %s", params.TimeRange),
		}
	}

	return nil
}
