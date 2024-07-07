package internal

import (
	"log"
	"strconv"
)

type PendulumMetric struct {
    Name string
    Value map[string]*PendulumValue
    ActivePct float32
}

type PendulumValue struct {
    ID string
    Count uint
    Time float32
}

func AggregatePendulumMetric(data [][]string, m int, ch chan<-PendulumMetric) {
    out := PendulumMetric {
        Name: data[0][m],
        Value: make(map[string]*PendulumValue),
        ActivePct: 0,
    }

    // iterate through each row of data
    for i := 1; i < len(data[:]); i++ {
        val := data[i][m]
        active, err := strconv.ParseBool(data[i][0])
        if err != nil {
            log.Fatal(err)
        }

        if active == true {
            out.ActivePct++

            // check if key exists in value map
            if out.Value[val] == nil {
                out.Value[val] = &PendulumValue {
                    ID: val,
                    Count: 1,
                    Time: 0,
                }
            } else {
                out.Value[val].Count++
                out.Value[val].Time = calculateTime(val)
            }
        } else {
            // inactive case
        }
    }

    // average out number of active datapoints
    out.ActivePct /= float32(len(data[:][0]))

    // pass output into channel
    ch <- out
}

// TODO: implement this
func calculateTime(_ string) float32 {
    return 0
}
