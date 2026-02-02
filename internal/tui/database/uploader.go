package database

import (
	"context"
	"errors"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/chat"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/powersync"
	"github.com/usetero/cli/internal/sqlite"
	"github.com/usetero/cli/internal/upload"
)

// uploadEventMsg wraps an upload event for the bubbletea message loop (internal).
type uploadEventMsg struct {
	event upload.Event
}

// uploadDoneMsg is sent when the uploader goroutine exits (internal).
type uploadDoneMsg struct{}

// UploadEventMsg wraps upload events for other TUI components to listen for.
type UploadEventMsg struct {
	Event upload.Event
}

// TokenRefresher provides access tokens for authentication.
type TokenRefresher interface {
	GetAccessToken(ctx context.Context) (string, error)
}

// Uploader manages the upload loop lifecycle.
type Uploader struct {
	ctx             context.Context
	db              sqlite.Database
	powersyncClient powersync.Client
	tokenRefresher  TokenRefresher
	conversations   api.Conversations
	messages        chat.Messages
	logger          log.Logger

	uploader *upload.Uploader
	done     chan struct{}
}

// NewUploader creates a new uploader model.
func NewUploader(
	ctx context.Context,
	db sqlite.Database,
	powersyncClient powersync.Client,
	tokenRefresher TokenRefresher,
	conversations api.Conversations,
	messages chat.Messages,
	logger log.Logger,
) *Uploader {
	return &Uploader{
		ctx:             ctx,
		db:              db,
		powersyncClient: powersyncClient,
		tokenRefresher:  tokenRefresher,
		conversations:   conversations,
		messages:        messages,
		logger:          logger,
	}
}

// Start starts the upload loop and returns a command to listen for events.
func (u *Uploader) Start() tea.Cmd {
	u.uploader = upload.New(u.db, u.powersyncClient, u.tokenRefresher, u.conversations, u.messages, u.logger)
	u.done = make(chan struct{})

	go func() {
		defer close(u.done)
		if err := u.uploader.Run(u.ctx); err != nil && !errors.Is(err, context.Canceled) {
			u.logger.Error("upload loop error", "error", err)
		}
	}()

	return u.listenEvents()
}

// Update handles upload-related messages.
func (u *Uploader) Update(msg tea.Msg) tea.Cmd {
	switch e := msg.(type) {
	case uploadEventMsg:
		// Convert upload events to Bubble Tea messages and continue listening
		var cmds []tea.Cmd
		cmds = append(cmds, u.listenEvents())

		switch event := e.event.(type) {
		case upload.MessageProcessingEvent:
			u.logger.Debug("message processing", "conversationID", event.ConversationID, "messageID", event.UserMessageID)
			cmds = append(cmds, func() tea.Msg {
				return UploadEventMsg{Event: event}
			})
		case upload.StalledEvent:
			u.logger.Warn("upload stalled", "error", event.Error, "table", event.Table)
		case upload.RecoveredEvent:
			u.logger.Info("upload recovered", "stalledFor", event.StalledFor)
		case upload.SyncingEvent:
			u.logger.Debug("upload syncing", "count", event.ProcessedCount)
		}

		return tea.Batch(cmds...)

	case uploadDoneMsg:
		u.uploader = nil
	}

	return nil
}

// Wait blocks until the uploader goroutine exits.
func (u *Uploader) Wait() {
	if u.done != nil {
		<-u.done
	}
}

// listenEvents returns a command that waits for the next upload event.
func (u *Uploader) listenEvents() tea.Cmd {
	if u.uploader == nil {
		return nil
	}
	events := u.uploader.Events()
	return func() tea.Msg {
		event, ok := <-events
		if !ok {
			return uploadDoneMsg{}
		}
		return uploadEventMsg{event: event}
	}
}
