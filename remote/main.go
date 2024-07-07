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

    data, err := internal.ReadPendulumLogFile(args[0])
    if err != nil {
        return err
    }

    // get number of columns
    n := len(data[0])

    // create waitgroup
    var wg sync.WaitGroup

    // create buffered channel to avoid deadlock in main
    res := make(chan internal.PendulumMetric, n - 2)

    // iterate through each metric column (except for 'active' and 'time') and create goroutine for each
    for m := 1; m < n - 1; m++ {
        wg.Add(1)
        go func(m int) {
            defer wg.Done()
            internal.AggregatePendulumMetric(data, m, res)
            fmt.Println("finished goroutine:", m)
        }(m)
    }

    // handle waitgroup in separate goroutine to allow main routine to
    // process results as they become available.
    var cleanup_wg sync.WaitGroup
    cleanup_wg.Add(1)
    go func() {
        wg.Wait()
        close(res)
        cleanup_wg.Done()
    }()

    // deal with results
    out := make([]internal.PendulumMetric, 0)
    for r := range res {
        out = append(out, r)
        fmt.Println(r.Name, len(r.Value))
    }

    // wait for cleanup goroutine to finish
    cleanup_wg.Wait()

    // TODO: compile report from results, display in temporary buffer

    // TODO: get/create custom temporary buffer
    // - would require either passing in a buffer or passing the buffer back to nvim
    // - alternatively would require creating on vim side with name, then getting
    //   from v.Buffers() call by name
    // bufs, err := v.Buffers()
    // if err != nil {
    //     return err
    // }

	// return v.WriteOut("Processing complete: " + out[3].Name) // For debugging
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
