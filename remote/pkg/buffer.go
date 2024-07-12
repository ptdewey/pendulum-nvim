package pkg

import (
	"pendulum-nvim/internal"
	"strings"

	"github.com/neovim/go-client/nvim"
)

// TODO: docs
func CreateBuffer(v *nvim.Nvim, filepath string) (nvim.Buffer, error) {
    // create a new buffer
    buf, err := v.CreateBuffer(false, true)
    if err != nil {
        return buf, err
    }

    // set buffer filetype to add some highlighting
    // TODO: a custom hl group would be considerably better
    if err := v.SetBufferOption(buf, "filetype", "markdown"); err != nil {
        return buf, err
    }

    // read pendulum data file
    data, err := internal.ReadPendulumLogFile(filepath)
    if err != nil {
        return buf, err
    }

    // get prettified buffer text
    buf_text := getBufText(data)

    // set contents of new buffer
    if err := v.SetBufferLines(buf, 0, -1, false, buf_text); err != nil {
        return buf, err
    }

    // set buffer close keymap
    kopts := make(map[string]bool)
    kopts["silent"] = true
    v.SetBufferKeyMap(buf, "n", "q", "<cmd>close!<CR>", kopts)

    return buf, nil
}

// TODO: docs
func getBufText(data [][]string) ([][]byte) {
    out := internal.AggregatePendulumMetrics(data)

    lines := internal.PrettifyMetrics(out)

    var buf_text [][]byte
    for _, l := range lines {
        splitLines := strings.Split(l, "\n")
        for _, line := range splitLines {
            buf_text = append(buf_text, []byte(line))
        }
    }

    return buf_text
}
