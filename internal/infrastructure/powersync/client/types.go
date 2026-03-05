package client

import "encoding/json"

// SyncStreamRequest is the request body for the sync stream endpoint.
type SyncStreamRequest struct {
	Buckets         []BucketRequest     `json:"buckets"`
	IncludeChecksum bool                `json:"include_checksum"`
	RawData         bool                `json:"raw_data"`
	BinaryData      bool                `json:"binary_data"`
	ClientID        ClientID            `json:"client_id"`
	Parameters      map[string]any      `json:"parameters,omitempty"`
	Streams         *StreamSubscription `json:"streams,omitempty"`
	AppMetadata     json.RawMessage     `json:"app_metadata,omitempty"`
}

// BucketRequest specifies a bucket to sync and the last known checkpoint.
type BucketRequest struct {
	Name  BucketName `json:"name"`
	After OpID       `json:"after"`
}

// BucketName identifies a PowerSync bucket.
type BucketName string

// OpID identifies an operation checkpoint in a stream/bucket.
type OpID string

// StreamSubscription defines stream subscription preferences.
type StreamSubscription struct {
	IncludeDefaults bool                          `json:"include_defaults"`
	Subscriptions   []RequestedStreamSubscription `json:"subscriptions"`
}

// RequestedStreamSubscription is a request to subscribe to a stream.
type RequestedStreamSubscription struct {
	Stream           string `json:"stream"`
	Parameters       string `json:"parameters,omitempty"`
	OverridePriority *int   `json:"override_priority,omitempty"`
}
