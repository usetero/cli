package usecase

import chatboundary "github.com/usetero/cli/internal/boundary/chat"

type ChatBoundaryStreamErrorMapper struct{}

func NewChatBoundaryStreamErrorMapper() *ChatBoundaryStreamErrorMapper {
	return &ChatBoundaryStreamErrorMapper{}
}

func (m *ChatBoundaryStreamErrorMapper) Classify(err error) string {
	return string(chatboundary.ClassifyStreamError(err))
}

func (m *ChatBoundaryStreamErrorMapper) UserFacing(err error) string {
	return chatboundary.UserFacingStreamError(err)
}
