package main

import (
	"log"

	"github.com/ptdewey/pendulum-server/internal/config"
	"github.com/ptdewey/pendulum-server/internal/lsp"
	"github.com/tliron/commonlog"
	protocol "github.com/tliron/glsp/protocol_3_16"
	"github.com/tliron/glsp/server"
)

var name = "pendulum-server"

func main() {
	// TODO: flags for on startup to set config vars
	config.Setup(
		config.WithLogFile("pendulum-log.csv"),
		config.WithDebug(true),
	)

	commonlog.Configure(1, &config.Config().LogFile)

	h := protocol.Handler{
		Initialize:              lsp.Initialize,
		Initialized:             lsp.Initialized,
		WorkspaceExecuteCommand: lsp.WorkspaceExecuteCommand,
	}

	s := server.NewServer(&h, "pendulum-server", config.Config().Debug)

	if err := s.RunStdio(); err != nil {
		log.Println(err)
	}
}
