package role

import "github.com/usetero/cli/internal/domains/preferences"

// SubmittedMsg reports that the user confirmed a role choice.
type SubmittedMsg struct {
	Role preferences.Role
}
