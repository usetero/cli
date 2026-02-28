package usecase

import (
	chatclient "github.com/usetero/cli/internal/api/chatclient"
)

func ClassifyStreamError(err error) string {
	return string(chatclient.ClassifyStreamError(err))
}

func UserFacingStreamError(err error) string {
	return chatclient.UserFacingStreamError(err)
}
