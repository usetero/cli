package turn

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/app/chat/msgs"
	chatclient "github.com/usetero/cli/internal/chat"
	"github.com/usetero/cli/internal/domain"
)

// StartStream begins streaming the assistant response.
func (m *Model) StartStream(messages []domain.Message, chatContext []domain.ContextEntity) tea.Cmd {
	m.scope.Debug("starting stream", "message_count", len(messages))
	m.state = StateStreaming

	ctx, cancel := context.WithCancelCause(context.Background())
	updates := make(chan streamUpdate, 10)
	m.stream = &streamState{updates: updates, cancel: cancel}

	go func() {
		defer close(updates)

		req := chatclient.Request{
			ConversationID:  m.conversationID.String(),
			Messages:        messages,
			ContextEntities: chatContext,
		}

		var lastSnapshot *chatclient.StreamSnapshot
		result, err := m.chatClient.StreamSnapshots(ctx, req, func(s chatclient.StreamSnapshot) {
			ss := s
			lastSnapshot = &ss
			if !s.Done {
				updates <- streamUpdate{message: s.Message, status: s.Status}
			}
		})

		if err != nil {
			updates <- streamUpdate{err: err, done: true}
		} else {
			var lastMessage *domain.Message
			var status chatclient.StreamStatus
			var abort string
			if lastSnapshot != nil {
				lastMessage = lastSnapshot.Message
				status = lastSnapshot.Status
				abort = lastSnapshot.AbortReason
			}
			updates <- streamUpdate{message: lastMessage, status: status, abort: abort, result: result, done: true}
		}
	}()

	return m.nextStreamUpdate()
}

// handleStreamUpdate processes a stream update and fires messages.
func (m *Model) handleStreamUpdate(update streamUpdate) tea.Cmd {
	if update.err != nil {
		// context.Canceled is expected when Cancel() was called — don't show an error.
		if m.stream != nil && m.stream.done {
			return nil
		}
		errorClass := chatclient.ClassifyStreamError(update.err)
		m.scope.Error("stream error", "class", string(errorClass), "error", update.err)
		m.assistantMessage.Cancel()
		m.state = StateComplete
		turnID := m.userMessage.ID()
		return func() tea.Msg {
			return msgs.StreamFailed{TurnID: turnID, Err: update.err}
		}
	}

	if update.message == nil {
		return m.nextStreamUpdate()
	}

	// Set assistant message ID once we have it from the stream
	if m.assistantMessage.ID() == "" && update.message.ID != "" {
		m.assistantMessage.SetID(update.message.ID)
		// Populate content immediately to avoid empty render before message round-trips
		m.assistantMessage.SetContent(update.message.Content)
	}

	if update.done {
		if update.status == chatclient.StreamStatusAborted {
			reason := update.abort
			if reason == "" {
				reason = "context_canceled"
			}
			m.scope.Info("stream aborted", "reason", reason)
			m.assistantMessage.Cancel()
			m.state = StateComplete

			// User-cancelled stream: keep current behavior (do not persist).
			if reason == errUserCancelled.Error() {
				return nil
			}

			msg := update.message
			if msg == nil {
				msg = &domain.Message{}
			}
			if msg.ID == "" {
				msg.ID = domain.NewMessageID()
			}
			msg.StopReason = "aborted"

			turnID := m.userMessage.ID()
			return tea.Batch(
				func() tea.Msg {
					return msgs.StreamCompleted{
						TurnID:  turnID,
						Message: *msg,
					}
				},
				m.persistAssistantMessage(msg),
			)
		}

		if update.message.ID == "" {
			update.message.ID = domain.NewMessageID()
		}
		m.scope.Info("stream completed", "stop_reason", update.message.StopReason)

		// Extract metadata from stream result if present
		var title string
		var contextWindow, inputTokens, outputTokens int
		if update.result != nil && update.result.Metadata != nil {
			title = update.result.Metadata.Title
			contextWindow = update.result.Metadata.ContextWindow
			inputTokens = update.result.Metadata.InputTokens
			outputTokens = update.result.Metadata.OutputTokens
		}

		if update.message.StopReason == "tool_use" {
			m.toolTracker.setPendingFromContent(update.message.Content)
			m.scope.Info("awaiting tool results", "pending", m.toolTracker.pendingCount(), "already_collected", m.toolTracker.collectedCount())
			m.state = reduceOnStreamDone(update.message.StopReason, m.toolTracker.collectedCount(), m.toolTracker.pendingCount())
			if m.state == StateComplete {
				m.scope.Info("all tools already completed")
			}
		} else {
			m.state = reduceOnStreamDone(update.message.StopReason, m.toolTracker.collectedCount(), m.toolTracker.pendingCount())
		}

		// Fire StreamCompleted and persist
		turnID := m.userMessage.ID()
		return tea.Batch(
			func() tea.Msg {
				return msgs.StreamCompleted{
					TurnID:        turnID,
					Message:       *update.message,
					Title:         title,
					ContextWindow: contextWindow,
					InputTokens:   inputTokens,
					OutputTokens:  outputTokens,
				}
			},
			m.persistAssistantMessage(update.message),
		)
	}

	// Fire AssistantContentUpdated and continue
	turnID := m.userMessage.ID()
	return tea.Batch(
		func() tea.Msg {
			return msgs.AssistantContentUpdated{TurnID: turnID, Message: *update.message}
		},
		m.nextStreamUpdate(),
	)
}

// nextStreamUpdate returns a command that waits for the next stream update.
func (m *Model) nextStreamUpdate() tea.Cmd {
	if m.stream == nil || m.stream.done {
		return nil
	}

	userMsgID := m.userMessage.ID()
	updates := m.stream.updates

	return func() tea.Msg {
		update, ok := <-updates
		if !ok {
			return streamUpdateMsg{
				turnID: userMsgID,
				update: streamUpdate{done: true},
			}
		}
		return streamUpdateMsg{
			turnID: userMsgID,
			update: update,
		}
	}
}
