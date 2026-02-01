package powersync

// Status represents the high-level state of the sync connection.
type Status string

// StatusUpdateMsg is a TUI message sent with current sync status.
// Used by the database syncer to notify the sync step of progress.
type StatusUpdateMsg struct {
	Status     Status
	SyncStatus *SyncStatus
	LastError  error
}

// SyncReadyMsg is a TUI message sent when sync has completed its first sync.
type SyncReadyMsg struct{}

// SyncStatusQueryMsg is a TUI message sent by the sync step to ask if sync is ready.
type SyncStatusQueryMsg struct{}

const (
	StatusDisconnected Status = "disconnected"
	StatusConnecting   Status = "connecting"
	StatusConnected    Status = "connected"
	StatusSyncing      Status = "syncing"
	StatusReconnecting Status = "reconnecting"
	StatusError        Status = "error"
)

// SyncStatus represents the detailed sync state from UpdateSyncStatus instructions.
type SyncStatus struct {
	Connected      bool              `json:"connected"`
	Connecting     bool              `json:"connecting"`
	PriorityStatus []PriorityStatus  `json:"priority_status"`
	Downloading    *DownloadProgress `json:"downloading"`
	Streams        []StreamStatus    `json:"streams"`
}

// PriorityStatus represents sync status for a specific priority level.
type PriorityStatus struct {
	Priority     int   `json:"priority"`
	LastSyncedAt *int  `json:"last_synced_at"`
	HasSynced    *bool `json:"has_synced"`
}

// StreamStatus represents the status of a sync stream subscription.
type StreamStatus struct {
	Name                    string          `json:"name"`
	Parameters              *string         `json:"parameters"`
	Priority                int             `json:"priority"`
	Active                  bool            `json:"active"`
	IsDefault               bool            `json:"is_default"`
	HasExplicitSubscription bool            `json:"has_explicit_subscription"`
	ExpiresAt               *int            `json:"expires_at"`
	LastSyncedAt            *int            `json:"last_synced_at"`
	Progress                *StreamProgress `json:"progress"`
}

// StreamProgress represents download progress for a stream.
type StreamProgress struct {
	Total      int `json:"total"`
	Downloaded int `json:"downloaded"`
}

// DownloadProgress represents the current download progress.
type DownloadProgress struct {
	Buckets map[string]BucketProgress `json:"buckets"`
}

// BucketProgress represents progress for a single bucket.
type BucketProgress struct {
	Priority    int `json:"priority"`
	AtLast      int `json:"at_last"`
	SinceLast   int `json:"since_last"`
	TargetCount int `json:"target_count"`
}

// TotalProgress returns the total progress across all buckets.
// Returns (downloaded, total) counts.
func (d *DownloadProgress) TotalProgress() (int, int) {
	if d == nil {
		return 0, 0
	}
	var downloaded, total int
	for _, b := range d.Buckets {
		downloaded += b.SinceLast
		total += b.TargetCount
	}
	return downloaded, total
}
