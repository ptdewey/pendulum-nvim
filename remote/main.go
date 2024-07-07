package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"pendulum-nvim/internal"
	"sync"

	"github.com/neovim/go-client/nvim"
)

// assumes args[0] is pendulum log file path
func RpcEventHandler(v *nvim.Nvim, args []string) error {
    if len(args) < 1 {
        return errors.New("Not enough arguments")
    }

    // TODO: get/create custom temporary buffer
    // - would require either passing in a buffer or passing the buffer back to nvim
    // - alternatively would require creating on vim side with name, then getting
    //   from v.Buffers() call by name
    // bufs, err := v.Buffers()
    // if err != nil {
    //     return err
    // }
    data, err := internal.ReadPendulumLogFile(args[0])
    if err != nil {
        return err
    }

    n := len(data[0])
    var wg sync.WaitGroup
    res := make(chan internal.PendulumMetric, n - 1)

    // iterate through each metric column (except for 'active') and create goroutine for each
    for i := 1; i < n; i++ {
        wg.Add(1)
        go func(i int) {
            defer wg.Done()
            internal.AggregatePendulumMetric(data, i, res)
        }(i)
    }

    // wait for all goroutines to finish
    wg.Wait()
    fmt.Println("Pre-close")
    close(res)
    fmt.Println("Finished closing")

    out := make([]internal.PendulumMetric, n)
    for r := range res {
        out = append(out, r)
        fmt.Println(r.Name, len(r.Value), r.ActivePct)
    }

    return errors.New("Finished?") // NOTE: for debugging

	// return v.WriteOut("Processing complete: " + out[3].Name) // For debugging
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
