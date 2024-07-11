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
    // TODO: get sorted top ~5 of each metric
    // - percentages would also be cool (maybe if they are greater than some %)
    // - include:
    //   - active percentage, active time, inactive time,
    keys := make([]string, len(metric.Value))
    var activeTime time.Duration = 0
    var totalTime time.Duration = 0
    var avgActivePct float32 = 0

    for k, v := range metric.Value {
        keys = append(keys, k)
        activeTime += v.ActiveTime
        totalTime += v.TotalTime
        avgActivePct += v.ActivePct
        // TODO: get most active times of day using timestamp string arrays
    }
    avgActivePct /= float32(len(metric.Value))

    // sort map by time spent active per key
    sort.SliceStable(keys, func(a int, b int) bool {
        return metric.Value[keys[a]].ActiveTime > metric.Value[keys[b]].ActiveTime
    })

    n := 5

    // NOTE: might need to hardcode some of the Metric names, i.e. filetype,
    // in order to get desired result to show total time spent with languages,
    // projects and more
    header := fmt.Sprintf("Top %d %ss:\n", n, cases.Title(language.English, cases.Compact).String(metric.Name))

    // NOTE: might need to change percentage string to inclue 'Average' to avoid misinforming users
    // TODO: include units
    metric_stats := fmt.Sprintf("Total time in %s: %f | Active time: %f | Active percentage: %f",
        metric.Name, totalTime.Hours(), activeTime.Hours(), avgActivePct,
    )

    // top n list
    list := make([]string, n)
    for i := 0; i < n; i++ {
        list = append(list, prettifyEntry(metric.Value[keys[i]], i))
    }

    // TODO: finish final output, possibly change to Sprintln
    out := fmt.Sprint(header, metric_stats, list)

    return out
}


// TODO: docs
func prettifyEntry(e *PendulumEntry, i int) string {
    return fmt.Sprintf(" %d. Total Time: %f, Active Time: %f, Active Percentage: %f", i + 1, e.TotalTime.Hours(), e.ActiveTime.Hours, e.ActivePct)
}
