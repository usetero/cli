package syncer

import "fmt"

// AccountID identifies the account scope used for PowerSync parameters.
type AccountID string

func (id AccountID) String() string { return string(id) }

// Validate checks account ID is present.
func (id AccountID) Validate() error {
	if id == "" {
		return fmt.Errorf("%w: account id is required", ErrInvalidInput)
	}
	return nil
}

// AccessToken is an auth token used for PowerSync HTTP requests.
type AccessToken string
