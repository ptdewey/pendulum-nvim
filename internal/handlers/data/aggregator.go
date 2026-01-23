package data

import (
	"context"
	"log"
	"regexp"
	"runtime"
	"strconv"
	"sync"
	"time"
)

// MetricsAggregator handles concurrent aggregation of pendulum metrics
type MetricsAggregator struct {
	workerCount int
	params      *MetricsParams
}

// NewMetricsAggregator creates a new metrics aggregator
func NewMetricsAggregator(params *MetricsParams) *MetricsAggregator {
	workerCount := min(runtime.NumCPU(), 8) // Cap at 8 workers for memory efficiency

	return &MetricsAggregator{
		workerCount: workerCount,
		params:      params,
	}
}

// AggregatePendulumMetrics aggregates metrics using concurrent workers
func (a *MetricsAggregator) AggregatePendulumMetrics(ctx context.Context, data [][]string) (*ProcessingResult, error) {
	startTime := time.Now()

	if len(data) == 0 {
		return &ProcessingResult{
			Metrics:   []PendulumMetric{},
			Processed: 0,
			Duration:  time.Since(startTime),
		}, nil
	}

	// Create exclude maps
	excludeMap := make(map[int]struct{})
	for _, section := range a.params.ReportSectionExcludes {
		if idx, exists := CSVColumns[section]; exists {
			excludeMap[idx] = struct{}{}
		}
	}

	// Create work channels
	jobs := make(chan aggregationJob, len(CSVColumns))
	results := make(chan PendulumMetric, len(CSVColumns))
	errors := make(chan error, len(CSVColumns))

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < a.workerCount; i++ {
		wg.Add(1)
		go a.worker(ctx, &wg, jobs, results, errors)
	}

	// Send jobs
	go func() {
		defer close(jobs)
		for colName, colIdx := range CSVColumns {
			if colName == "active" || colName == "time" {
				continue
			}
			if _, excluded := excludeMap[colIdx]; excluded {
				continue
			}

			select {
			case jobs <- aggregationJob{
				data:     data,
				colIndex: colIdx,
				colName:  colName,
			}:
			case <-ctx.Done():
				return
			}
		}
	}()

	// Close results when all workers are done
	go func() {
		wg.Wait()
		close(results)
		close(errors)
	}()

	// Collect results
	metrics := make([]PendulumMetric, len(CSVColumns))
	var firstError error

	for {
		select {
		case result, ok := <-results:
			if !ok {
				// Channel closed, we're done
				goto done
			}
			metrics[result.Index] = result

		case err, ok := <-errors:
			if !ok {
				// Errors channel closed, ignore
				continue
			}
			if firstError == nil {
				firstError = err
			}
			log.Printf("Worker error: %v", err)

		case <-ctx.Done():
			return nil, &MetricsError{
				Type:    ErrProcessingTimeout,
				Message: "metrics aggregation cancelled",
				Cause:   ctx.Err(),
			}
		}
	}

done:
	if firstError != nil {
		return nil, firstError
	}

	// Filter out empty metrics
	var filteredMetrics []PendulumMetric
	for _, metric := range metrics {
		if metric.Name != "" && len(metric.Value) > 0 {
			filteredMetrics = append(filteredMetrics, metric)
		}
	}

	return &ProcessingResult{
		Metrics:   filteredMetrics,
		Processed: len(data) - 1, // Exclude header
		Duration:  time.Since(startTime),
	}, nil
}

type aggregationJob struct {
	data     [][]string
	colIndex int
	colName  string
}

// worker processes aggregation jobs
func (a *MetricsAggregator) worker(ctx context.Context, wg *sync.WaitGroup, jobs <-chan aggregationJob, results chan<- PendulumMetric, errors chan<- error) {
	defer wg.Done()

	for job := range jobs {
		select {
		case <-ctx.Done():
			return
		default:
		}

		metric, err := a.aggregateMetric(job.data, job.colIndex, job.colName)
		if err != nil {
			errors <- err
			continue
		}

		results <- metric
	}
}

// aggregateMetric aggregates a single metric column
func (a *MetricsAggregator) aggregateMetric(data [][]string, colIdx int, colName string) (PendulumMetric, error) {
	metric := PendulumMetric{
		Name:  data[0][colIdx],
		Index: colIdx,
		Value: make(map[string]*PendulumEntry),
	}

	timecol := CSVColumns["time"]

	// Handle cwd vs directory naming inconsistency
	filterColName := colName
	if colName == "cwd" {
		filterColName = "directory"
	}

	// Compile exclusion patterns
	var exclusionPatterns []*regexp.Regexp
	if filters, exists := a.params.ReportExcludes[filterColName]; exists {
		var err error
		exclusionPatterns, err = CompileRegexPatterns(filters)
		if err != nil {
			return metric, err
		}
	}

	// Create time range filter once (pre-computes boundaries)
	timeFilter, err := NewTimeRangeFilter(a.params.TimeRange, a.params.TimeZone)
	if err != nil {
		return metric, err
	}

	// Process each row
	for i := 1; i < len(data); i++ {
		if len(data[i]) <= colIdx || len(data[i]) <= timecol {
			continue // Skip malformed rows
		}

		active, err := strconv.ParseBool(data[i][0])
		if err != nil {
			log.Printf("Error parsing boolean at row %d, value: %s, error: %v", i, data[i][0], err)
			continue
		}

		// Check time range using pre-computed filter
		inRange, err := timeFilter.InRange(data[i][timecol])
		if err != nil {
			log.Printf("Error checking timestamp range: %v", err)
			continue
		}
		if !inRange {
			continue
		}

		val := data[i][colIdx]
		if IsExcluded(val, exclusionPatterns) {
			continue
		}

		// Initialize entry if doesn't exist
		if metric.Value[val] == nil {
			metric.Value[val] = &PendulumEntry{
				ID:               val,
				ActiveCount:      0,
				TotalCount:       0,
				ActiveTime:       0,
				TotalTime:        0,
				Timestamps:       make([]string, 0),
				ActiveTimestamps: make([]string, 0),
				ActivePct:        0,
			}
		}
		entry := metric.Value[val]

		// Update total metrics
		entry.updateTotalMetrics(data[i][timecol], a.params.TimeoutLen)

		// Update active metrics if active
		if active {
			entry.updateActiveMetrics(data[i][timecol], a.params.TimeoutLen)
		}
	}

	// Calculate active percentages
	a.calculateActivePercentages(metric.Value)

	return metric, nil
}

