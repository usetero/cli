package uploader

import (
	"context"
	"fmt"
	"time"

	"github.com/usetero/cli/internal/infrastructure/logging"
	psclient "github.com/usetero/cli/internal/infrastructure/powersync/client"
	psdb "github.com/usetero/cli/internal/infrastructure/powersync/db"
)

type batchResult struct {
	processedCount int
	stalledEntry   *psdb.Mutation
}

type batchProcessor struct {
	store      *psdb.Store
	client     psclient.Client
	tokens     TokenSource
	log        logging.Scope
	policy     RunPolicy
	dispatcher *mutationDispatcher
	notifier   SyncNotifier
	wait       func(ctx context.Context, d time.Duration)
}

func newBatchProcessor(store *psdb.Store, client psclient.Client, tokens TokenSource, log logging.Scope) *batchProcessor {
	return &batchProcessor{
		store:      store,
		client:     client,
		tokens:     tokens,
		log:        log,
		policy:     DefaultRunPolicy(),
		dispatcher: newMutationDispatcher(),
		wait: func(ctx context.Context, d time.Duration) {
			select {
			case <-ctx.Done():
			case <-time.After(d):
			}
		},
	}
}

func (p *batchProcessor) UploadNextBatch(ctx context.Context) (batchResult, error) {
	batch, err := p.nextBatch(ctx)
	if err != nil || len(batch) == 0 {
		return batchResult{}, err
	}

	if err := p.prepareClient(ctx); err != nil {
		return batchResult{}, err
	}

	for i := range batch {
		if err := p.handleMutationWithRetry(ctx, batch[i]); err != nil {
			stalledEntry := batch[i]
			return batchResult{stalledEntry: &stalledEntry}, err
		}
	}

	if err := p.completeBatch(ctx, batch); err != nil {
		return batchResult{}, err
	}
	p.notifyUploadCompleted(ctx)

	return batchResult{processedCount: len(batch)}, nil
}

func (p *batchProcessor) nextBatch(ctx context.Context) ([]psdb.Mutation, error) {
	batch, err := p.store.NextMutationBatch(ctx)
	if err != nil {
		return nil, fmt.Errorf("get next mutation batch: %w", err)
	}
	return batch, nil
}

func (p *batchProcessor) prepareClient(ctx context.Context) error {
	token, err := p.tokens.GetAccessToken(ctx)
	if err != nil {
		return fmt.Errorf("get access token: %w", err)
	}
	p.client.SetToken(token)
	return nil
}

func (p *batchProcessor) handleMutationWithRetry(ctx context.Context, mutation psdb.Mutation) error {
	var lastErr error
	for attempt := 0; attempt <= p.policy.MaxRetries; attempt++ {
		if attempt > 0 {
			p.wait(ctx, p.policy.RetryDelay*time.Duration(attempt))
		}
		if err := p.dispatcher.Dispatch(ctx, mutation); err != nil {
			lastErr = err
			continue
		}
		return nil
	}
	return lastErr
}

func (p *batchProcessor) completeBatch(ctx context.Context, batch []psdb.Mutation) error {
	clientID, err := p.store.ClientID(ctx)
	if err != nil {
		return fmt.Errorf("get client id: %w", err)
	}
	checkpoint, err := p.client.GetWriteCheckpoint(ctx, psclient.ClientID(clientID))
	if err != nil {
		return fmt.Errorf("get write checkpoint: %w", err)
	}
	checkpointInt, err := checkpoint.ParseInt()
	if err != nil {
		return fmt.Errorf("parse write checkpoint: %w", err)
	}

	lastID := batch[len(batch)-1].ID
	if err := p.store.CompleteUploadedBatch(ctx, lastID, psdb.OpID(checkpointInt)); err != nil {
		return fmt.Errorf("complete uploaded batch: %w", err)
	}
	return nil
}

func (p *batchProcessor) notifyUploadCompleted(ctx context.Context) {
	if p.notifier == nil {
		return
	}
	if err := p.notifier.NotifyUploadCompleted(ctx); err != nil {
		p.log.Warn("sync notifier failed", "error", err)
	}
}
