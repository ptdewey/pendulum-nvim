package pkg

import (
	"errors"
	"fmt"
)

// NOTE: add new fields here and to following function to extend functionalities
// - The fields in this struct should mirror the "command_args" table in `lua/pendulum/remote.lua`
type PendulumArgs struct {
	LogFile               string
	Timeout               float64
	TopN                  int
	TimeRange             string
	ReportExcludes        map[string]interface{}
	ReportSectionExcludes []interface{}
}

// Parse input arguments from lua table args
func ParsePendlumArgs(args map[string]interface{}) (*PendulumArgs, error) {
	logFile, ok := args["log_file"].(string)
	if !ok {
		return nil, errors.New("log_file missing or not a string. " +
			fmt.Sprintf("Type: %T\n", args["log_file"]))
	}

	topN, ok := args["top_n"].(int64)
	if !ok {
		return nil, errors.New("top_n missing or not a number. " +
			fmt.Sprintf("Type: %T\n", args["top_n"]))
	}

	timerLen, ok := args["timer_len"].(int64)
	if !ok {
		return nil, errors.New("timer_len missing or not a number. " +
			fmt.Sprintf("Type: %T\n", args["timer_len"]))
	}

	timeRange, ok := args["time_range"].(string)
	if !ok {
		return nil, errors.New("timeRange missing or not a string. " +
			fmt.Sprintf("Type: %T\n", args["time_range"]))
	}

	reportExcludes, ok := args["report_excludes"].(map[string]interface{})
	if !ok {
		return nil, errors.New("report_excludes missing or not an map. " +
			fmt.Sprintf("Type: %T\n", args["report_excludes"]))
	}

	reportSectionExcludes, ok := args["report_section_excludes"].([]interface{})
	if !ok {
		return nil, errors.New("report_excludes missing or not a list. " +
			fmt.Sprintf("Type: %T\n", args["report_section_excludes"]))
	}

	out := PendulumArgs{
		LogFile:               logFile,
		Timeout:               float64(timerLen),
		TopN:                  int(topN),
		TimeRange:             timeRange,
		ReportExcludes:        reportExcludes,
		ReportSectionExcludes: reportSectionExcludes,
	}

	return &out, nil
}
