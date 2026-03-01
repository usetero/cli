package upload

import (
	"context"

	api "github.com/usetero/cli/internal/boundary/graphql"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/powersync/db"
)

// serviceHandler handles uploading service mutations to the GraphQL API.
// Services are server-owned — the client only patches them (e.g. enable/disable).
type serviceHandler struct {
	services api.Services
	scope    log.Scope
}

func newServiceHandler(services api.Services, scope log.Scope) *serviceHandler {
	return &serviceHandler{
		services: services,
		scope:    scope.Child("services"),
	}
}

func (h *serviceHandler) Handle(ctx context.Context, entry *db.CrudEntry, emit Emitter) error {
	_ = emit

	switch entry.Op {
	case db.OpPatch:
		return h.handlePatch(ctx, entry)
	default:
		h.scope.Warn("unsupported service op, dropping", "op", entry.Op, "rowId", entry.RowID)
		return nil
	}
}

func (h *serviceHandler) handlePatch(ctx context.Context, entry *db.CrudEntry) error {
	if enabledVal, ok := entry.Data["enabled"]; ok {
		id := domain.ServiceID(entry.RowID)
		if toBool(enabledVal) {
			return h.services.EnableService(ctx, id)
		}
		return h.services.DisableService(ctx, id)
	}

	// No mutation available for the patched fields
	h.scope.Warn("no mutation for patched fields, dropping", "rowId", entry.RowID, "fields", fieldNames(entry.Data))
	return nil
}

// toBool converts a value from the CRUD entry data to a boolean.
// SQLite stores booleans as integers (0/1), and JSON decoding may
// produce float64 or bool depending on the source.
func toBool(v any) bool {
	switch val := v.(type) {
	case bool:
		return val
	case float64:
		return val != 0
	case int64:
		return val != 0
	case int:
		return val != 0
	default:
		return false
	}
}

// fieldNames returns the keys of a map for logging.
func fieldNames(data map[string]any) []string {
	names := make([]string, 0, len(data))
	for k := range data {
		names = append(names, k)
	}
	return names
}
