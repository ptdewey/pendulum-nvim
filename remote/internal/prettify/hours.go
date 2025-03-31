package prettify

import (
	"fmt"
	"sort"
	"time"
	"unicode/utf8"

	"github.com/ptdewey/pendulum-nvim/internal/data"
	"github.com/ptdewey/pendulum-nvim/pkg/args"
)

type hourDuration struct {
	hour     int
	duration time.Duration
}

func PrettifyActiveHours(hours *data.PendulumHours) []string {
	pendulumArgs := args.PendulumArgs()
	return []string{
		prettifyActiveHours(hours, pendulumArgs.NHours,
			pendulumArgs.TimeFormat, pendulumArgs.TimeZone),
	}
}

func prettifyActiveHours(hours *data.PendulumHours, n int, timeFormat string, timeZone string) string {
	hourCountsActive := make(map[int]int)

	hourDurationsActive := make(map[int]time.Duration)
	hourDurationsTotal := make(map[int]time.Duration)
	weekHourDurationsActive := make(map[int]time.Duration)
	weekHourDurationsTotal := make(map[int]time.Duration)

	loc, err := time.LoadLocation(timeZone)
	if err != nil {
		loc = time.UTC
	}

	layout := "2006-01-02 15:04:05"
	for _, ts := range hours.ActiveTimestamps {
		var t time.Time

		t, err := time.Parse(layout, ts)
		if err != nil {
			fmt.Println("Failed to parse timestamp: ", ts)
			continue
		}

		hourCountsActive[t.In(loc).Hour()]++
	}

	for k, v := range hours.ActiveTimeHours {
		t := time.Date(2006, 1, 2, k, 0, 0, 0, time.UTC)
		hourDurationsActive[t.In(loc).Hour()] += v

	}

	for k, v := range hours.TotalTimeHours {
		t := time.Date(2006, 1, 2, k, 0, 0, 0, time.UTC)
		hourDurationsTotal[t.In(loc).Hour()] += v
	}

	for k, v := range hours.ActiveTimeHoursRecent {
		t := time.Date(2006, 1, 2, k, 0, 0, 0, time.UTC)
		weekHourDurationsActive[t.In(loc).Hour()] += v
	}

	for k, v := range hours.TotalTimeHoursRecent {
		t := time.Date(2006, 1, 2, k, 0, 0, 0, time.UTC)
		weekHourDurationsTotal[t.In(loc).Hour()] += v
	}

	// Create a slice of hourDuration structs to sort by duration
	var hourDurationSlice []hourDuration
	for hour, duration := range hourDurationsActive {
		hourDurationSlice = append(hourDurationSlice, hourDuration{hour: hour, duration: duration})
	}

	// Sort by duration (largest first)
	sort.SliceStable(hourDurationSlice, func(a, b int) bool {
		return hourDurationSlice[a].duration > hourDurationSlice[b].duration
	})

	if n > len(hourDurationSlice) {
		n = len(hourDurationSlice)
	}

	var overallHoursWidth int
	var recentHoursWidth int

	for _, d := range hourDurationsActive {
		w := utf8.RuneCountInString(fmt.Sprintf("%v", d))
		if overallHoursWidth < w {
			overallHoursWidth = w
		}
	}

	for _, d := range weekHourDurationsActive {
		w := utf8.RuneCountInString(fmt.Sprintf("%v", d))
		if recentHoursWidth < w {
			recentHoursWidth = w
		}
	}

	// Column width stuff
	bulletWidth := len(fmt.Sprintf("%d", n))
	// TODO: store largest duration string lengths

	// Header formatting
	out := fmt.Sprintf("# Times Most Active\n")
	out += fmt.Sprintf("%*s  %-*s %-*s %-*s %-*s\n",
		bulletWidth, "", 7, "Time", 21, "Overall (Active %)",
		23, "This Week (Active %)", 3, "Entry Count",
	)

	// Loop through results and format accordingly
	for i := range n {
		h24 := hourDurationSlice[i].hour
		c := hourCountsActive[h24]
		dur := hourDurationsActive[h24]
		weeklyDur := weekHourDurationsActive[h24]

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

		out += fmt.Sprintf("%*d. %2d%s %-*s%-*v (%.2f%%)%-*s %-*v (%.2f%%)%-*s %-*d\n",
			bulletWidth, i+1,
			h, period,
			3, "",
			overallHoursWidth, dur,
			float64(dur)/float64(hourDurationsTotal[h24])*100,
			3, "",
			recentHoursWidth, weeklyDur,
			float64(weeklyDur)/float64(weekHourDurationsTotal[h24])*100,
			6, "",
			3, c,
		)
	}

	return out
}
