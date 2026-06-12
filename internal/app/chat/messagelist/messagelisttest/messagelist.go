// Package messagelisttest provides test helpers for the messagelist package.
package messagelisttest

import (
	"testing"

	"github.com/usetero/cli/internal/app/chat/messagelist"
	"github.com/usetero/cli/internal/app/chat/usecase"
	"github.com/usetero/cli/internal/boundary/chat/chattest"
	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/styles"
)

// New creates a messagelist.Model wired with test dependencies.
// Uses ephemeral in-memory chat deps, mock chat client, real theme, and test logger.
func New(t *testing.T, width, height int) *messagelist.Model {
	t.Helper()

	theme := styles.NewTheme(true)
	client := &chattest.MockClient{}
	scope := logtest.NewScope(t)
	runtimeDeps := usecase.NewRuntimeDeps(client)

	m := messagelist.New(theme, runtimeDeps, nil, scope)
	m.SetSize(width, height)
	return m
}
