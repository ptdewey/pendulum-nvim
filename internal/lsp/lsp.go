package lsp

import (
	"fmt"
	"log"

	"github.com/ptdewey/pendulum-server/internal/handlers"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

const (
	cmdLogActivity           string = "pendulum.logActivity"
	cmdActivityPing          string = "pendulum.activityPing"
	cmdStartSession          string = "pendulum.startSession"
	cmdEndSession            string = "pendulum.endSession"
	cmdGenerateMetricsReport string = "pendulum.generateMetricsReport"
	cmdGenerateHourlyReport  string = "pendulum.generateHourlyReport"
)

func Initialize(ctx *glsp.Context, params *protocol.InitializeParams) (any, error) {
	capabilities := protocol.ServerCapabilities{
		ExecuteCommandProvider: &protocol.ExecuteCommandOptions{
			Commands: []string{
				cmdLogActivity, cmdActivityPing, cmdStartSession, cmdEndSession,
				cmdGenerateMetricsReport, cmdGenerateHourlyReport,
			},
		},
	}

	return protocol.InitializeResult{
		Capabilities: capabilities,
		ServerInfo: &protocol.InitializeResultServerInfo{
			Name:    "pendulum-server",
			Version: &[]string{"2.0.0"}[0],
		},
	}, nil
}

func Initialized(ctx *glsp.Context, params *protocol.InitializedParams) error {
	log.Println("Server initialized successfully")

	ctx.Notify(protocol.ServerWindowLogMessage, &protocol.ShowMessageParams{
		Type:    protocol.MessageTypeInfo,
		Message: "pendulum-server is ready",
	})

	return nil
}

func Shutdown(ctx *glsp.Context) error {
	log.Println("pendulum-server shutting down")

	// Clean up activity manager
	if am := handlers.GetActivityManager(); am != nil {
		am.Stop()
	}

	return nil
}

func WorkspaceExecuteCommand(ctx *glsp.Context, params *protocol.ExecuteCommandParams) (any, error) {
	switch params.Command {
	case cmdLogActivity:
		return handlers.LogActivity(ctx, params.Arguments)
	case cmdActivityPing:
		return handlers.ActivityPing(ctx, params.Arguments)
	case cmdStartSession:
		return handlers.StartSession(ctx, params.Arguments)
	case cmdEndSession:
		return handlers.EndSession(ctx, params.Arguments)
	case cmdGenerateMetricsReport:
		return handlers.GenerateMetricsReport(ctx, params.Arguments)
	case cmdGenerateHourlyReport:
		return handlers.GenerateHourlyReport(ctx, params.Arguments)
	default:
		return nil, fmt.Errorf("unknown command: %s", params.Command)
	}
}

// These may of potential use (removing the need for some autocommands)
// TextDocumentDidOpen           TextDocumentDidOpenFunc
// TextDocumentDidChange         TextDocumentDidChangeFunc
// TextDocumentWillSave          TextDocumentWillSaveFunc
// TextDocumentWillSaveWaitUntil TextDocumentWillSaveWaitUntilFunc
// TextDocumentDidSave           TextDocumentDidSaveFunc
// TextDocumentDidClose          TextDocumentDidCloseFunc
