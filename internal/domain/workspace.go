package domain

import "github.com/google/uuid"

// WorkspaceID is a unique identifier for a workspace.
type WorkspaceID string

// NewWorkspaceID generates a new unique WorkspaceID.
func NewWorkspaceID() WorkspaceID { return WorkspaceID(uuid.New().String()) }

func (id WorkspaceID) String() string { return string(id) }

// Workspace represents a workspace within an account.
type Workspace struct {
	ID   WorkspaceID `json:"id"`
	Name string      `json:"name"`
}
