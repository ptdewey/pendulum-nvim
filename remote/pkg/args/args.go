package args

import (
	"errors"
	"fmt"
)

// Add new fields here and to following function to extend functionalities
// - The fields in this struct should mirror the "command_args" table in `lua/pendulum/remote.lua`
type PendulumNvimArgs struct {
	LogFile               string
	Timeout               float64
	NMetrics              int
	NHours                int
	TimeRange             string
	ReportExcludes        map[string]any
	ReportSectionExcludes []any
	View                  string
	TimeFormat            string
	TimeZone              string
}

var pendulumArgs PendulumNvimArgs

func PendulumArgs() *PendulumNvimArgs {
	return &pendulumArgs
}

// Parse input arguments from Lua table args
func ParsePendlumArgs(args map[string]any) error {
	logFile, ok := args["log_file"].(string)
	if !ok {
		return errors.New("log_file missing or not a string. " +
			fmt.Sprintf("Type: %T\n", args["log_file"]))
	}

	topN, ok := args["top_n"].(int64)
	if !ok {
		return errors.New("top_n missing or not a number. " +
			fmt.Sprintf("Type: %T\n", args["top_n"]))
	}

	hoursN, ok := args["hours_n"].(int64)
	if !ok {
		return errors.New("hours_n missing or not a number. " +
			fmt.Sprintf("Type: %T\n", args["hours_n"]))
	}

	timerLen, ok := args["timer_len"].(int64)
	if !ok {
		return errors.New("timer_len missing or not a number. " +
			fmt.Sprintf("Type: %T\n", args["timer_len"]))
	}

	timeRange, ok := args["time_range"].(string)
	if !ok {
		return errors.New("time_range missing or not a string. " +
			fmt.Sprintf("Type: %T\n", args["time_range"]))
	}

	reportExcludes, ok := args["report_excludes"].(map[string]any)
	if !ok {
		return errors.New("report_excludes missing or not an map. " +
			fmt.Sprintf("Type: %T\n", args["report_excludes"]))
	}

	reportSectionExcludes, ok := args["report_section_excludes"].([]any)
	if !ok {
		return errors.New("report_excludes missing or not a list. " +
			fmt.Sprintf("Type: %T\n", args["report_section_excludes"]))
	}

	view, ok := args["view"].(string)
	if !ok {
		return errors.New("view missing or not a string. " + fmt.Sprintf("Type: %T\n", args["view"]))
	}

	timeFormat, ok := args["time_format"].(string)
	if !ok {
		return errors.New("time_format missing or not a string. " + fmt.Sprintf("Type: %T\n", args["time_format"]))
	}

	timeZone, ok := args["time_zone"].(string)
	if !ok {
		return errors.New("time_zone missing or not a string. " + fmt.Sprintf("Type: %T\n", args["time_zone"]))
	}

	pendulumArgs = PendulumNvimArgs{
		LogFile:               logFile,
		Timeout:               float64(timerLen),
		NMetrics:              int(topN),
		NHours:                int(hoursN),
		TimeRange:             timeRange,
		ReportExcludes:        reportExcludes,
		ReportSectionExcludes: reportSectionExcludes,
		View:                  view,
		TimeFormat:            timeFormat,
		TimeZone:              timeZone,
	}

	return nil
}
