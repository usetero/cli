package usecase

import chatclient "github.com/usetero/cli/internal/api/chatclient"

type ChatClientStreamErrorMapper struct{}

func NewChatClientStreamErrorMapper() *ChatClientStreamErrorMapper {
	return &ChatClientStreamErrorMapper{}
}

func (m *ChatClientStreamErrorMapper) Classify(err error) string {
	return string(chatclient.ClassifyStreamError(err))
}

func (m *ChatClientStreamErrorMapper) UserFacing(err error) string {
	return chatclient.UserFacingStreamError(err)
}
