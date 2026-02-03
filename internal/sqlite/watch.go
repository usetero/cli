package sqlite

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"time"
)

const (
	// pollInterval is how often we check for table changes.
	pollInterval = 50 * time.Millisecond
)

// Subscription receives notifications when watched tables change.
// Call Stop() when done to prevent leaks.
type Subscription struct {
	ch     chan []Table
	db     *database
	closed bool
	mu     sync.Mutex
}

// Changes returns a channel that receives table names when they change.
func (s *Subscription) Changes() <-chan []Table {
	return s.ch
}

// Stop unsubscribes and closes the channel.
func (s *Subscription) Stop() {
	s.db.unsubscribe(s)
}

// watchState manages change detection and subscriber notification.
type watchState struct {
	mu          sync.RWMutex
	subscribers map[*Subscription]struct{}
	installed   bool
	cancel      context.CancelFunc
}

// installUpdateHooks enables change tracking via PowerSync's update hooks.
// Starts a background goroutine that polls for changes and notifies subscribers.
// The goroutine stops when ctx is cancelled.
// Called automatically by Open() when the PowerSync extension is loaded.
func (d *database) installUpdateHooks(ctx context.Context) error {
	d.watch.mu.Lock()
	defer d.watch.mu.Unlock()

	if d.watch.installed {
		return nil
	}

	_, err := d.db.ExecContext(ctx, "SELECT powersync_update_hooks('install')")
	if err != nil {
		return err
	}

	// Start background polling goroutine
	pollCtx, cancel := context.WithCancel(ctx)
	d.watch.cancel = cancel
	go d.pollForChanges(pollCtx)

	d.watch.installed = true
	return nil
}

// pollForChanges runs in the background, checking for table changes periodically.
func (d *database) pollForChanges(ctx context.Context) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			d.checkForChanges()
		}
	}
}

// Subscribe returns a Subscription that receives table names when they change.
// The channel has a buffer of 1 to avoid blocking the poller.
// Call Stop() on the subscription when done to prevent leaks.
func (d *database) Subscribe() *Subscription {
	d.watch.mu.Lock()
	defer d.watch.mu.Unlock()

	sub := &Subscription{
		ch: make(chan []Table, 1),
		db: d,
	}
	if d.watch.subscribers == nil {
		d.watch.subscribers = make(map[*Subscription]struct{})
	}
	d.watch.subscribers[sub] = struct{}{}
	return sub
}

// unsubscribe removes a subscriber and closes its channel.
func (d *database) unsubscribe(sub *Subscription) {
	sub.mu.Lock()
	defer sub.mu.Unlock()

	if sub.closed {
		return
	}

	d.watch.mu.Lock()
	defer d.watch.mu.Unlock()

	if _, ok := d.watch.subscribers[sub]; ok {
		delete(d.watch.subscribers, sub)
		close(sub.ch)
		sub.closed = true
	}
}

// checkForChanges queries the update hooks and notifies subscribers.
func (d *database) checkForChanges() {
	d.watch.mu.RLock()
	if !d.watch.installed || len(d.watch.subscribers) == 0 {
		d.watch.mu.RUnlock()
		return
	}
	d.watch.mu.RUnlock()

	// Query for changed tables
	var result string
	err := d.db.QueryRowContext(context.Background(), "SELECT powersync_update_hooks('get')").Scan(&result)
	if err != nil {
		return
	}

	// Parse JSON array of raw table names
	var rawTables []string
	if err := json.Unmarshal([]byte(result), &rawTables); err != nil {
		return
	}

	if len(rawTables) == 0 {
		return
	}

	// Convert to typed tables, deduplicating
	seen := make(map[Table]bool)
	var tables []Table
	for _, raw := range rawTables {
		// Try direct lookup first
		if t, ok := knownTables[raw]; ok {
			if !seen[t] {
				seen[t] = true
				tables = append(tables, t)
			}
			continue
		}
		// Try stripping ps_data__ prefix
		if strings.HasPrefix(raw, "ps_data__") {
			name := strings.TrimPrefix(raw, "ps_data__")
			if t, ok := knownTables[name]; ok {
				if !seen[t] {
					seen[t] = true
					tables = append(tables, t)
				}
			}
		}
	}

	if len(tables) == 0 {
		return
	}

	// Notify subscribers (non-blocking)
	d.watch.mu.RLock()
	defer d.watch.mu.RUnlock()

	for sub := range d.watch.subscribers {
		select {
		case sub.ch <- tables:
		default:
			// Channel full, skip (subscriber is slow)
		}
	}
}
