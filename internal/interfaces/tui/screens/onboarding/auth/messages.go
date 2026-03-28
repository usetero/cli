package auth

import "github.com/usetero/cli/internal/domains/identity"

type deviceFlowStartedMsg struct {
	Flow identity.DeviceFlow
}

type deviceFlowCompletedMsg struct {
	User identity.User
}

type deviceFlowFailedMsg struct {
	Err error
}

type browserOpenedMsg struct {
	Err error
}
