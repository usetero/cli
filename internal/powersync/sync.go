package powersync

import (
	"context"
	"fmt"

	"github.com/usetero/cli/internal/sqlite"
)

// Sync manages the PowerSync connection for a database.
type Sync struct {
	config *Config
	db     sqlite.Database
	client *Client
}

// NewSync creates a new PowerSync sync manager.
func NewSync(config *Config) *Sync {
	return &Sync{
		config: config,
	}
}

// Start loads the PowerSync extension, initializes the schema, and starts syncing.
func (s *Sync) Start(ctx context.Context, db sqlite.Database, accountID, token string) error {
	if s.client != nil {
		return fmt.Errorf("already started")
	}

	s.db = db

	// Load the PowerSync extension into SQLite
	extPath, err := ExtensionPath()
	if err != nil {
		return fmt.Errorf("get extension path: %w", err)
	}

	if err := db.LoadExtension(extPath); err != nil {
		return fmt.Errorf("load extension: %w", err)
	}

	// Initialize the schema
	schema := sqlite.DefaultSchema()
	schemaJSON, err := schema.JSON()
	if err != nil {
		return fmt.Errorf("marshal schema: %w", err)
	}

	_, err = db.Exec("SELECT powersync_replace_schema(?)", schemaJSON)
	if err != nil {
		return fmt.Errorf("replace schema: %w", err)
	}

	// Create and start sync client
	s.client = NewClient(s.config.Endpoint, db)
	if err := s.client.Connect(ctx, accountID, token); err != nil {
		return fmt.Errorf("start sync: %w", err)
	}

	return nil
}

// Stop stops syncing. The database remains open.
func (s *Sync) Stop() {
	if s.client != nil {
		s.client.Disconnect()
		s.client = nil
	}
	s.db = nil
}

// Status returns the current sync status.
func (s *Sync) Status() Status {
	if s.client == nil {
		return StatusDisconnected
	}
	return s.client.Status()
}

// IsRunning returns true if sync is active.
func (s *Sync) IsRunning() bool {
	return s.client != nil
}
