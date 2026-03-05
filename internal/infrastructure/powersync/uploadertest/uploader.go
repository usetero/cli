package uploadertest

import (
	"context"

	psuploader "github.com/usetero/cli/internal/infrastructure/powersync/uploader"
)

// Mock is a functional mock for uploader run/event behavior.
type Mock struct {
	RunFn    func(ctx context.Context) error
	EventsCh chan psuploader.Event
}

func NewMock() *Mock {
	return &Mock{EventsCh: make(chan psuploader.Event, 16)}
}

func (m *Mock) Run(ctx context.Context) error {
	if m.RunFn != nil {
		return m.RunFn(ctx)
	}
	<-ctx.Done()
	close(m.EventsCh)
	return ctx.Err()
}

func (m *Mock) Events() <-chan psuploader.Event {
	return m.EventsCh
}
