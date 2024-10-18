package pkg

import (
	"pendulum-nvim/internal"
	"strings"

	"github.com/neovim/go-client/nvim"
)

// CreateBuffer creates a new Neovim buffer and populates it with data.
func CreateBuffer(v *nvim.Nvim, args PendulumArgs) (nvim.Buffer, error) {
	// create a new buffer
	buf, err := v.CreateBuffer(false, true)
	if err != nil {
		return buf, err
	}

	// set buffer filetype to add some highlighting
	if err := v.SetBufferOption(buf, "filetype", "markdown"); err != nil {
		return buf, err
	}

	// read pendulum data file
	data, err := internal.ReadPendulumLogFile(args.LogFile)
	if err != nil {
		return buf, err
	}

	// get prettified buffer text
	bufText := getBufText(data, args)

	// set contents of new buffer
	if err := v.SetBufferLines(buf, 0, -1, false, bufText); err != nil {
		return buf, err
	}

	// set buffer close keymap
	kopts := map[string]bool{
		"silent": true,
	}
	if err := v.SetBufferKeyMap(buf, "n", "q", "<cmd>close!<CR>", kopts); err != nil {
		return buf, err
	}

	return buf, nil
}

// getBufText processes the pendulum data and returns the text to be displayed in the buffer.
//
// Parameters:
// - data: A 2D slice of strings representing the pendulum data.
// - timeoutLen: A float64 representing the timeout length.
// - n: An integer representing the number of data points to aggregate.
// - rangeType: A string representing the time window to aggregate data for ("all" is recommended)
//
// Returns:
// - A 2D slice of bytes representing the text to be set in the buffer.
func getBufText(data [][]string, args PendulumArgs) [][]byte {
	out := internal.AggregatePendulumMetrics(data[:], args.Timeout, args.TimeRange, args.Sections)
	lines := internal.PrettifyMetrics(out, args.TopN)

	var bufText [][]byte
	for _, l := range lines {
		splitLines := strings.Split(l, "\n")
		for _, line := range splitLines {
			bufText = append(bufText, []byte(line))
		}
	}

	return bufText
}
