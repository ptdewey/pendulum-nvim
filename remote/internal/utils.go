package internal

import (
	"fmt"
	"time"
)

// calculate the duration between two string timestamps (curr - prev)
func calcDuration(curr string, prev string) (time.Duration, error) {
    layout := "2006-01-02 15:04:05"

    curr_t, err := time.Parse(layout, curr)
    if err != nil {
        return time.Duration(0), err
    }

    prev_t, err := time.Parse(layout, prev)
    if err != nil {
        return time.Duration(0), err
    }

    return curr_t.Sub(prev_t), nil
}

// reformat time.Duration values into a more readable format
func formatDuration(d time.Duration) string {
    if d >= 24 * time.Hour {
        days := d / (24 * time.Hour)
        return fmt.Sprintf("%dd", days)
    } else if d >= time.Hour {
        hours := d / time.Hour
        return fmt.Sprintf("%dh", hours)
    } else if d >= time.Minute {
        minutes := d / time.Minute
        return fmt.Sprintf("%dm", minutes)
    } else {
        seconds := d / time.Second
        return fmt.Sprintf("%ds", seconds)
    }
}
