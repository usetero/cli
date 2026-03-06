package client

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	syncStreamPath      = "/sync/stream"
	writeCheckpointPath = "/write-checkpoint2.json"
)

// Client provides HTTP access to the PowerSync service.
type Client interface {
	SyncStream(ctx context.Context, req *SyncStreamRequest, handler LineHandler) error
	GetWriteCheckpoint(ctx context.Context, clientID ClientID) (WriteCheckpoint, error)
	SetToken(token AccessToken)
}

// HTTPDoer makes HTTP requests.
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// LineHandler is called for each NDJSON line from the sync stream.
type LineHandler func(line []byte) error

// HTTPClient is the concrete PowerSync HTTP client.
type HTTPClient struct {
	endpoint string
	token    AccessToken
	http     HTTPDoer
}

// NewClient creates a new HTTP PowerSync client.
func NewClient(endpoint string) *HTTPClient {
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		panic("powersync client requires endpoint")
	}
	if _, err := url.ParseRequestURI(endpoint); err != nil {
		panic(fmt.Sprintf("powersync client requires valid endpoint: %v", err))
	}
	return &HTTPClient{
		endpoint: endpoint,
		http:     http.DefaultClient,
	}
}

// SetHTTPDoer overrides the HTTP transport (for testing).
func (c *HTTPClient) SetHTTPDoer(doer HTTPDoer) {
	if doer != nil {
		c.http = doer
	}
}

// SetToken updates the auth token used by requests.
func (c *HTTPClient) SetToken(token AccessToken) {
	c.token = token
}

// SyncStream opens /sync/stream and feeds each NDJSON line to handler.
func (c *HTTPClient) SyncStream(ctx context.Context, req *SyncStreamRequest, handler LineHandler) error {
	u, err := c.endpointURL(syncStreamPath)
	if err != nil {
		return err
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), io.NopCloser(bytes.NewReader(body)))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	httpReq.ContentLength = int64(len(body))
	httpReq.Header.Set("Authorization", "Bearer "+string(c.token))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/x-ndjson")

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return &Error{Kind: ErrorKindTransient, Message: "connection failed", Err: err}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return &Error{Kind: classifyHTTPStatus(resp.StatusCode), StatusCode: resp.StatusCode, Message: string(msg)}
	}

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 64*1024), 16*1024*1024)
	for scanner.Scan() {
		if err := ctx.Err(); err != nil {
			return err
		}
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		if err := handler(line); err != nil {
			return err
		}
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read stream: %w", err)
	}
	return nil
}

// GetWriteCheckpoint fetches write checkpoint from /write-checkpoint2.json.
func (c *HTTPClient) GetWriteCheckpoint(ctx context.Context, clientID ClientID) (WriteCheckpoint, error) {
	u, err := c.endpointURL(writeCheckpointPath)
	if err != nil {
		return "", err
	}
	q := u.Query()
	q.Set("client_id", clientID.String())
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+string(c.token))

	resp, err := c.http.Do(req)
	if err != nil {
		return "", &Error{Kind: ErrorKindTransient, Message: "connection failed", Err: err}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return "", &Error{Kind: classifyHTTPStatus(resp.StatusCode), StatusCode: resp.StatusCode, Message: string(msg)}
	}

	var out struct {
		WriteCheckpoint WriteCheckpoint `json:"write_checkpoint"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("parse response: %w", err)
	}
	return out.WriteCheckpoint, nil
}

func (c *HTTPClient) endpointURL(path string) (*url.URL, error) {
	u, err := url.Parse(c.endpoint)
	if err != nil {
		return nil, fmt.Errorf("parse endpoint: %w", err)
	}
	u.Path = path
	u.RawQuery = ""
	u.Fragment = ""
	return u, nil
}
