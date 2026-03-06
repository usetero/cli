package uploader

import (
	"context"

	psdb "github.com/usetero/cli/internal/infrastructure/powersync/db"
)

type mutationDispatcher struct {
	handlers map[psdb.TableName]MutationHandler
}

func newMutationDispatcher() *mutationDispatcher {
	return &mutationDispatcher{
		handlers: make(map[psdb.TableName]MutationHandler),
	}
}

func (d *mutationDispatcher) SetHandler(table psdb.TableName, handler MutationHandler) {
	if handler == nil {
		delete(d.handlers, table)
		return
	}
	d.handlers[table] = handler
}

func (d *mutationDispatcher) Dispatch(ctx context.Context, mutation psdb.Mutation) error {
	handler, ok := d.handlers[mutation.Table]
	if !ok {
		return UnknownMutationHandlerError{Table: mutation.Table}
	}
	return handler.Handle(ctx, mutation)
}
