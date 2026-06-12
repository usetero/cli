package usecase

import (
	"context"

	chatboundary "github.com/usetero/cli/internal/boundary/chat"
)

// RuntimeDeps contains orchestration dependencies for app/chat.
type RuntimeDeps struct {
	StreamRunner       StreamRunner
	StreamErrorMapper  StreamErrorMapper
	AssistantPersister AssistantPersister
	ToolLoop           ToolLoop
	OrphanCleaner      OrphanMessageCleaner
	EffectContext      context.Context
}

// NewRuntimeDeps wires the chat orchestration dependencies. Chat is ephemeral:
// messages live only in memory for the session, so the persistence collaborators
// are in-memory stand-ins rather than SQLite-backed stores.
func NewRuntimeDeps(client chatboundary.Client) RuntimeDeps {
	gateway := NewChatBoundaryGateway(client)
	return RuntimeDeps{
		StreamRunner:       NewChatStreamRunner(gateway),
		StreamErrorMapper:  NewChatBoundaryStreamErrorMapper(),
		AssistantPersister: NewMemoryAssistantPersister(),
		ToolLoop:           NewMemoryToolLoop(),
		OrphanCleaner:      NewMemoryOrphanMessageCleaner(),
		EffectContext:      context.Background(),
	}
}

// WithEffectContext returns a copy that uses ctx as the base for UI-triggered effects.
func (d RuntimeDeps) WithEffectContext(ctx context.Context) RuntimeDeps {
	if ctx == nil {
		ctx = context.Background()
	}
	d.EffectContext = ctx
	return d
}
