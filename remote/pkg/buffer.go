package pkg

import (
	// "fmt"
	"pendulum-nvim/internal"
	"strings"

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
    // FIX: not working
    err = v.SetBufferLines(buf, 0, -1, false, buf_text)
    if err != nil {
        return buf, err
    }


    // set buffer close keymap
    opts := make(map[string]bool)
    opts["silent"] = true
    v.SetBufferKeyMap(buf, "n", "q", "<cmd>close!<CR>", opts)

    return buf, nil
}

// TODO: docs
func getBufText(data [][]string) ([][]byte) {
    out := internal.AggregatePendelumMetrics(data)

    // TODO: get this working, it doesnt seem to be populating correctly
    lines := internal.PrettifyMetrics(out)
    var buf_text [][]byte
    for _, l := range lines {
        // fmt.Println(l)
        splitLines := strings.Split(l, "\n")
        for _, line := range splitLines {
            buf_text = append(buf_text, []byte(line))
        }
    }

    return buf_text
}
