package data

import (
	"log"
	"pendulum-nvim/pkg/args"
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
	ActivePct        float64
}

var csvColumns = map[string]int{
	"active":    0,
	"branch":    1,
	"directory": 2,
	"file":      3,
	"filetype":  4,
	"project":   5,
	"time":      6,
}

func AggregatePendulumMetrics(data [][]string) []PendulumMetric {
	// create waitgroup
	var wg sync.WaitGroup

	// create buffered channel to store results and avoid deadlock in main
	res := make(chan PendulumMetric, len(data[0]))

	excludeMap := make(map[int]struct{})
	for _, section := range args.PendulumArgs().ReportSectionExcludes {
		if idx, exists := csvColumns[section.(string)]; exists {
			excludeMap[idx] = struct{}{}
		}
	}

	timeoutLen := args.PendulumArgs().Timeout
	reportExcludes := args.PendulumArgs().ReportExcludes

	// iterate through each metric column as specified in Sections config
	// and create goroutine for each
	for colName, colIdx := range csvColumns {
		if colName == "active" || colName == "time" {
			continue
		}

		if _, excluded := excludeMap[colIdx]; excluded {
			continue
		}

		wg.Add(1)
		go func(m int) {
			defer wg.Done()
			aggregatePendulumMetric(
				data,
				m,
				timeoutLen,
				reportExcludes,
				res,
			)
		}(colIdx)
	}

	go func() {
		wg.Wait()
		close(res)
	}()

	// deal with results
	out := make([]PendulumMetric, len(data[0]))
	for r := range res {
		out[r.Index] = r
	}

	return out
}

func aggregatePendulumMetric(
	data [][]string,
	colIdx int,
	timeoutLen float64,
	reportExcludes map[string]any,
	ch chan<- PendulumMetric,
) {
	out := PendulumMetric{
		Name:  data[0][colIdx],
		Index: colIdx,
		Value: make(map[string]*PendulumEntry),
	}

	timecol := csvColumns["time"]
	colName := out.Name
	if colName == "cwd" {
		// This is a bit hacky. csv uses cwd, code uses directory.
		// TODO: consolidate these two terms?
		colName = "directory"
	}

	exclusionPatterns, err := compileRegexPatterns(reportExcludes[colName])
	if err != nil {
		panic(err)
	}

	// iterate through each row of data
	for i := 1; i < len(data[:]); i++ {
		active, err := strconv.ParseBool(data[i][0])
		if err != nil {
			log.Printf("Error parsing boolean at row %d, value: %s, error: %v", i, data[i][0], err)
		}

		timeRange := args.PendulumArgs().TimeRange
		if timeRange != "all" {
			inRange, _ := isTimestampInRange(data[i][timecol], timeRange)
			if !inRange {
				continue
			}
		}

		val := data[i][colIdx]
		if isExcluded(val, exclusionPatterns) {
			continue
		}

		// check if key doesn't exist in value map
		if out.Value[val] == nil {
			out.Value[val] = &PendulumEntry{
				ID:               val,
				ActiveCount:      0,
				TotalCount:       0,
				ActiveTime:       0,
				TotalTime:        0,
				Timestamps:       make([]string, 0),
				ActiveTimestamps: make([]string, 0),
				ActivePct:        0,
			}
		}
		entry := out.Value[val]

		entry.updateTotalMetrics(data[i][timecol], timeoutLen)

		if active == true {
			entry.updateActiveMetrics(data[i][timecol], timeoutLen)
		}
	}

	calculateActivePercentages(out.Value)

	// pass output into channel
	ch <- out
}

func (entry *PendulumEntry) updateTotalMetrics(timestampStr string, timeoutLen float64) {
	entry.Timestamps = append(entry.Timestamps, timestampStr)
	tt, _ := timeDiff(entry.Timestamps, timeoutLen, false)
	entry.TotalCount++
	entry.TotalTime += tt
}

func (entry *PendulumEntry) updateActiveMetrics(timestampStr string, timeoutLen float64) {
	entry.ActiveTimestamps = append(entry.ActiveTimestamps, timestampStr)
	at, _ := timeDiff(entry.ActiveTimestamps, timeoutLen, false)
	entry.ActiveCount++
	entry.ActiveTime += at
}

func calculateActivePercentages(values map[string]*PendulumEntry) {
	for _, v := range values {
		if v.TotalTime > 0 {
			v.ActivePct = float64(v.ActiveTime) / float64(v.TotalTime)
		}
	}
}
