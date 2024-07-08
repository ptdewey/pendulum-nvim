package main

import (
	"errors"
	"log"
	"os"
	"pendulum-nvim/pkg"

	"github.com/neovim/go-client/nvim"
)

// assumes args[0] is pendulum log file path
func RpcEventHandler(v *nvim.Nvim, args []string) error {
    if len(args) < 1 {
        return errors.New("Not enough arguments")
    }

    // create and populate metrics report buffer
    buf, err := pkg.CreateBuffer(v, args[0])
    if err != nil {
        return err
    }

    // open popup window
    err = pkg.CreatePopupWindow(v, buf) // TODO: possibly end function by returning result of popup
    if err != nil {
        return err
    }

    // TODO: any necessary cleanup

    return nil
}

func main() {
    log.SetFlags(0)

    // redirect stdout to avoid garbling rpc stream
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