// updateTotalMetrics updates total count and time for an entry
func (entry *PendulumEntry) updateTotalMetrics(timestampStr string, timeoutLen float64) {
	entry.Timestamps = append(entry.Timestamps, timestampStr)
	tt, _ := TimeDiff(entry.Timestamps, timeoutLen, false)
	entry.TotalCount++
	entry.TotalTime += tt
}

// updateActiveMetrics updates active count and time for an entry
func (entry *PendulumEntry) updateActiveMetrics(timestampStr string, timeoutLen float64) {
	entry.ActiveTimestamps = append(entry.ActiveTimestamps, timestampStr)
	at, _ := TimeDiff(entry.ActiveTimestamps, timeoutLen, false)
	entry.ActiveCount++
	entry.ActiveTime += at
}

// calculateActivePercentages calculates the active percentage for all entries
func (a *MetricsAggregator) calculateActivePercentages(values map[string]*PendulumEntry) {
	for _, v := range values {
		if v.TotalTime > 0 {
			v.ActivePct = float64(v.ActiveTime) / float64(v.TotalTime)
		}
	}
}

// AggregatePendulumHours aggregates hourly activity data from CSV records
func (a *MetricsAggregator) AggregatePendulumHours(ctx context.Context, data [][]string) (*HoursResult, error) {
	startTime := time.Now()

	if len(data) <= 1 {
		return &HoursResult{
			Hours: &PendulumHours{
				ActiveTimestamps:      []string{},
				Timestamps:            []string{},
				ActiveTimeHours:       make(map[int]time.Duration),
				ActiveTimeHoursRecent: make(map[int]time.Duration),
				TotalTimeHours:        make(map[int]time.Duration),
				TotalTimeHoursRecent:  make(map[int]time.Duration),
			},
			Processed: 0,
			Duration:  time.Since(startTime),
		}, nil
	}

	hours := &PendulumHours{
		ActiveTimestamps:      []string{},
		Timestamps:            []string{},
		ActiveTimeHours:       make(map[int]time.Duration),
		ActiveTimeHoursRecent: make(map[int]time.Duration),
		TotalTimeHours:        make(map[int]time.Duration),
		TotalTimeHoursRecent:  make(map[int]time.Duration),
	}

	timecol := CSVColumns["time"]

	// Create time range filter for "recent" (last week)
	weekFilter, err := NewTimeRangeFilter("week", a.params.TimeZone)
	if err != nil {
		return nil, err
	}

	for i := 1; i < len(data); i++ {
		select {
		case <-ctx.Done():
			return nil, &MetricsError{
				Type:    ErrProcessingTimeout,
				Message: "hours aggregation cancelled",
				Cause:   ctx.Err(),
			}
		default:
		}

		if len(data[i]) <= timecol {
			continue
		}

		active, err := strconv.ParseBool(data[i][0])
		if err != nil {
			log.Printf("Error parsing boolean at row %d, value: %s, error: %v", i, data[i][0], err)
			continue
		}

		timestampStr := data[i][timecol]
		a.updateTotalHours(hours, timestampStr, weekFilter)

		if active {
			a.updateActiveHours(hours, timestampStr, weekFilter)
		}
	}

	return &HoursResult{
		Hours:     hours,
		Processed: len(data) - 1,
		Duration:  time.Since(startTime),
	}, nil
}

// updateTotalHours updates total time per hour
func (a *MetricsAggregator) updateTotalHours(hours *PendulumHours, timestampStr string, weekFilter *TimeRangeFilter) {
	hours.Timestamps = append(hours.Timestamps, timestampStr)

	t, err := time.Parse("2006-01-02 15:04:05", timestampStr)
	if err != nil {
		log.Printf("Error parsing timestamp: %s, error: %v", timestampStr, err)
		return
	}

	tth, _ := TimeDiff(hours.Timestamps, a.params.TimeoutLen, true)
	hours.TotalTimeHours[t.Hour()] += tth

	inRange, _ := weekFilter.InRange(timestampStr)
	if inRange {
		hours.TotalTimeHoursRecent[t.Hour()] += tth
	}
}

// updateActiveHours updates active time per hour
func (a *MetricsAggregator) updateActiveHours(hours *PendulumHours, timestampStr string, weekFilter *TimeRangeFilter) {
	hours.ActiveTimestamps = append(hours.ActiveTimestamps, timestampStr)

	t, err := time.Parse("2006-01-02 15:04:05", timestampStr)
	if err != nil {
		log.Printf("Error parsing timestamp: %s, error: %v", timestampStr, err)
		return
	}

	ath, _ := TimeDiff(hours.ActiveTimestamps, a.params.TimeoutLen, true)
	hours.ActiveTimeHours[t.Hour()] += ath

	inRange, _ := weekFilter.InRange(timestampStr)
	if inRange {
		hours.ActiveTimeHoursRecent[t.Hour()] += ath
	}
}
