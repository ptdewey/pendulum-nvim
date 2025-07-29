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
	commonlog.Configure(1, &config.Config().LspLogFile)

	// TODO: flags for on startup to set config vars

	if err := config.Setup(
		config.WithActivityFile("pendulum-log.csv"),
		config.WithDebug(true),
	); err != nil {
		log.Println(err)
		return
	}

	h := protocol.Handler{
		Initialize:              lsp.Initialize,
		Initialized:             lsp.Initialized,
		WorkspaceExecuteCommand: lsp.WorkspaceExecuteCommand,
	}

	s := server.NewServer(&h, name, config.Config().Debug)

	if err := s.RunStdio(); err != nil {
		log.Println(err)
	}
}
