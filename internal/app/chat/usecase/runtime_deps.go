package usecase

import (
	chatboundary "github.com/usetero/cli/internal/boundary/chat"
	"github.com/usetero/cli/internal/sqlite"
)

// RuntimeDeps contains orchestration dependencies for app/chat.
type RuntimeDeps struct {
	StreamRunner       StreamRunner
	StreamErrorMapper  StreamErrorMapper
	AssistantPersister AssistantPersister
	ToolLoop           ToolLoop
}

func NewRuntimeDeps(db sqlite.DB, client chatboundary.Client) RuntimeDeps {
	gateway := NewChatBoundaryGateway(client)
	return RuntimeDeps{
		StreamRunner:       NewChatStreamRunner(gateway),
		StreamErrorMapper:  NewChatBoundaryStreamErrorMapper(),
		AssistantPersister: NewSQLiteAssistantPersister(db),
		ToolLoop:           NewSQLiteToolLoop(db),
	}
}
