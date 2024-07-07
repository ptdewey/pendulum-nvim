package internal

import (
	"log"
	"strconv"
	"time"
)

type PendulumMetric struct {
    Name string
    Value map[string]*PendulumEntry
}

type PendulumEntry struct {
    ID string
    ActiveCount uint
    TotalCount uint
    ActiveTime time.Duration
    TotalTime time.Duration
    ActiveTimestamps []string
    Timestamps []string
    ActivePct float32
}

// TODO: docs
func AggregatePendulumMetric(data [][]string, m int, ch chan<-PendulumMetric) {
    out := PendulumMetric {
        Name: data[0][m],
        Value: make(map[string]*PendulumEntry),
    }
    timecol := len(data[0]) - 1

    // iterate through each row of data
    for i := 1; i < len(data[:]); i++ {
        val := data[i][m]
        active, err := strconv.ParseBool(data[i][0])
        if err != nil {
            log.Fatal(err)
        }

        // check if key doesn't exist in value map
        if out.Value[val] == nil {
            out.Value[val] = &PendulumEntry {
                ID: val,
                ActiveCount: 0,
                TotalCount: 0,
                ActiveTime: 0,
                TotalTime: 0,
                Timestamps: make([]string, 0),
                ActiveTimestamps: make([]string, 0),
            }
        }
        pv := out.Value[val]

        // metrics aggregation
        pv.TotalCount++
        pv.Timestamps = append(pv.Timestamps, data[i][timecol])
        t, err := timeDiff(pv.Timestamps)
        if err != nil {
            return
        }
        pv.TotalTime += t

        // active-only metrics aggregation
        if active == true {
            pv.ActiveCount++
            pv.ActiveTimestamps = append(pv.ActiveTimestamps, data[i][timecol])
            t, err = timeDiff(pv.ActiveTimestamps)
            if err != nil {
                return
            }
            pv.ActiveTime += t
        }
    }

    // calculate active percentage
    // TODO: change to time rather than count? (or both?)
    for _, v := range out.Value {
        // v.ActivePct = float32(v.ActiveCount) / float32(v.TotalCount)
        v.ActivePct = float32(v.ActiveTime) / float32(v.TotalTime)
    }

    // pass output into channel
    ch <- out
}

// TODO: docs
// TEST: make sure this works as intended
func timeDiff(timestamps []string) (time.Duration, error) {
    n := len(timestamps)
    if n < 2 {
        return time.Duration(0), nil
    }

    curr, prev := timestamps[n - 1], timestamps[n - 2]
    d, err := calcDuration(curr, prev)
    if err != nil {
        return time.Duration(0), err
    }

    // if difference between timestamps exceeds timeout length then editor was closed between sessions.
    timeout_len := 120.0 // TODO: get this from init.lua
    if d.Seconds() > timeout_len {
        return time.Duration(0), nil
    }

    return d, nil
}
