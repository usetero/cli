package upload

import (
	"context"

	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/powersync/db"
)

// serviceHandler handles uploading service mutations to the GraphQL API.
// Only PATCH operations that change the "enabled" field are uploaded.
// All other operations are silently discarded to prevent stalling the queue.
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

	if entry.Op != db.OpPatch {
		h.scope.Debug("ignoring non-patch op", "op", entry.Op, "rowId", entry.RowID)
		return nil
	}

	enabledVal, ok := entry.Data["enabled"]
	if !ok {
		h.scope.Debug("ignoring patch without enabled field", "rowId", entry.RowID)
		return nil
	}

	enabled := toBool(enabledVal)
	if enabled {
		h.scope.Debug("enabling service", "rowId", entry.RowID)
		return h.services.EnableService(ctx, entry.RowID)
	}

	h.scope.Debug("disabling service", "rowId", entry.RowID)
	return h.services.DisableService(ctx, entry.RowID)
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
