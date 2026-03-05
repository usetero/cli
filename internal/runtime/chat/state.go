package chat

import domainchat "github.com/usetero/cli/internal/domains/chat"

type MessageView struct {
	ID      domainchat.MessageID
	Role    domainchat.Role
	Content string
}

type State struct {
	ConversationID domainchat.ConversationID
	Messages       []MessageView
	Streaming      bool
	CanSend        bool
	Error          string
}
