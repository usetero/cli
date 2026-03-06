package sessiontest

import (
	"context"

	psuploader "github.com/usetero/cli/internal/infrastructure/powersync/uploader"
)

type Uploader struct {
	EventsCh chan psuploader.Event
	RunErr   error
	RunFn    func(ctx context.Context) error
}

var _ interface {
	Run(ctx context.Context) error
	Events() <-chan psuploader.Event
} = (*Uploader)(nil)

func NewUploader() *Uploader {
	return &Uploader{EventsCh: make(chan psuploader.Event, 16)}
}

func (u *Uploader) Run(ctx context.Context) error {
	if u.RunFn != nil {
		return u.RunFn(ctx)
	}
	<-ctx.Done()
	close(u.EventsCh)
	if u.RunErr != nil {
		return u.RunErr
	}
	return ctx.Err()
}

func (u *Uploader) Events() <-chan psuploader.Event {
	return u.EventsCh
}
