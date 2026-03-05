package client

import (
	"fmt"
	"strconv"
)

// AccessToken is a bearer token used for PowerSync HTTP requests.
type AccessToken string

// ClientID is the PowerSync client identifier.
type ClientID string

// WriteCheckpoint is the server write checkpoint value.
type WriteCheckpoint string

func (id ClientID) String() string { return string(id) }

func (cp WriteCheckpoint) String() string { return string(cp) }

// ParseInt parses the checkpoint into an integer operation ID.
func (cp WriteCheckpoint) ParseInt() (int64, error) {
	v, err := strconv.ParseInt(string(cp), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid write checkpoint %q: %w", cp, err)
	}
	return v, nil
}
