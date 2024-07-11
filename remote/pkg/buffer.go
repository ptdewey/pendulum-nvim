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
    out := internal.AggregatePendelumMetrics(data)

    buf_text := make([][]byte, len(out))
    lines := internal.PrettifyMetrics(out)
    for i, l := range lines {
        fmt.Println(l)
        buf_text[i] = []byte(l)
    }

    return buf_text
}
