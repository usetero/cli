package chat

import (
	"time"

	conversationsdb "github.com/usetero/cli/internal/domains/chat/db/conversationsgen"
	messagesdb "github.com/usetero/cli/internal/domains/chat/db/messagesgen"
)

func toConversationsDBConversationID(id ConversationID) string {
	return string(id)
}

func toMessagesDBMessageID(id MessageID) string {
	return string(id)
}

func toMessagesDBConversationID(id ConversationID) string {
	return string(id)
}

func toMessagesDBRole(role Role) string {
	return string(role)
}

func fromConversationsDBConversation(row conversationsdb.Conversation) Conversation {
	conversation := Conversation{
		ID:    ConversationID(row.ID),
		Title: row.Title,
	}
	if t, err := time.Parse(time.RFC3339, row.CreatedAt); err == nil {
		conversation.CreatedAt = t
	}
	return conversation
}

func fromMessagesDBMessage(row messagesdb.Message) Message {
	message := Message{
		ID:             MessageID(row.ID),
		ConversationID: ConversationID(row.ConversationID),
		Role:           Role(row.Role),
		Content:        row.Content,
	}
	if t, err := time.Parse(time.RFC3339, row.CreatedAt); err == nil {
		message.CreatedAt = t
	}
	return message
}
