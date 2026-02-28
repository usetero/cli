package msgs

import "github.com/usetero/cli/internal/auth"

// Authenticated is emitted when authentication succeeds.
type Authenticated struct {
	User auth.User
}
