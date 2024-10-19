package internal

import (
	"log"
	"strconv"
	"sync"
	"time"
)

type PendulumMetric struct {
	Name  string
	Index int
	Value map[string]*PendulumEntry
}

type PendulumEntry struct {
	ID               string
	ActiveCount      uint
	TotalCount       uint
	ActiveTime       time.Duration
	TotalTime        time.Duration
	ActiveTimestamps []string
	Timestamps       []string
	ActivePct        float32
}

var csvColumns = map[string]int{
	"active":      0,
	"branch":      1,
	"directories": 2,
	"file":        3,
	"filetype":    4,
	"project":     5,
	"time":        6,
}

// AggregatePendulumMetrics processes the input data to compute metrics for each column.
//
// Parameters:
// - data: A 2D slice of strings representing the pendulum data.
// - timeout_len: A float64 representing the timeout length.
// - rangeType: A string representing the time window to aggregate data for ("all" is recommended)
//
// Returns:
// - A slice of PendulumMetric structs containing the aggregated metrics.
func AggregatePendulumMetrics(
	data [][]string,
	timeout_len float64,
	rangeType string,
	reportExcludes map[string]interface{}) []PendulumMetric {
	// create waitgroup
	var wg sync.WaitGroup

	// create buffered channel to store results and avoid deadlock in main
	res := make(chan PendulumMetric, len(reportExcludes))

	// iterate through each metric column as specified in Sections config
	// and create goroutine for each
	for col := range len(csvColumns) {
		if col == csvColumns["active"] || col == csvColumns["time"] {
			continue
		}

		is_excluded := false
		for _, ex := range reportExcludes["groups"].([]interface{}) {
			if col == csvColumns[ex.(string)] {
				is_excluded = true
				break
			}
		}

		if is_excluded {
			continue
		}

		wg.Add(1)
		go func(m int) {
			defer wg.Done()
			aggregatePendulumMetric(
				data,
				m,
				timeout_len,
				rangeType,
				res,
			)
		}(col)
	}

	// handle waitgroup in separate goroutine to allow main routine
	// to process results as they become available.
	var cleanup_wg sync.WaitGroup
	cleanup_wg.Add(1)

	go func() {
		wg.Wait()
		close(res)
		cleanup_wg.Done()
	}()

	// deal with results
	out := make([]PendulumMetric, len(data[0]))
	for r := range res {
		out[r.Index] = r
	}

	// wait for cleanup goroutine to finish
	cleanup_wg.Wait()

	return out
}

// aggregatePendulumMetric aggregates metrics for a specific column of data.
//
// Parameters:
// - data: A 2D slice of strings representing the pendulum data.
// - m: An integer representing the column index to process.
// - timeout_len: A float64 representing the timeout length.
// - rangeType: A string representing the time window to aggregate data for ("all" is recommended)
// - ch: A channel to send the aggregated PendulumMetric.
//
// // Returns:
// // - None
func aggregatePendulumMetric(
	data [][]string,
	m int,
	timeout_len float64,
	rangeType string,
	ch chan<- PendulumMetric) {
	out := PendulumMetric{
		Name:  data[0][m],
		Index: m,
		Value: make(map[string]*PendulumEntry),
	}

	timecol := csvColumns["time"]

	// iterate through each row of data
	for i := 1; i < len(data[:]); i++ {
		active, err := strconv.ParseBool(data[i][0])
		if err != nil {
			log.Printf("Error parsing boolean at row %d, value: %s, error: %v", i, data[i][0], err)
		}

		// FIX: hour timeframe is broken (likely others as well -- probably caused by timezones)
		// TODO: add header to popup window showing the timeframe used (in buffer.go)
		// PERF: this check anecdotally makes the popup feel slower
		inRange, err := isTimestampInRange(data[i][timecol], rangeType)
		if err != nil {
			return
		}

		if !inRange {
			continue
		}

		// check if key doesn't exist in value map
		val := data[i][m]
		if out.Value[val] == nil {
			out.Value[val] = &PendulumEntry{
				ID:               val,
				ActiveCount:      0,
				TotalCount:       0,
				ActiveTime:       0,
				TotalTime:        0,
				Timestamps:       make([]string, 0),
				ActiveTimestamps: make([]string, 0),
			}
		}
		pv := out.Value[val]

		// metrics aggregation
		pv.TotalCount++
		pv.Timestamps = append(pv.Timestamps, data[i][timecol])
		t, err := timeDiff(pv.Timestamps, timeout_len)
		if err != nil {
			return
		}

		pv.TotalTime += t

		// active-only metrics aggregation
		if active == true {
			pv.ActiveCount++
			pv.ActiveTimestamps = append(pv.ActiveTimestamps, data[i][timecol])
			t, err := timeDiff(pv.ActiveTimestamps, timeout_len)
			if err != nil {
				return
			}

			pv.ActiveTime += t
		}
	}

	// calculate active percentage
	for _, v := range out.Value {
		v.ActivePct = float32(v.ActiveTime) / float32(v.TotalTime)
	}

	// pass output into channel
	ch <- out
}

// timeDiff calculates the time difference between the last two timestamps.
//
// Parameters:
// - timestamps: A slice of strings representing the timestamps.
// - timeout_len: A float64 representing the timeout length.
//
// Returns:
// - A time.Duration representing the time difference if it is within the timeout length.
// - An error if there is an issue parsing the timestamps.
func timeDiff(timestamps []string, timeout_len float64) (time.Duration, error) {
	n := len(timestamps)
	if n < 2 {
		return time.Duration(0), nil
	}

	curr, prev := timestamps[n-1], timestamps[n-2]
	d, err := calcDuration(curr, prev)
	if err != nil {
		return time.Duration(0), err
	}

	// if difference between timestamps exceeds timeout length then editor was closed between sessions.
	if d.Seconds() > timeout_len {
		return time.Duration(0), nil
	}

	return d, nil
}
