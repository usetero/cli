package chat

import (
	"github.com/google/uuid"
	domainchat "github.com/usetero/cli/internal/domains/chat"
	chattools "github.com/usetero/cli/internal/domains/chat/tools"
)

func toWireConversationID(id domainchat.ConversationID) string {
	return string(id)
}

func toWireToolName(name chattools.Name) string {
	return string(name)
}

func toDomainToolName(name string) chattools.Name {
	return chattools.Name(name)
}

func newDomainMessageID() domainchat.MessageID {
	return domainchat.MessageID(uuid.NewString())
}
