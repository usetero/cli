package chat

import (
	"context"
	"fmt"
	"strings"

	domainchat "github.com/usetero/cli/internal/domains/chat"
	infrachat "github.com/usetero/cli/internal/infrastructure/chat"
)

func (r *Runtime) SendUserText(ctx context.Context, text string) error {
	text = strings.TrimSpace(text)
	if text == "" {
		return fmt.Errorf("text is required")
	}
	if err := r.requireReadyToSend(); err != nil {
		return err
	}

	conversationID, err := r.ensureConversation(ctx)
	if err != nil {
		return err
	}
	messageID, err := r.persistUserMessage(ctx, conversationID, text)
	if err != nil {
		return err
	}

	r.mu.Lock()
	r.history = append(r.history, userTextWireMessage(text))
	r.state.Messages = append(r.state.Messages, MessageView{
		ID:      messageID,
		Role:    domainchat.RoleUser,
		Content: text,
	})
	r.state.Error = ""
	r.state.Streaming = true
	r.state.CanSend = false
	streamCtx, cancel := context.WithCancel(ctx)
	r.cancel = cancel
	requestMessages := append([]infrachat.Message(nil), r.history...)
	snapshot := cloneState(r.state)
	r.mu.Unlock()
	r.publish(snapshot)

	go r.runStream(streamCtx, conversationID, requestMessages)
	return nil
}
