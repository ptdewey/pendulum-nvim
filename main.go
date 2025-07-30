package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/ptdewey/pendulum-server/internal/config"
	"github.com/ptdewey/pendulum-server/internal/lsp"
	"github.com/tliron/commonlog"
	protocol "github.com/tliron/glsp/protocol_3_16"
	"github.com/tliron/glsp/server"
)

var name = "pendulum-server"

func main() {
	// Parse command line flags
	var (
		csvPath         = flag.String("csv-path", "pendulum-log.csv", "Path to CSV activity log file")
		activityTimeout = flag.Int("activity-timeout", 5, "Activity timeout in seconds (how long to wait before marking as inactive)")
		checkInterval   = flag.Int("check-interval", 1, "Activity check interval in seconds (how often to check activity status)")
		debug           = flag.Bool("debug", false, "Enable debug mode with verbose logging")
		help            = flag.Bool("help", false, "Show this help information")
	)

	flag.Usage = func() {
		fmt.Printf("Usage: %s [options]\n\n", name)
		fmt.Println("Pendulum LSP server for activity tracking in Neovim")
		fmt.Println("\nOptions:")
		flag.PrintDefaults()
		fmt.Println("\nExamples:")
		fmt.Printf("  %s --csv-path activity.csv\n", name)
		fmt.Printf("  %s --csv-path logs/activity.csv --activity-timeout 10 --check-interval 2\n", name)
		fmt.Printf("  %s --debug --activity-timeout 3\n", name)
	}

	flag.Parse()

	if *help {
		flag.Usage()
		return
	}

	// Setup configuration with CLI flags
	if err := config.Setup(
		config.WithActivityFile(*csvPath),
		config.WithTimeoutLen(time.Duration(*activityTimeout)*time.Second),
		config.WithTimerLen(time.Duration(*checkInterval)*time.Second),
		config.WithDebug(*debug),
	); err != nil {
		log.Println(err)
		return
	}

	commonlog.Configure(1, &config.Config().LspLogFile)

	h := protocol.Handler{
		Initialize:              lsp.Initialize,
		Initialized:             lsp.Initialized,
		Shutdown:                lsp.Shutdown,
		WorkspaceExecuteCommand: lsp.WorkspaceExecuteCommand,
	}

	s := server.NewServer(&h, name, config.Config().Debug)

	if err := s.RunStdio(); err != nil {
		log.Println(err)
	}
}
