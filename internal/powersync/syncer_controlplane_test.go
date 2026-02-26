package powersync

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/powersync/extension"
)

func TestSyncer_ControlPlaneAccessIsSerialized(t *testing.T) {
	t.Parallel()

	cp := &blockingControlPlane{}
	s := &syncer{
		control: cp,
		scope:   logtest.NewScope(t),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			_ = s.processLine(ctx, []byte(`{"token_expires_in":3600}`))
		}()
		go func() {
			defer wg.Done()
			_ = s.NotifyUploadCompleted(ctx)
		}()
	}
	wg.Wait()

	if got := cp.maxInFlight.Load(); got > 1 {
		t.Fatalf("control plane operations overlapped, max_in_flight=%d", got)
	}
}

func TestSyncer_NotifyUploadCompleted_ContextCancellationPropagates(t *testing.T) {
	t.Parallel()

	cp := &contextAwareControlPlane{}
	s := &syncer{
		control: cp,
		scope:   logtest.NewScope(t),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	err := s.NotifyUploadCompleted(ctx)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("NotifyUploadCompleted() error = %v, want deadline exceeded", err)
	}
}

func TestSyncer_ProcessLine_ContextCancellationPropagates(t *testing.T) {
	t.Parallel()

	cp := &contextAwareControlPlane{}
	s := &syncer{
		control: cp,
		scope:   logtest.NewScope(t),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	err := s.processLine(ctx, []byte(`{"token_expires_in":3600}`))
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("processLine() error = %v, want deadline exceeded", err)
	}
}

func TestSyncer_NotifyUploadCompleted_ConcurrentWithStop_IsSafe(t *testing.T) {
	t.Parallel()

	cp := &blockingControlPlane{}
	s := &syncer{
		control: cp,
		scope:   logtest.NewScope(t),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = s.NotifyUploadCompleted(ctx)
		}()
	}

	// Stop concurrently while notifications are in flight.
	s.Stop()
	wg.Wait()
}

func TestSyncer_ProcessLine_ControlPlaneUnavailable(t *testing.T) {
	t.Parallel()

	s := &syncer{
		scope: logtest.NewScope(t),
	}

	err := s.processLine(context.Background(), []byte(`{"token_expires_in":3600}`))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, errControlPlaneUnavailable) {
		t.Fatalf("processLine() error = %v, want errControlPlaneUnavailable", err)
	}
}

type blockingControlPlane struct {
	inFlight    atomic.Int32
	maxInFlight atomic.Int32
}

func (b *blockingControlPlane) recordCall() {
	current := b.inFlight.Add(1)
	for {
		max := b.maxInFlight.Load()
		if current <= max {
			break
		}
		if b.maxInFlight.CompareAndSwap(max, current) {
			break
		}
	}
	time.Sleep(2 * time.Millisecond)
	b.inFlight.Add(-1)
}

func (b *blockingControlPlane) Start(context.Context, extension.StartRequest) ([]extension.Instruction, error) {
	b.recordCall()
	return nil, nil
}

func (b *blockingControlPlane) SendTextLine(context.Context, string) ([]extension.Instruction, error) {
	b.recordCall()
	return nil, nil
}

func (b *blockingControlPlane) NotifyConnection(context.Context, extension.ConnectionEvent) ([]extension.Instruction, error) {
	b.recordCall()
	return nil, nil
}

func (b *blockingControlPlane) NotifyTokenRefreshed(context.Context) ([]extension.Instruction, error) {
	b.recordCall()
	return nil, nil
}

func (b *blockingControlPlane) NotifyUploadCompleted(context.Context) ([]extension.Instruction, error) {
	b.recordCall()
	return nil, nil
}

func (b *blockingControlPlane) Close() error {
	return nil
}

type contextAwareControlPlane struct{}

func (*contextAwareControlPlane) Start(context.Context, extension.StartRequest) ([]extension.Instruction, error) {
	return nil, nil
}

func (*contextAwareControlPlane) SendTextLine(ctx context.Context, _ string) ([]extension.Instruction, error) {
	<-ctx.Done()
	return nil, ctx.Err()
}

func (*contextAwareControlPlane) NotifyConnection(context.Context, extension.ConnectionEvent) ([]extension.Instruction, error) {
	return nil, nil
}

func (*contextAwareControlPlane) NotifyTokenRefreshed(context.Context) ([]extension.Instruction, error) {
	return nil, nil
}

func (*contextAwareControlPlane) NotifyUploadCompleted(ctx context.Context) ([]extension.Instruction, error) {
	<-ctx.Done()
	return nil, ctx.Err()
}

func (*contextAwareControlPlane) Close() error {
	return nil
}

func TestSyncer_NotifyUploadCompleted_PropagatesControlPlaneError(t *testing.T) {
	t.Parallel()

	cp := &errorControlPlane{err: fmt.Errorf("boom")}
	s := &syncer{
		control: cp,
		scope:   logtest.NewScope(t),
	}

	err := s.NotifyUploadCompleted(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

type errorControlPlane struct {
	err error
}

func (e *errorControlPlane) Start(context.Context, extension.StartRequest) ([]extension.Instruction, error) {
	return nil, nil
}

func (e *errorControlPlane) SendTextLine(context.Context, string) ([]extension.Instruction, error) {
	return nil, nil
}

func (e *errorControlPlane) NotifyConnection(context.Context, extension.ConnectionEvent) ([]extension.Instruction, error) {
	return nil, nil
}

func (e *errorControlPlane) NotifyTokenRefreshed(context.Context) ([]extension.Instruction, error) {
	return nil, nil
}

func (e *errorControlPlane) NotifyUploadCompleted(context.Context) ([]extension.Instruction, error) {
	return nil, e.err
}

func (e *errorControlPlane) Close() error {
	return nil
}
