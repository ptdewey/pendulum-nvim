package pkg

import (
	"pendulum-nvim/internal/data"
	"pendulum-nvim/internal/prettify"
	"pendulum-nvim/pkg/args"
	"strings"

	"github.com/neovim/go-client/nvim"
)

func CreateBuffer(v *nvim.Nvim) (nvim.Buffer, error) {
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
	data, err := data.ReadPendulumLogFile(args.PendulumArgs().LogFile)
	if err != nil {
		return buf, err
	}

	// get prettified buffer text
	bufText := getBufText(data)

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
//
// Returns:
// - A 2D slice of bytes representing the text to be set in the buffer.
func getBufText(pendulumData [][]string) [][]byte {
	pendulumArgs := args.PendulumArgs()

	out := data.AggregatePendulumMetrics(
		pendulumData[:],
		pendulumArgs.Timeout,
		pendulumArgs.TimeRange,
		pendulumArgs.ReportSectionExcludes,
		pendulumArgs.ReportExcludes,
	)

	var lines []string
	switch pendulumArgs.View {
	case "metrics":
		lines = prettify.PrettifyMetrics(out)
	case "hours":
		lines = prettify.PrettifyActiveHours(out)
	default:
		lines = prettify.PrettifyMetrics(out)
	}

	var bufText [][]byte
	for _, l := range lines {
		splitLines := strings.Split(l, "\n")
		for _, line := range splitLines {
			bufText = append(bufText, []byte(line))
		}
	}

	return bufText
}
