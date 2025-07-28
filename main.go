package main

import (
	"log"

	"github.com/ptdewey/pendulum-server/internal/lsp"
	"github.com/tliron/commonlog"
	protocol "github.com/tliron/glsp/protocol_3_16"
	"github.com/tliron/glsp/server"
)

var (
	name    = "pendulum-server"
	debug   = true
	logFile = "pendulum-log.csv"
)

func main() {
	commonlog.Configure(1, &logFile)

	h := protocol.Handler{
		Initialize:              lsp.Initialize,
		Initialized:             lsp.Initialized,
		WorkspaceExecuteCommand: lsp.WorkspaceExecuteCommand,
	}

	s := server.NewServer(&h, "pendulum-server", debug)

	if err := s.RunStdio(); err != nil {
		log.Println(err)
	}
}
