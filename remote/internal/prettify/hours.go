package prettify

import (
	"fmt"
	"pendulum-nvim/internal/data"
	"pendulum-nvim/pkg/args"
	"sort"
	"time"
)

type hourFreq struct {
	hour  int
	count int
}

func PrettifyActiveHours(metrics []data.PendulumMetric) []string {
	pendulumArgs := args.PendulumArgs()
	for _, metric := range metrics {
		if metric.Name != "" && len(metric.Value) != 0 {
			return []string{prettifyActiveHours(metric, pendulumArgs.NHours,
				pendulumArgs.TimeFormat, pendulumArgs.TimeZone)}
		}
	}

	return []string{}
}

type hourDuration struct {
	hour     int
	duration time.Duration
}

func prettifyActiveHours(metric data.PendulumMetric, n int, timeFormat string, timeZone string) string {
	hourCounts := make(map[int]int)
	hourDurations := make(map[int]time.Duration)
	weekHourDurations := make(map[int]time.Duration)
	totalCount := 0

	loc, err := time.LoadLocation(timeZone)
	if err != nil {
		loc = time.UTC
	}

	layout := "2006-01-02 15:04:05"
	for _, entry := range metric.Value {
		for _, ts := range entry.ActiveTimestamps {
			var t time.Time

			t, err := time.Parse(layout, ts)
			if err != nil {
				fmt.Println("Failed to parse timestamp: ", ts)
				continue
			}

			hourCounts[t.In(loc).Hour()]++
			totalCount++
		}

		for k, v := range entry.ActiveTimeHours {
			t := time.Date(2006, 1, 2, k, 0, 0, 0, time.UTC)
			hourDurations[t.In(loc).Hour()] += v
		}

		for k, v := range entry.ActiveTimeHoursRecent {
			t := time.Date(2006, 1, 2, k, 0, 0, 0, time.UTC)
			weekHourDurations[t.In(loc).Hour()] += v
		}
	}

	// Create a slice of hourDuration structs to sort by duration
	var hourDurationSlice []hourDuration
	for hour, duration := range hourDurations {
		hourDurationSlice = append(hourDurationSlice, hourDuration{hour: hour, duration: duration})
	}

	// Sort by duration (largest first)
	sort.SliceStable(hourDurationSlice, func(a, b int) bool {
		return hourDurationSlice[a].duration > hourDurationSlice[b].duration
	})

	if n > len(hourDurationSlice) {
		n = len(hourDurationSlice)
	}

	// Column width stuff
	bulletWidth := len(fmt.Sprintf("%d", n))
	timeWidth := 8
	overallWidth := 13
	weeklyWidth := 13
	countWidth := 3

	// Header formatting
	out := fmt.Sprintf("# Times Most Active\n")
	out += fmt.Sprintf("%*s  %-*s %-*s %-*s %-*s\n",
		bulletWidth, "", timeWidth, "Time", overallWidth, "Overall",
		weeklyWidth, "This Week", countWidth, "Entry Count",
	)

	// Loop through results and format accordingly
	for i := range n {
		h24 := hourDurationSlice[i].hour
		c := hourCounts[h24]
		dur := hourDurations[h24]
		weeklyDur := weekHourDurations[h24]

		h := h24
		var period string
		if timeFormat == "12h" {
			h = h24 % 12
			if h == 0 {
				h = 12
			}
			period = "AM"
			if h24 >= 12 {
				period = "PM"
			}
		}

		out += fmt.Sprintf("%*d. %2d%s %-*s %-*v %-*v %-*d (%.2f%%)\n",
			bulletWidth, i+1,
			h, period,
			timeWidth-5, "",
			overallWidth, dur,
			weeklyWidth, weeklyDur,
			countWidth, c, float64(c)/float64(totalCount)*100,
		)
	}

	return out
}
