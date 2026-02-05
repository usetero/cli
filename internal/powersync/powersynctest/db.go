// Package powersynctest provides test utilities for the powersync package.
//
// For database test helpers, use powersync/db/dbtest.
// For mock API clients, use powersync/api/apitest.
package powersynctest

import (
	"io"
	"log/slog"

	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/powersync"
	"github.com/usetero/cli/internal/powersync/api/apitest"
)

// NewSyncerWithMockClient creates a Syncer with a mock client for testing.
func NewSyncerWithMockClient(endpoint string, tokenRefresher powersync.TokenRefresher, mock *apitest.MockClient) powersync.Syncer {
	return powersync.NewSyncer(
		endpoint,
		tokenRefresher,
		discardLogger(),
		powersync.WithClientFactory(apitest.NewMockClientFactory(mock)),
	)
}

func discardLogger() log.Logger {
	return log.Wrap(slog.New(slog.NewTextHandler(io.Discard, nil)))
}
