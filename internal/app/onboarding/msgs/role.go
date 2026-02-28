package msgs

// Role constants.
const (
	RolePlatform = "platform"
	RoleEngineer = "engineer"
)

// RoleSelected is emitted when user selects their role.
type RoleSelected struct {
	Role string
}
