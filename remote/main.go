package main

import (
	"errors"
	"log"
	"os"
	"pendulum-nvim/pkg"

	"github.com/neovim/go-client/nvim"
)

// RpcEventHandler handles RPC events and creates a metrics report popup.
//
// Parameters:
//   - v: A pointer to the Neovim instance.
//   - args: A slice of strings where args[0] is expected to be the pendulum
//     log file path, args[1] is the timeout length (active -> inactive),
//     and args[2] is the number of top entries to show for each metric.
//
// Returns:
//   - An error if there are not enough arguments, if the buffer cannot be created,
//     or if the popup window cannot be created.
func RpcEventHandler(v *nvim.Nvim, args []string) error {
	if len(args) < 3 {
		return errors.New("Not enough arguments")
	}

	// create and populate metrics report buffer
	buf, err := pkg.CreateBuffer(v, args)
	if err != nil {
		return err
	}

	// open popup window
	if err := pkg.CreatePopupWindow(v, buf); err != nil {
		return err
	}

	return err
}

func main() {
	log.SetFlags(0)

	// only use stdout for stderr
	stdout := os.Stdout
	os.Stdout = os.Stderr

	v, err := nvim.New(os.Stdin, stdout, stdout, log.Printf)
	if err != nil {
		log.Fatal(err)
	}

	v.RegisterHandler("pendulum", RpcEventHandler)

	// run rpc message loop
	if err := v.Serve(); err != nil {
		log.Fatal(err)
		return
	}
}
