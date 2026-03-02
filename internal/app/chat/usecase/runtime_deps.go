package usecase

import (
	"context"

	chatboundary "github.com/usetero/cli/internal/boundary/chat"
	"github.com/usetero/cli/internal/sqlite"
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

func NewRuntimeDeps(db sqlite.DB, client chatboundary.Client) RuntimeDeps {
	gateway := NewChatBoundaryGateway(client)
	return RuntimeDeps{
		StreamRunner:       NewChatStreamRunner(gateway),
		StreamErrorMapper:  NewChatBoundaryStreamErrorMapper(),
		AssistantPersister: NewSQLiteAssistantPersister(db),
		ToolLoop:           NewSQLiteToolLoop(db),
		OrphanCleaner:      NewSQLiteOrphanMessageCleaner(db),
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
