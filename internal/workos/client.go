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
	audiences  []string
	httpClient *http.Client
}

// NewClient creates a new WorkOS client.
// Audiences are included in token requests so JWTs contain the aud claim.
// Both the Tero API and PowerSync validate audience.
func NewClient(clientID string, audiences ...string) *Client {
	if len(audiences) == 0 {
		panic("workos: at least one audience is required")
	}
	return &Client{
		baseURL:    baseURL,
		clientID:   clientID,
		audiences:  audiences,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}
