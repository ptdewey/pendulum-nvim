package main

import (
	"errors"
	"log"
	"os"
	"pendulum-nvim/pkg"
	"pendulum-nvim/pkg/args"

	"github.com/neovim/go-client/nvim"
)

// RpcEventHandler handles the RPC call from Lua and creates a buffer with pendulum data.
func RpcEventHandler(v *nvim.Nvim, luaArgs map[string]any) error {
	// Extract and validate arguments from input table
	err := args.ParsePendlumArgs(luaArgs)
	if err != nil {
		return err
	}

	// Call CreateBuffer with the struct
	buf, err := pkg.CreateBuffer(v)
	if err != nil {
		return err
	}

	// Open popup window
	// if err := pkg.CreateNewTab(v, buf); err != nil {
	if err := pkg.CreatePopupWindow(v, buf); err != nil {
		return err
	}

	return nil
}

func main() {
	log.SetFlags(0)

	// Redirect stdout to stderr
	stdout := os.Stdout
	os.Stdout = os.Stderr

	// Connect to Neovim
	v, err := nvim.New(os.Stdin, stdout, stdout, log.Printf)
	if err != nil {
		log.Fatal(err)
	}

	// Register the "pendulum" RPC handler, which receives Lua tables
	v.RegisterHandler("pendulum", func(v *nvim.Nvim, args ...any) error {
		// Expecting the first argument to be a map (Lua table)
		if len(args) < 1 {
			return errors.New("not enough arguments")
		}

		// Parse the first argument as a map
		argMap, ok := args[0].(map[string]any)
		if !ok {
			return errors.New("expected a map as the first argument")
		}

		// Call the actual handler with the parsed map
		return RpcEventHandler(v, argMap)
	})

	// Run the RPC message loop
	if err := v.Serve(); err != nil {
		log.Fatal(err)
	}
}
