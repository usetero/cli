package accountselect

import "github.com/usetero/cli/internal/domains/tenancy"

// SelectedMsg reports an account choice from the select page.
type SelectedMsg struct {
	AccountID tenancy.AccountID
}
