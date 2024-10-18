package pkg

import (
	"errors"
	"fmt"
)

// NOTE: add new fields here and to following function to extend functionalities
// - The fields in this struct should mirror the "command_args" table in `lua/pendulum/remote.lua`
type PendulumArgs struct {
	LogFile          string
	Timeout          float64
	TopN             int
	TimeRange        string
	Sections         []interface{}
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

	sections, ok := args["sections"].([]interface{})
	if !ok {
		return nil, errors.New("sections missing or not an array. " +
			fmt.Sprintf("Type: %T\n", args["sections"]))
	}

	out := PendulumArgs{
		LogFile:          logFile,
		Timeout:          float64(timerLen),
		TopN:             int(topN),
		TimeRange:        timeRange,
		Sections:         sections,
	}

	return &out, nil
}
