package internal

import (
	"fmt"
	"sort"
	"time"
)

type hourFreq struct {
	hour  int
	count int
}

func PrettifyActiveHours(metrics []PendulumMetric, n int, timeFormat string) []string {
	for _, metric := range metrics {
		if metric.Name != "" && len(metric.Value) != 0 {
			return []string{prettifyActiveHours(metric, n, timeFormat)}
		}
	}

	return []string{}
}

func prettifyActiveHours(metric PendulumMetric, n int, timeFormat string) string {
	hourCounts := make(map[int]int)
	layout := "2006-01-02 15:04:05"

	// TODO: ensure there is no weird timezone stuff going on (1am maybe shouldn't be an active time for me)
	// - conversions may need to be done (I think pendulum uses UTC -- there was some sort of local TZ issue before)?
	for _, entry := range metric.Value {
		for _, ts := range entry.ActiveTimestamps {
			t, err := time.Parse(layout, ts)
			if err != nil {
				fmt.Println("Failed to parse timestamp: ", ts)
				continue
			}

			hourCounts[t.Hour()]++
		}
	}

	var sortedHours []hourFreq
	for hour, count := range hourCounts {
		sortedHours = append(sortedHours, hourFreq{hour: hour, count: count})
	}

	sort.SliceStable(sortedHours, func(a int, b int) bool {
		return sortedHours[a].count > sortedHours[b].count
	})

	if n > len(sortedHours) {
		n = len(sortedHours)
	}

	// TODO: convert occurrence count into percentages (count is difficult to interpret)

	out := fmt.Sprintf("# Top %d Most Active Hours:\n", n)
	for i := range n {
		h := sortedHours[i].hour
		c := sortedHours[i].count

		var period string
		if timeFormat == "12h" {
			h12 := h % 12
			if h12 == 0 {
				h12 = 12
			}
			period = "AM"
			if h >= 12 {
				period = "PM"
			}
			h = h12
		}

		// out += fmt.Sprintf(" %d. %02d:00%s-%02d:59%s : %d occurrences\n", i, h, period, h, period, c)
		out += fmt.Sprintf(" %d. %2d%s : %d occurrences\n", i+1, h, period, c)
	}

	return out
}
