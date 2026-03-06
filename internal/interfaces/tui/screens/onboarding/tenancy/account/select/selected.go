package accountselect

import "github.com/usetero/cli/internal/domains/tenancy"

// SelectedMsg reports that the user confirmed an account selection.
type SelectedMsg struct {
	AccountID tenancy.AccountID
}
