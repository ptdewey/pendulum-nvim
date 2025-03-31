package data

import (
	"log"
	"strconv"
	"time"

	"github.com/ptdewey/pendulum-nvim/pkg/args"
)

type PendulumHours struct {
	ActiveTimestamps      []string
	Timestamps            []string
	ActiveTimeHours       map[int]time.Duration
	ActiveTimeHoursRecent map[int]time.Duration
	TotalTimeHours        map[int]time.Duration
	TotalTimeHoursRecent  map[int]time.Duration
}

func AggregatePendlulumHours(data [][]string) *PendulumHours {
	out := &PendulumHours{
		ActiveTimestamps:      []string{},
		Timestamps:            []string{},
		ActiveTimeHours:       map[int]time.Duration{},
		ActiveTimeHoursRecent: map[int]time.Duration{},
		TotalTimeHours:        map[int]time.Duration{},
		TotalTimeHoursRecent:  map[int]time.Duration{},
	}

	timeoutLen := args.PendulumArgs().Timeout

	// TODO: Exclude handling for hours tab should likely be different from metrics
	// due to this hours running only once (as opposed to once per column).
	// Presence of any excluded term in a row would exclude the time from the entire row here.
	//
	// var exclusionPatterns []*regexp.Regexp
	// for _, colName := range data[0] {
	// 	if colName == "branches" || colName == "projects" {
	// 		continue
	// 	}
	//
	// 	if patterns, exists := args.PendulumArgs().ReportExcludes[colName]; exists {
	// 		compiledPatterns, err := compileRegexPatterns(patterns)
	// 		if err != nil {
	// 			panic(err)
	// 		}
	// 		exclusionPatterns = append(exclusionPatterns, compiledPatterns...)
	// 	}
	// }

	for i := 1; i < len(data[:]); i++ {
		// excluded := false
		// for _, val := range data[i] {
		// 	if isExcluded(val, exclusionPatterns) {
		// 		excluded = true
		// 		break
		// 	}
		// }
		// if excluded {
		// 	continue
		// }

		active, err := strconv.ParseBool(data[i][0])
		if err != nil {
			log.Printf("Error parsing boolean at row %d, value: %s, error: %v", i, data[i][0], err)
		}

		timestampStr := data[i][csvColumns["time"]]
		out.updateTotalHours(timestampStr, timeoutLen)

		if active == true {
			out.updateActiveHours(timestampStr, timeoutLen)
		}
	}

	return out
}

// Recent (last week) total hours
func (p *PendulumHours) updateTotalHours(timestampStr string, timeoutLen float64) {
	p.Timestamps = append(p.Timestamps, timestampStr)
	t, _ := time.Parse("2006-01-02 15:04:05", timestampStr)

	tth, _ := timeDiff(p.Timestamps, timeoutLen, true)
	p.TotalTimeHours[t.Hour()] += tth

	inRange, _ := isTimestampInRange(timestampStr, "week")
	if inRange {
		p.TotalTimeHoursRecent[t.Hour()] += tth
	}
}

// Extract active time per hour
func (entry *PendulumHours) updateActiveHours(timestampStr string, timeoutLen float64) {
	entry.ActiveTimestamps = append(entry.ActiveTimestamps, timestampStr)
	t, _ := time.Parse("2006-01-02 15:04:05", timestampStr)

	ath, _ := timeDiff(entry.ActiveTimestamps, timeoutLen, true)
	entry.ActiveTimeHours[t.Hour()] += ath

	if inRange, _ := isTimestampInRange(timestampStr, "week"); inRange {
		entry.ActiveTimeHoursRecent[t.Hour()] += ath
	}
}
