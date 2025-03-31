package prettify

import (
	"fmt"
	"math"
	"sort"

	"github.com/ptdewey/pendulum-nvim/internal/data"
	"github.com/ptdewey/pendulum-nvim/pkg/args"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// PrettifyMetrics converts a slice of PendulumMetric structs into a slice of formatted strings.
//
// Parameters:
// - metrics: A slice of PendulumMetric structs containing the metrics data.
//
// Returns:
// - A slice of strings where each string is a formatted representation of a metric.
func PrettifyMetrics(metrics []data.PendulumMetric) []string {
	var lines []string

	// TODO: add printing of plugin name, log file path, and report generation time
	// also time frame of the report
	// - Do this in a utility function to use with the active time display as well

	// iterate over each metric
	for _, metric := range metrics {
		// TODO: redefine order? (might require hardcoding)
		if metric.Name != "" && len(metric.Value) != 0 {
			lines = append(lines, prettifyMetric(metric, args.PendulumArgs().NMetrics))
		}
	}

	return lines
}

// Function prettifyMetric converts a single PendulumMetric struct into a formatted string.
//
// Parameters:
// - metric: A PendulumMetric struct containing the metric data.
// - n: An integer specifying the number of top entries to include in the output.
//
// Returns:
// - A string formatted to display the top n entries of the metric.
func prettifyMetric(metric data.PendulumMetric, n int) string {
	keys := make([]string, 0, len(metric.Value))
	for k := range metric.Value {
		keys = append(keys, k)
	}

	// Sort map by time spent active per key
	sort.SliceStable(keys, func(a int, b int) bool {
		return metric.Value[keys[a]].ActiveTime > metric.Value[keys[b]].ActiveTime
	})

	if n > len(keys) {
		n = len(keys)
	}

	// Find longest length ID value in top 5 to align text width
	l := 15
	for i := range n {
		il := len(truncatePath(metric.Value[keys[i]].ID))
		if l < il {
			l = il
		}
	}

	// write out top n list
	name := cases.Title(language.English, cases.Compact).String(metric.Name)
	out := fmt.Sprintf("# Top %d %s:\n", n, prettifyMetricName(name))
	for i := range n {
		if math.IsNaN(float64(metric.Value[keys[i]].ActivePct)) {
			continue
		}

		out = fmt.Sprintln(out, prettifyEntry(metric.Value[keys[i]], i, l, n))
	}

	return out
}

// prettifyEntry converts a single PendulumEntry into a formatted string.
//
// Parameters:
// - e: A pointer to a PendulumEntry struct containing the entry data.
// - i: An integer representing the index of the entry in the list.
// - l: An integer specifying the width for aligning the ID column.
// - n: An integer specifying the number of top entries to include in the output.
//
// Returns:
// - A formatted string representing the entry.
func prettifyEntry(e *data.PendulumEntry, i int, l int, n int) string {
	format := fmt.Sprintf("%%%dd. %%-%ds: Total Time %%+6s, Active Time %%+6s (%%-5.2f%%%%)",
		len(fmt.Sprintf("%d", n)), l+1)
	return fmt.Sprintf(format,
		i+1, truncatePath(e.ID), formatDuration(e.TotalTime),
		formatDuration(e.ActiveTime), e.ActivePct*100)
}

// prettifyMetricName converts metric names into a more readable form.
//
// Parameters:
// - name: A string containing the metric name.
//
// Returns:
// - A string representing the prettified metric name.
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
