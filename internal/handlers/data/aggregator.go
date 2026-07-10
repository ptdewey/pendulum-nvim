package data

import (
	"context"
	"regexp"
	"strconv"
	"time"
)

// MetricsAggregator derives reports from one chronological interval stream.
type MetricsAggregator struct {
	workerCount int
	params      *MetricsParams
}
type sample struct {
	active bool
	values []string
	at     time.Time
	raw    string
	// barrier prevents a valid sample after invalid input from being joined to
	// the prior valid sample. Unknown activity must never be invented.
	barrier bool
}
type interval struct {
	sample     sample
	start, end time.Time
}

func NewMetricsAggregator(params *MetricsParams) *MetricsAggregator {
	return &MetricsAggregator{workerCount: 1, params: params}
}

func (a *MetricsAggregator) AggregatePendulumMetrics(ctx context.Context, rows [][]string) (*ProcessingResult, error) {
	started := time.Now()
	samples, rejected := parseSamples(rows)
	if len(samples) == 0 {
		return &ProcessingResult{Metrics: []PendulumMetric{}, Rejected: rejected, Duration: time.Since(started)}, nil
	}
	if err := ctx.Err(); err != nil {
		return nil, cancelledError(err)
	}
	filter, err := NewTimeRangeFilter(a.params.TimeRange, a.params.TimeZone)
	if err != nil {
		return nil, err
	}
	// Build once. Every dimension derives from this same immutable stream.
	intervals := buildIntervals(samples, a.params.TimeoutLen)
	excludedColumns := map[int]bool{}
	for _, name := range a.params.ReportSectionExcludes {
		if i, ok := CSVColumns[name]; ok {
			excludedColumns[i] = true
		}
	}
	metrics := make([]PendulumMetric, 0, 5)
	for _, i := range []int{1, 2, 3, 4, 5} {
		if excludedColumns[i] {
			continue
		}
		if err := ctx.Err(); err != nil {
			return nil, cancelledError(err)
		}
		name := columnName(i)
		patterns, err := a.exclusionPatterns(name)
		if err != nil {
			return nil, err
		}
		metric := PendulumMetric{Name: name, Index: i, Value: map[string]*PendulumEntry{}}
		for _, in := range intervals {
			clipped, ok := filter.Clip(in.start, in.end)
			if !ok {
				continue
			}
			value := in.sample.values[i]
			if IsExcluded(value, patterns) {
				continue
			}
			entry := metric.Value[value]
			if entry == nil {
				entry = &PendulumEntry{ID: value}
				metric.Value[value] = entry
			}
			d := clipped.end.Sub(clipped.start)
			entry.TotalCount++
			entry.TotalTime += d
			if in.sample.active {
				entry.ActiveCount++
				entry.ActiveTime += d
			}
		}
		for _, entry := range metric.Value {
			if entry.TotalTime > 0 {
				entry.ActivePct = float64(entry.ActiveTime) / float64(entry.TotalTime)
			}
		}
		metrics = append(metrics, metric)
	}
	return &ProcessingResult{Metrics: metrics, Processed: len(samples), Rejected: rejected, Duration: time.Since(started)}, nil
}

func (a *MetricsAggregator) AggregatePendulumHours(ctx context.Context, rows [][]string) (*HoursResult, error) {
	started := time.Now()
	samples, rejected := parseSamples(rows)
	if err := ctx.Err(); err != nil {
		return nil, cancelledError(err)
	}
	loc, err := time.LoadLocation(a.params.TimeZone)
	if err != nil {
		return nil, &MetricsError{Type: ErrInvalidParameters, Message: "load timezone", Cause: err}
	}
	week, err := NewTimeRangeFilter("week", a.params.TimeZone)
	if err != nil {
		return nil, err
	}
	hours := newPendulumHours()
	for _, s := range samples {
		hours.Timestamps = append(hours.Timestamps, s.raw)
		if s.active {
			hours.ActiveTimestamps = append(hours.ActiveTimestamps, s.raw)
		}
	}
	for _, in := range buildIntervals(samples, a.params.TimeoutLen) {
		if err := ctx.Err(); err != nil {
			return nil, cancelledError(err)
		}
		allocateHours(hours.TotalTimeHours, in.start, in.end, loc)
		if in.sample.active {
			allocateHours(hours.ActiveTimeHours, in.start, in.end, loc)
		}
		if clipped, ok := week.Clip(in.start, in.end); ok {
			allocateHours(hours.TotalTimeHoursRecent, clipped.start, clipped.end, loc)
			if in.sample.active {
				allocateHours(hours.ActiveTimeHoursRecent, clipped.start, clipped.end, loc)
			}
		}
	}
	return &HoursResult{Hours: hours, Processed: len(samples), Rejected: rejected, Duration: time.Since(started)}, nil
}

func parseSamples(rows [][]string) ([]sample, int) {
	if len(rows) <= 1 {
		return nil, 0
	}
	samples := make([]sample, 0, len(rows)-1)
	rejected := 0
	barrier := false
	for _, row := range rows[1:] {
		if len(row) <= CSVColumns["time"] {
			rejected++
			barrier = true
			continue
		}
		active, err := strconv.ParseBool(row[CSVColumns["active"]])
		if err != nil {
			rejected++
			barrier = true
			continue
		}
		at, err := ParseTimestamp(row[CSVColumns["time"]])
		if err != nil {
			rejected++
			barrier = true
			continue
		}
		if len(samples) > 0 && at.Before(samples[len(samples)-1].at) {
			// Do not reorder logs: that hides clock/data errors and would change
			// interval ownership. It is an explicit discontinuity instead.
			rejected++
			barrier = true
			continue
		}
		samples = append(samples, sample{active: active, values: append([]string(nil), row...), at: at, raw: row[CSVColumns["time"]], barrier: barrier})
		barrier = false
	}
	return samples, rejected
}
func buildIntervals(samples []sample, timeoutSeconds float64) []interval {
	out := make([]interval, 0, max(0, len(samples)-1))
	for i := 1; i < len(samples); i++ {
		start, end := samples[i-1].at, samples[i].at
		if !samples[i].barrier && end.Sub(start).Seconds() <= timeoutSeconds {
			out = append(out, interval{sample: samples[i-1], start: start, end: end})
		}
	}
	return out
}
func (a *MetricsAggregator) exclusionPatterns(name string) ([]*regexp.Regexp, error) {
	return CompileRegexPatterns(a.params.ReportExcludes[name])
}
func columnName(index int) string {
	for name, i := range CSVColumns {
		if i == index {
			return name
		}
	}
	return ""
}
func newPendulumHours() *PendulumHours {
	return &PendulumHours{ActiveTimestamps: []string{}, Timestamps: []string{}, ActiveTimeHours: map[int]time.Duration{}, ActiveTimeHoursRecent: map[int]time.Duration{}, TotalTimeHours: map[int]time.Duration{}, TotalTimeHoursRecent: map[int]time.Duration{}}
}
func allocateHours(target map[int]time.Duration, start, end time.Time, loc *time.Location) {
	for start.Before(end) {
		local := start.In(loc)
		next := time.Date(local.Year(), local.Month(), local.Day(), local.Hour()+1, 0, 0, 0, loc).UTC()
		if !next.After(start) {
			next = start.Truncate(time.Hour).Add(time.Hour)
		}
		if next.After(end) {
			next = end
		}
		target[local.Hour()] += next.Sub(start)
		start = next
	}
}
func cancelledError(err error) error {
	return &MetricsError{Type: ErrProcessingTimeout, Message: "metrics aggregation cancelled", Cause: err}
}
