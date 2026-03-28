package chat

import (
	"time"

	conversationsdb "github.com/usetero/cli/internal/domains/chat/db/conversationsgen"
	messagesdb "github.com/usetero/cli/internal/domains/chat/db/messagesgen"
)

func ptrString(v string) *string {
	return &v
}

func derefString(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

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

func fromConversationsDBListRow(row conversationsdb.ListRow) Conversation {
	conversation := Conversation{
		ID:    ConversationID(derefString(row.ID)),
		Title: row.Title,
	}
	if t, err := time.Parse(time.RFC3339, derefString(row.CreatedAt)); err == nil {
		conversation.CreatedAt = t
	}
	return conversation
}

func fromMessagesDBListRow(row messagesdb.ListByConversationRow) Message {
	message := Message{
		ID:             MessageID(derefString(row.ID)),
		ConversationID: ConversationID(derefString(row.ConversationID)),
		Role:           Role(derefString(row.Role)),
		Content:        derefString(row.Content),
	}
	if t, err := time.Parse(time.RFC3339, derefString(row.CreatedAt)); err == nil {
		message.CreatedAt = t
	}
	return message
}
