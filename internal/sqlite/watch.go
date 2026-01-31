package sqlite

import (
	"context"
	"encoding/json"
	"sync"
)

// Subscription represents an active change subscription.
// Call Stop() when done to prevent leaks.
type Subscription struct {
	ch     chan []string
	db     *DB
	closed bool
}

// Changes returns a receive-only channel of table names that changed.
func (s *Subscription) Changes() <-chan []string {
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
}

// InstallUpdateHooks enables change tracking via PowerSync's update hooks.
// This must be called after loading the PowerSync extension.
func (d *DB) InstallUpdateHooks(ctx context.Context) error {
	d.watch.mu.Lock()
	defer d.watch.mu.Unlock()

	if d.watch.installed {
		return nil
	}

	_, err := d.db.ExecContext(ctx, "SELECT powersync_update_hooks('install')")
	if err != nil {
		return err
	}

	d.watch.installed = true
	return nil
}

// Subscribe returns a Subscription that receives table names when they change.
// The channel has a buffer of 1 to avoid blocking writers.
// Call Stop() on the subscription when done to prevent leaks.
func (d *DB) Subscribe() *Subscription {
	d.watch.mu.Lock()
	defer d.watch.mu.Unlock()

	sub := &Subscription{
		ch: make(chan []string, 1),
		db: d,
	}
	if d.watch.subscribers == nil {
		d.watch.subscribers = make(map[*Subscription]struct{})
	}
	d.watch.subscribers[sub] = struct{}{}
	return sub
}

// unsubscribe removes a subscriber and closes its channel.
func (d *DB) unsubscribe(sub *Subscription) {
	d.watch.mu.Lock()
	defer d.watch.mu.Unlock()

	if sub.closed {
		return
	}

	if _, ok := d.watch.subscribers[sub]; ok {
		delete(d.watch.subscribers, sub)
		close(sub.ch)
		sub.closed = true
	}
}

// checkForChanges queries the update hooks and notifies subscribers.
// Called internally after write operations.
func (d *DB) checkForChanges() {
	d.watch.mu.RLock()
	if !d.watch.installed || len(d.watch.subscribers) == 0 {
		d.watch.mu.RUnlock()
		return
	}
	d.watch.mu.RUnlock()

	// Query for changed tables
	// Use background context since this is called after writes complete
	var result string
	err := d.db.QueryRowContext(context.Background(), "SELECT powersync_update_hooks('get')").Scan(&result)
	if err != nil {
		return // Silently ignore errors - don't break writes
	}

	// Parse JSON array
	var tables []string
	if err := json.Unmarshal([]byte(result), &tables); err != nil {
		return
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
