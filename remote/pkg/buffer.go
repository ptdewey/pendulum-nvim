package pkg

import (
	"pendulum-nvim/internal"
	"strconv"
	"strings"

	"github.com/neovim/go-client/nvim"
)

// CreateBuffer creates a new Neovim buffer and populates it with data.
//
// Parameters:
// - v: A pointer to the Neovim instance.
// - args: A slice of strings where:
//   - args[0] is the path to the pendulum log file.
//   - args[1] is the timeout length as a string.
//   - args[2] is the number of data points to aggregate.
//
// Returns:
// - The created buffer.
// - An error if any step in the buffer creation or data population fails.
func CreateBuffer(v *nvim.Nvim, args []string) (nvim.Buffer, error) {
	// create a new buffer
	buf, err := v.CreateBuffer(false, true)
	if err != nil {
		return buf, err
	}

	// set buffer filetype to add some highlighting
	// TODO: a custom hl group would be considerably nicer looking
	if err := v.SetBufferOption(buf, "filetype", "markdown"); err != nil {
		return buf, err
	}

	// read pendulum data file
	data, err := internal.ReadPendulumLogFile(args[0])
	if err != nil {
		return buf, err
	}

	timeout_len, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		return buf, err
	}

	n, err := strconv.Atoi(args[2])
	if err != nil {
		return buf, err
	}

	// get prettified buffer text
	buf_text := getBufText(data, timeout_len, n)

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

// getBufText processes the pendulum data and returns the text to be displayed in the buffer.
//
// Parameters:
// - data: A 2D slice of strings representing the pendulum data.
// - timeout_len: A float64 representing the timeout length.
// - n: An integer representing the number of data points to aggregate.
//
// Returns:
// - A 2D slice of bytes representing the text to be set in the buffer.
func getBufText(data [][]string, timeout_len float64, n int) [][]byte {
	out := internal.AggregatePendulumMetrics(data, timeout_len)

	lines := internal.PrettifyMetrics(out, n)

	var buf_text [][]byte
	for _, l := range lines {
		splitLines := strings.Split(l, "\n")
		for _, line := range splitLines {
			buf_text = append(buf_text, []byte(line))
		}
	}

	return buf_text
}
