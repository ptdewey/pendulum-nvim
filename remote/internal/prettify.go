package internal

import (
	"fmt"
	"sort"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// TODO: docs
// TODO: options table (map of some sort?)
func PrettifyMetrics(metrics []PendulumMetric) []string {
    var lines []string

    // iterate over each metric
    for _, metric := range metrics {
        // TODO: sort by most common, show top n (defined in opts)
        lines = append(lines, prettifyMetric(metric))
    }

    return lines
}

// TODO: docs
func prettifyMetric(metric PendulumMetric) string {
    // set top n list length
    n := 5

    // TODO: get sorted top ~5 of each metric
    // - percentages would also be cool (maybe if they are greater than some %)
    // - include:
    //   - active percentage, active time, inactive time,
    var activeTime time.Duration = 0
    var totalTime time.Duration = 0
    var avgActivePct float32 = 0

    keys := make([]string, 0, len(metric.Value))
    for k, v := range metric.Value {
        keys = append(keys, k)
        // NOTE: sums could be moved to PendelumMetric type and calculated in goroutines
        activeTime += v.ActiveTime
        totalTime += v.TotalTime
        // TODO: get most active times of day using timestamp string arrays
    }
    avgActivePct /= float32(len(metric.Value))

    // sort map by time spent active per key
    vals := metric.Value
    sort.SliceStable(keys, func(a int, b int) bool {
        return vals[keys[a]].ActiveTime > vals[keys[b]].ActiveTime
    })

    // NOTE: might need to hardcode some of the Metric names, i.e. filetype,
    // in order to get desired result to show total time spent with languages,
    // projects and more
    name := cases.Title(language.English, cases.Compact).String(metric.Name)
    header := fmt.Sprintf("%s:", name)
    metric_stats := fmt.Sprintf("Total time spent in %s: %s, active time: %s (%f%%)",
        metric.Name, formatDuration(totalTime), formatDuration(activeTime),
        float32(activeTime) / float32(activeTime + totalTime) * 100,
    )

    // top n list
    list := make([]string, n)
    list = append(list, fmt.Sprintf("Top %d %ss:\n", n, name))
    for i := 0; i < n; i++ {
        list = append(list, prettifyEntry(metric.Value[keys[i]], i))
    }

    // TODO: finish final output, possibly change to Sprintln
    out := fmt.Sprintln(header, metric_stats, list)

    return out
}

// TODO: docs
func prettifyEntry(e *PendulumEntry, i int) string {
    return fmt.Sprintf(" %d. Total Time: %s, Active Time: %s (%f%%)\n",
        i + 1, formatDuration(e.TotalTime),
        formatDuration(e.ActiveTime), e.ActivePct * 100)
}
