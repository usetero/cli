package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Client communicates with the PowerSync API.
type Client struct {
	cfg Config
}

// NewClient creates a new PowerSync API client.
func NewClient(cfg Config) *Client {
	return &Client{cfg: cfg}
}

// FetchSchema fetches the database schema from PowerSync.
func (c *Client) FetchSchema() (*SchemaResponse, error) {
	req, err := http.NewRequest("POST", c.cfg.URL+"/api/admin/v1/schema", strings.NewReader("{}"))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.cfg.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("schema API returned %d: %s", resp.StatusCode, body)
	}

	var result SchemaResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

// FetchSyncRules fetches the sync rules from PowerSync.
func (c *Client) FetchSyncRules() (*SyncRulesResponse, error) {
	req, err := http.NewRequest("GET", c.cfg.URL+"/api/sync-rules/v1/current", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.cfg.Token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("sync-rules API returned %d: %s", resp.StatusCode, body)
	}

	var result SyncRulesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}
