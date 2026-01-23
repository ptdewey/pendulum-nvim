package prettify

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/ptdewey/pendulum-server/internal/handlers/data"
)

// MetricsFormatter handles formatting of metrics data for display
type MetricsFormatter struct {
	params *data.MetricsParams
}

// NewMetricsFormatter creates a new metrics formatter
func NewMetricsFormatter(params *data.MetricsParams) *MetricsFormatter {
	return &MetricsFormatter{params: params}
}

// FormatMetrics converts a slice of PendulumMetric structs into formatted strings
func (f *MetricsFormatter) FormatMetrics(metrics []data.PendulumMetric) []string {
	var lines []string

	// Add header with metadata
	header := f.generateHeader()
	if header != "" {
		lines = append(lines, header)
	}

	// Format each metric
	for i, metric := range metrics {
		if metric.Name != "" && len(metric.Value) != 0 {
			formatted := f.formatMetric(metric, f.params.TopN)
			lines = append(lines, formatted)

			// Add empty line between metrics (but not after the last one)
			if i < len(metrics)-1 {
				lines = append(lines, "")
			}
		}
	}

	return lines
}

// generateHeader creates a header with report metadata
func (f *MetricsFormatter) generateHeader() string {
	var parts []string

	parts = append(parts, "# Pendulum Metrics Report")
	parts = append(parts, fmt.Sprintf("**Generated:** %s", time.Now().Format("2006-01-02 15:04:05")))
	parts = append(parts, fmt.Sprintf("**Time Range:** %s", f.params.TimeRange))
	parts = append(parts, fmt.Sprintf("**Log File:** %s", truncateHome(f.params.LogFile)))
	parts = append(parts, "")

	return strings.Join(parts, "\n")
}

// formatMetric converts a single PendulumMetric struct into a formatted string
func (f *MetricsFormatter) formatMetric(metric data.PendulumMetric, n int) string {
	keys := make([]string, 0, len(metric.Value))
	for k := range metric.Value {
		keys = append(keys, k)
	}

	// Sort by active time (descending)
	sort.SliceStable(keys, func(a, b int) bool {
		return metric.Value[keys[a]].ActiveTime > metric.Value[keys[b]].ActiveTime
	})

	n = min(n, len(keys))

	// Find longest ID for alignment
	maxIDLen := 15
	for i := 0; i < n; i++ {
		idLen := len(f.truncatePath(metric.Value[keys[i]].ID))
		maxIDLen = max(maxIDLen, idLen)
	}

	// Generate formatted output
	name := f.titleCase(metric.Name)
	var out strings.Builder
	fmt.Fprintf(&out, "## Top %d %s\n", n, f.prettifyMetricName(name))

	for i := 0; i < n; i++ {
		entry := metric.Value[keys[i]]
		if math.IsNaN(entry.ActivePct) {
			continue
		}
		out.WriteString(f.formatEntry(entry, i+1, maxIDLen, n) + "\n")
	}

	return out.String()
}

// formatEntry converts a single PendulumEntry into a formatted string
func (f *MetricsFormatter) formatEntry(e *data.PendulumEntry, rank int, maxIDLen int, totalRanks int) string {
	rankWidth := len(fmt.Sprintf("%d", totalRanks))
	format := fmt.Sprintf("%%%dd. %%-%ds: Total %%6s, Active %%6s (%%-5.2f%%%%)",
		rankWidth, maxIDLen+1)

	return fmt.Sprintf(format,
		rank,
		f.truncatePath(e.ID),
		f.formatDuration(e.TotalTime),
		f.formatDuration(e.ActiveTime),
		e.ActivePct*100)
}

// prettifyMetricName converts metric names into a more readable form
func (f *MetricsFormatter) prettifyMetricName(name string) string {
	switch name {
	case "Cwd", "Directory":
		return "Directories"
	case "Branch":
		return "Branches"
	case "File":
		return "Files"
	case "Filetype":
		return "File Types"
	case "Project":
		return "Projects"
	default:
		return fmt.Sprintf("%ss", name)
	}
}

