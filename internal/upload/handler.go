package upload

import (
	"context"

	"github.com/usetero/cli/internal/powersync/db"
)

// Handler processes CRUD entries for a specific table.
type Handler interface {
	Handle(ctx context.Context, entry *db.CrudEntry, emit Emitter) error
}

// Emitter is a function that handlers use to emit events.
type Emitter func(Event)
