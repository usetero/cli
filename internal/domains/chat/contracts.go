package chat

var (
	_ ConversationService = (*LocalConversationService)(nil)
	_ MessageService      = (*LocalMessageService)(nil)
)
