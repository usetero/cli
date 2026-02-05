package domain

import "github.com/google/uuid"

// AccountID is a unique identifier for an account.
type AccountID string

// NewAccountID generates a new unique AccountID.
func NewAccountID() AccountID { return AccountID(uuid.New().String()) }

func (id AccountID) String() string { return string(id) }

// Account represents a billing/Datadog account within an organization.
type Account struct {
	ID   AccountID `json:"id"`
	Name string    `json:"name"`
}

// FilterValue returns the string used for filtering/searching.
func (a Account) FilterValue() string { return a.Name }
