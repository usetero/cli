package usecase

import (
	chatclient "github.com/usetero/cli/internal/api/chatclient"
	"github.com/usetero/cli/internal/sqlite"
)

// RuntimeDeps contains orchestration dependencies for app/chat.
type RuntimeDeps struct {
	StreamRunner       StreamRunner
	AssistantPersister AssistantPersister
	ToolLoop           ToolLoop
}

func NewRuntimeDeps(db sqlite.DB, client chatclient.Client) RuntimeDeps {
	gateway := NewChatClientGateway(client)
	return RuntimeDeps{
		StreamRunner:       NewChatStreamRunner(gateway),
		AssistantPersister: NewSQLiteAssistantPersister(db),
		ToolLoop:           NewSQLiteToolLoop(db),
	}
}
