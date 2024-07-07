package internal

import "time"

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
