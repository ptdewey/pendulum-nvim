package internal

import (
	"fmt"
	"sort"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// TODO: docs
// TODO: options table (map of some sort?)
func PrettifyMetrics(metrics []PendulumMetric) []string {
    var lines []string

    // iterate over each metric
    for _, metric := range metrics {
        // TODO: redefine order? (might require hardcoding)
        lines = append(lines, prettifyMetric(metric))
    }

    return lines
}

// TODO: docs
func prettifyMetric(metric PendulumMetric) string {
    // set top n list length
    n := 5 // TODO: pass this in as an option

    keys := make([]string, 0, len(metric.Value))
    for k := range metric.Value {
        keys = append(keys, k)
        // TODO: get most active times of day using timestamp string arrays
    }

    // sort map by time spent active per key
    sort.SliceStable(keys, func(a int, b int) bool {
        return metric.Value[keys[a]].ActiveTime > metric.Value[keys[b]].ActiveTime
    })

    // find longest length ID value in top 5 to align text width
    l := 15
    for i := 0; i < n; i++ {
        il := len(truncatePath(metric.Value[keys[i]].ID))
        if l < il {
            l = il
        }
    }

    // write out top n list
    name := cases.Title(language.English, cases.Compact).String(metric.Name)
    out := fmt.Sprintf("# Top %d %s:\n", n, prettifyMetricName(name))
    for i := 0; i < n; i++ {
        out = fmt.Sprintln(out, prettifyEntry(metric.Value[keys[i]], i, l))
    }

    return out
}

// TODO: docs
func prettifyEntry(e *PendulumEntry, i int, l int) string {
    return fmt.Sprintf(" %d. %-*s: Total Time %+6s, Active Time %+6s (%-5.2f%%)",
        i + 1, l + 1, truncatePath(e.ID), formatDuration(e.TotalTime),
        formatDuration(e.ActiveTime), e.ActivePct * 100)
}

func prettifyMetricName(name string) string {
    switch name {
    case "Cwd":
        return "Directories"
    case "Branch":
        return "Branches"
    default:
        return fmt.Sprintf("%ss", name)
    }
}
