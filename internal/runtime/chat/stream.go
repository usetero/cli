package chat

import (
	"context"

	domainchat "github.com/usetero/cli/internal/domains/chat"
	infrachat "github.com/usetero/cli/internal/infrastructure/chat"
)

func (r *Runtime) runStream(ctx context.Context, conversationID domainchat.ConversationID, messages []infrachat.Message) {
	for {
		assistantIndex := r.appendAssistantPlaceholder()
		assistantText := ""
		toolUses := make([]infrachat.ToolUse, 0, 2)

		_, err := r.client.Stream(ctx, infrachat.Request{
			ConversationID: toWireConversationID(conversationID),
			Messages:       messages,
			Tools:          r.toolDefinitions(),
		}, func(ev infrachat.Event) {
			r.mu.Lock()
			defer r.mu.Unlock()

			switch ev.Type {
			case infrachat.EventTypeTextDelta:
				assistantText += ev.TextContent
				r.state.Messages[assistantIndex].Content = assistantText
			case infrachat.EventTypeToolUse:
				if ev.ToolUse != nil {
					toolUses = append(toolUses, *ev.ToolUse)
				}
			}
			r.publishLocked()
		})

		if err != nil {
			r.finishStreamWithError(err)
			return
		}

		if len(toolUses) == 0 {
			r.finishStreamWithAssistantText(assistantText)
			return
		}

		toolResults, summary := r.executeToolUses(ctx, toolUses)

		r.mu.Lock()
		r.history = append(r.history, assistantToolUseWireMessage(toolUses...))
		r.history = append(r.history, toolResultWireMessage(toolResults...))
		r.state.Messages = append(r.state.Messages, MessageView{
			ID:      newDomainMessageID(),
			Role:    domainchat.RoleUser,
			Content: summary,
		})
		messages = append([]infrachat.Message(nil), r.history...)
		r.mu.Unlock()
	}
}

func (r *Runtime) appendAssistantPlaceholder() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	idx := len(r.state.Messages)
	r.state.Messages = append(r.state.Messages, MessageView{
		ID:      newDomainMessageID(),
		Role:    domainchat.RoleAssistant,
		Content: "",
	})
	r.publishLocked()
	return idx
}

func (r *Runtime) finishStreamWithError(err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cancel = nil
	r.state.Streaming = false
	r.state.CanSend = true
	r.state.Error = err.Error()
	r.publishLocked()
}

func (r *Runtime) finishStreamWithAssistantText(assistantText string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cancel = nil
	r.state.Streaming = false
	r.state.CanSend = true
	r.state.Error = ""
	r.history = append(r.history, assistantTextWireMessage(assistantText))
	r.publishLocked()
}