// truncatePath truncates long file paths for better display
func (f *MetricsFormatter) truncatePath(path string) string {
	maxLen := 50

	path = truncateHome(path)

	if len(path) <= maxLen {
		return path
	}

	// For file paths, show the end part
	if strings.Contains(path, string(filepath.Separator)) {
		parts := strings.FieldsFunc(path, func(c rune) bool {
			return c == filepath.Separator
		})

		if len(parts) > 1 {
			// Try to show last few parts
			result := parts[len(parts)-1]
			for i := len(parts) - 2; i >= 0 && len(result) < maxLen-3; i-- {
				candidate := parts[i] + string(filepath.Separator) + result
				if len(candidate) <= maxLen-3 {
					result = candidate
				} else {
					break
				}
			}
			if len(result) < len(path) {
				return "..." + result
			}
		}
	}

	// Fallback: truncate from the start
	return "..." + path[len(path)-maxLen+3:]
}

// formatDuration formats a time duration for display
func (f *MetricsFormatter) formatDuration(d time.Duration) string {
	if d == 0 {
		return "0s"
	}

	if d >= 24*time.Hour {
		days := float64(d) / float64(24*time.Hour)
		return fmt.Sprintf("%.2fd", days)
	} else if d >= time.Hour {
		hours := float64(d) / float64(time.Hour)
		return fmt.Sprintf("%.2fh", hours)
	} else if d >= time.Minute {
		minutes := float64(d) / float64(time.Minute)
		return fmt.Sprintf("%.2fm", minutes)
	} else {
		seconds := float64(d) / float64(time.Second)
		return fmt.Sprintf("%.2fs", seconds)
	}
}

// titleCase converts a string to title case
func (f *MetricsFormatter) titleCase(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
}

func truncateHome(path string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}

	if strings.HasPrefix(path, home) {
		rpath, err := filepath.Rel(home, path)
		if err == nil {
			path = "~" + string(filepath.Separator) + rpath
		}
	}

	return path
}

// hourDuration is a helper struct for sorting hours by duration
type hourDuration struct {
	hour     int
	duration time.Duration
}

// FormatHours converts PendulumHours into a formatted string report
func (f *MetricsFormatter) FormatHours(hours *data.PendulumHours, topN int) []string {
	var lines []string

	// Add header
	lines = append(lines, f.generateHoursHeader())

	// Format the hours report
	formatted := f.formatHoursReport(hours, topN)
	lines = append(lines, formatted)

	return lines
}

// generateHoursHeader creates a header for the hours report
func (f *MetricsFormatter) generateHoursHeader() string {
	var parts []string

	parts = append(parts, "# Pendulum Hours Report")
	parts = append(parts, fmt.Sprintf("**Generated:** %s", time.Now().Format("2006-01-02 15:04:05")))
	parts = append(parts, fmt.Sprintf("**Log File:** %s", truncateHome(f.params.LogFile)))
	parts = append(parts, "")

	return strings.Join(parts, "\n")
}

