package account

import accountruntime "github.com/usetero/cli/internal/runtime/account"

// RuntimeEventMsg carries one account-runtime event back into the Bubble Tea loop.
type RuntimeEventMsg struct {
	Event accountruntime.Event
}

// RuntimeClosedMsg reports that the current account runtime has shut down.
type RuntimeClosedMsg struct{}
