package pkg

import (
	"fmt"
	"pendulum-nvim/internal"

	"github.com/neovim/go-client/nvim"
)

// TODO: docs
func CreateBuffer(v *nvim.Nvim, filepath string) (nvim.Buffer, error) {
    // create a new nvim buffer
    buf, err := v.CreateBuffer(false, true)
    if err != nil {
        return buf, err
    }

    // read pendulum data file
    data, err := internal.ReadPendulumLogFile(filepath)
    if err != nil {
        return buf, err
    }

    buf_text := getBufText(data)

    // set contents of new buffer
    v.SetBufferLines(buf, 0, -1, true, buf_text)

    // set buffer close keymap
    opts := make(map[string]bool)
    opts["silent"] = true
    v.SetBufferKeyMap(buf, "n", "q", "<cmd>close!<CR>", opts)

    return buf, nil
}

// TODO: docs
func getBufText(data [][]string) ([][]byte) {
    // TODO: sort resulting order
    out := internal.AggregatePendelumMetrics(data)

    // convert PendulumMetric array output to string array (1D)
    buf_text := make([][]byte, len(out))
    for i, pm := range out {
        // TODO: create report output format (probably external function)

        // line := fmt.Sprintf("")
        // for k, e := range pm.Value {
        // }
        // buf_text[i] = []byte(fmt.Sprintf("%f ", pm.Value.ActiveTime.Minutes(), ))
        // fmt.Print(k, " ", e.ActiveTime.Minutes(), " " , e.TotalTime.Minutes(), " ", len(e.Timestamps), " ", len(e.ActiveTimestamps), " ", e.ActivePct, "\n")

        // TODO: make this pretty
        buf_text[i] = []byte(fmt.Sprint(pm.Name, " ", pm.Value))
    }

    return buf_text
}
