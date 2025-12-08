package workos

import (
	"net/http"
	"time"
)

const baseURL = "https://api.workos.com"

// Client provides access to WorkOS device code flow authentication.
type Client struct {
	baseURL    string
	clientID   string
	httpClient *http.Client
}

// NewClient creates a new WorkOS client.
func NewClient(clientID string) *Client {
	return &Client{
		baseURL:    baseURL,
		clientID:   clientID,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}
