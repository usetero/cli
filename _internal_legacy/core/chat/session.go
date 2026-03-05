package chat

import "github.com/usetero/cli/internal/domain"

// Session owns in-memory message history for one active chat loop.
// Persistence is handled separately by callers.
type Session struct {
	conversationID domain.ConversationID
	history        []domain.Message
}

// NewSession creates a session with an initial history snapshot.
func NewSession(conversationID domain.ConversationID, initial []domain.Message) *Session {
	return &Session{
		conversationID: conversationID,
		history:        cloneMessages(initial),
	}
}

// Messages returns a defensive copy of the current history.
func (s *Session) Messages() []domain.Message {
	return cloneMessages(s.history)
}

// AppendMessage appends a message to history.
func (s *Session) AppendMessage(message domain.Message) {
	if message.ID == "" {
		return
	}
	s.history = append(s.history, cloneMessage(message))
}

// AppendUserTextMessage appends a user text message and returns it.
func (s *Session) AppendUserTextMessage(messageID domain.MessageID, text string) domain.Message {
	msg := domain.Message{
		ID:             messageID,
		ConversationID: s.conversationID,
		Role:           domain.RoleUser,
		Content: []domain.Block{{
			Index: 0,
			Type:  domain.BlockTypeText,
			Text:  &domain.TextBlock{Content: text},
		}},
	}
	s.history = append(s.history, msg)
	return msg
}

// AppendUserToolResultsMessage appends a user tool_result message and returns it.
func (s *Session) AppendUserToolResultsMessage(messageID domain.MessageID, results []domain.ToolResult) domain.Message {
	msg := buildToolResultMessage(s.conversationID, messageID, results)
	s.history = append(s.history, msg)
	return msg
}

// RecordAssistantMessage upserts an assistant message in history by message ID.
func (s *Session) RecordAssistantMessage(message domain.Message) {
	if message.ID == "" {
		return
	}
	message.Role = domain.RoleAssistant
	for i := range s.history {
		if s.history[i].ID == message.ID {
			s.history[i] = cloneMessage(message)
			return
		}
	}
	s.history = append(s.history, cloneMessage(message))
}

// RemoveMessagesByID removes all messages whose IDs are in ids.
func (s *Session) RemoveMessagesByID(ids []domain.MessageID) {
	if len(ids) == 0 || len(s.history) == 0 {
		return
	}
	drop := make(map[domain.MessageID]struct{}, len(ids))
	for _, id := range ids {
		drop[id] = struct{}{}
	}
	kept := s.history[:0]
	for _, msg := range s.history {
		if _, ok := drop[msg.ID]; ok {
			continue
		}
		kept = append(kept, msg)
	}
	s.history = kept
}

func buildToolResultMessage(conversationID domain.ConversationID, messageID domain.MessageID, results []domain.ToolResult) domain.Message {
	blocks := make([]domain.Block, len(results))
	for i, result := range results {
		r := result
		blocks[i] = domain.Block{
			Index:      i,
			Type:       domain.BlockTypeToolResult,
			ToolResult: &r,
		}
	}
	return domain.Message{
		ID:             messageID,
		ConversationID: conversationID,
		Role:           domain.RoleUser,
		Content:        blocks,
	}
}

func cloneMessages(in []domain.Message) []domain.Message {
	out := make([]domain.Message, len(in))
	for i := range in {
		out[i] = cloneMessage(in[i])
	}
	return out
}

func cloneMessage(in domain.Message) domain.Message {
	out := in
	if in.Content != nil {
		out.Content = make([]domain.Block, len(in.Content))
		for i := range in.Content {
			out.Content[i] = cloneBlock(in.Content[i])
		}
	}
	return out
}

func cloneBlock(in domain.Block) domain.Block {
	out := in
	if in.Text != nil {
		v := *in.Text
		out.Text = &v
	}
	if in.Thinking != nil {
		v := *in.Thinking
		out.Thinking = &v
	}
	if in.ToolUse != nil {
		v := *in.ToolUse
		if in.ToolUse.Input != nil {
			v.Input = append([]byte(nil), in.ToolUse.Input...)
		}
		out.ToolUse = &v
	}
	if in.ToolResult != nil {
		v := *in.ToolResult
		if in.ToolResult.Content != nil {
			v.Content = cloneAnyMap(in.ToolResult.Content)
		}
		out.ToolResult = &v
	}
	return out
}

func cloneAnyMap(in map[string]any) map[string]any {
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