// formatHoursReport formats the hourly activity data
func (f *MetricsFormatter) formatHoursReport(hours *data.PendulumHours, n int) string {
	// Convert hour durations to local timezone
	loc, err := time.LoadLocation(f.params.TimeZone)
	if err != nil {
		loc = time.UTC
	}

	hourCountsActive := make(map[int]int)
	hourDurationsActive := make(map[int]time.Duration)
	hourDurationsTotal := make(map[int]time.Duration)
	weekHourDurationsActive := make(map[int]time.Duration)
	weekHourDurationsTotal := make(map[int]time.Duration)

	// Count active timestamps per hour
	layout := "2006-01-02 15:04:05"
	for _, ts := range hours.ActiveTimestamps {
		t, err := time.Parse(layout, ts)
		if err != nil {
			continue
		}
		hourCountsActive[t.In(loc).Hour()]++
	}

	// Convert hours to local timezone
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

	// Create and sort slice by active duration (with hour as secondary key for determinism)
	var hourDurationSlice []hourDuration
	for hour, duration := range hourDurationsActive {
		hourDurationSlice = append(hourDurationSlice, hourDuration{hour: hour, duration: duration})
	}

	sort.SliceStable(hourDurationSlice, func(a, b int) bool {
		if hourDurationSlice[a].duration != hourDurationSlice[b].duration {
			return hourDurationSlice[a].duration > hourDurationSlice[b].duration
		}
		// Secondary sort by hour for deterministic ordering when durations are equal
		return hourDurationSlice[a].hour < hourDurationSlice[b].hour
	})

	if n > len(hourDurationSlice) {
		n = len(hourDurationSlice)
	}

	if n == 0 {
		return "No hourly activity data available."
	}

	// Calculate column widths for alignment
	var overallHoursWidth int
	var recentHoursWidth int

	for _, d := range hourDurationsActive {
		w := len(f.formatDuration(d))
		if overallHoursWidth < w {
			overallHoursWidth = w
		}
	}

	for _, d := range weekHourDurationsActive {
		w := len(f.formatDuration(d))
		if recentHoursWidth < w {
			recentHoursWidth = w
		}
	}

	// Ensure minimum widths
	overallHoursWidth = max(overallHoursWidth, 6)
	recentHoursWidth = max(recentHoursWidth, 6)

	bulletWidth := len(fmt.Sprintf("%d", n))

	// Calculate column widths for proper alignment
	// Column format: "duration (pct%)" where pct is 5.2f = 6 chars + " (" + ")" = 9 extra
	overallColWidth := overallHoursWidth + 9
	recentColWidth := recentHoursWidth + 9

	// Ensure column widths are at least as wide as headers
	overallHeader := "Overall (Active %)"
	recentHeader := "This Week (Active %)"
	overallColWidth = max(overallColWidth, len(overallHeader))
	recentColWidth = max(recentColWidth, len(recentHeader))

	// Calculate max entry count width for right-alignment
	maxEntryCount := 0
	for i := 0; i < min(n, len(hourDurationSlice)); i++ {
		h24 := hourDurationSlice[i].hour
		if c := hourCountsActive[h24]; c > maxEntryCount {
			maxEntryCount = c
		}
	}
	entryCountWidth := max(len(fmt.Sprintf("%d", maxEntryCount)), len("Entry Count"))

	var out strings.Builder
	out.WriteString("## Times Most Active\n")
	out.WriteString(fmt.Sprintf("%*s  %-5s  %*s  %*s  %*s\n",
		bulletWidth, "",
		"Time",
		overallColWidth, overallHeader,
		recentColWidth, recentHeader,
		entryCountWidth, "Entry Count"))

	for i := 0; i < n; i++ {
		h24 := hourDurationSlice[i].hour
		c := hourCountsActive[h24]
		dur := hourDurationsActive[h24]
		weeklyDur := weekHourDurationsActive[h24]

		h := h24
		var period string
		if f.params.TimeFormat == "12h" {
			h = h24 % 12
			if h == 0 {
				h = 12
			}
			period = "AM"
			if h24 >= 12 {
				period = "PM"
			}
		}

		var overallPct, recentPct float64
		if total, exists := hourDurationsTotal[h24]; exists && total > 0 {
			overallPct = float64(dur) / float64(total) * 100
		}
		if total, exists := weekHourDurationsTotal[h24]; exists && total > 0 {
			recentPct = float64(weeklyDur) / float64(total) * 100
		}

		// Format the duration + percentage as a single column value (right-aligned)
		overallStr := fmt.Sprintf("%*s (%5.2f%%)", overallHoursWidth, f.formatDuration(dur), overallPct)
		recentStr := fmt.Sprintf("%*s (%5.2f%%)", recentHoursWidth, f.formatDuration(weeklyDur), recentPct)

		out.WriteString(fmt.Sprintf("%*d. %2d%-2s  %*s  %*s  %*d\n",
			bulletWidth, i+1,
			h, period,
			overallColWidth, overallStr,
			recentColWidth, recentStr,
			entryCountWidth, c,
		))
	}

	return out.String()
}
