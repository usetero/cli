package upload

import (
	"context"
	"fmt"

	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/powersync/db"
)

// policyHandler handles uploading policy changes to the GraphQL API.
type policyHandler struct {
	policies api.LogEventPolicies
	scope    log.Scope
}

func newPolicyHandler(policies api.LogEventPolicies, scope log.Scope) *policyHandler {
	return &policyHandler{
		policies: policies,
		scope:    scope.Child("policies"),
	}
}

func (h *policyHandler) Handle(ctx context.Context, entry *db.CrudEntry, emit Emitter) error {
	_ = emit // Policy handler doesn't emit events currently

	switch entry.Op {
	case db.OpPatch:
		return h.handlePatch(ctx, entry)
	default:
		// We only handle PATCH (updates) for policies - they're created server-side
		h.scope.Debug("ignoring policy op", "op", entry.Op, "id", entry.RowID)
		return nil
	}
}

func (h *policyHandler) handlePatch(ctx context.Context, entry *db.CrudEntry) error {
	// Check if this is an approval (approved_at is set)
	if approvedAt, ok := entry.Data["approved_at"]; ok && approvedAt != nil {
		err := h.policies.Approve(ctx, entry.RowID)
		if err != nil {
			return fmt.Errorf("approve policy: %w", err)
		}
		h.scope.Debug("approved policy", "id", entry.RowID)
		return nil
	}

	// If neither, just log and skip
	h.scope.Debug("policy patch with no approve", "id", entry.RowID, "data", entry.Data)
	return nil
}
