// Package messagelisttest provides test helpers for the messagelist package.
package messagelisttest

import (
	"testing"

	"github.com/usetero/cli/internal/api/chatclient/chattest"
	"github.com/usetero/cli/internal/app/chat/messagelist"
	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/powersync/db/dbtest"
	"github.com/usetero/cli/internal/styles"
)

// New creates a messagelist.Model wired with test dependencies.
// Uses a real SQLite database, mock chat client, real theme, and test logger.
func New(t *testing.T, width, height int) *messagelist.Model {
	t.Helper()

	theme := styles.NewTheme(true)
	db := dbtest.OpenTestDB(t)
	client := &chattest.MockClient{}
	scope := logtest.NewScope(t)

	m := messagelist.New(theme, db, client, nil, scope)
	m.SetSize(width, height)
	return m
}
