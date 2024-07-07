package internal

import (
	"log"
	"strconv"
)

type PendulumMetric struct {
    Name string
    Value map[string]*PendulumValue
}

type PendulumValue struct {
    ID string
    Count uint
    Time float32
    ActivePct float32
}

// TODO: docs
func AggregatePendulumMetric(data [][]string, m int, ch chan<-PendulumMetric) {
    out := PendulumMetric {
        Name: data[0][m],
        Value: make(map[string]*PendulumValue),
    }

    // iterate through each row of data
    for i := 1; i < len(data[:]); i++ {
        val := data[i][m]
        active, err := strconv.ParseBool(data[i][0])
        if err != nil {
            log.Fatal(err)
        }

        if active == true {
            // check if key exists in value map
            if out.Value[val] == nil {
                out.Value[val] = &PendulumValue {
                    ID: val,
                    Count: 1,
                    Time: 0,
                    ActivePct: 1, // TODO: deal with activepct
                }
            } else {
                out.Value[val].Count++
                out.Value[val].Time = calculateTime(val)
            }
        } else {
            // inactive case
        }
    }

    // TODO: migrate this to a value-by-value basis
    // average out number of active datapoints
    // out.ActivePct /= float32(len(data[:]) - 1)

    // pass output into channel
    ch <- out
}

// TODO: implement
// - timestamps are taken upon opening a buffer, and leaving vim, and every {180} seconds
// - NOTE: might need to take in the number of seconds between polls
// - we know that if there are 2 of the same named timestamps more than {180} seconds apart,
//   then vim was not open and can ignore
// - if timestamps are within range, take time between name and next_name and add to time
// - NOTE: might need to add "vim closed" event to csv output to make this easier
func calculateTime(_ string) float32 {
    return 0
}
