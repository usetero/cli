package upload

import (
	"context"

	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/powersync/db"
)

// policyHandler handles uploading log event policy mutations to the GraphQL API.
type policyHandler struct {
	policies api.Policies
	scope    log.Scope
}

func newPolicyHandler(policies api.Policies, scope log.Scope) *policyHandler {
	return &policyHandler{
		policies: policies,
		scope:    scope.Child("policies"),
	}
}

func (h *policyHandler) Handle(ctx context.Context, entry *db.CrudEntry, emit Emitter) error {
	_ = emit

	switch entry.Op {
	case db.OpPatch:
		return h.handlePatch(ctx, entry)
	default:
		h.scope.Warn("unsupported policy op, dropping", "op", entry.Op, "rowId", entry.RowID)
		return nil
	}
}

func (h *policyHandler) handlePatch(ctx context.Context, entry *db.CrudEntry) error {
	if _, ok := entry.Data["approved_at"]; ok {
		return h.policies.ApprovePolicy(ctx, entry.RowID)
	}
	if _, ok := entry.Data["dismissed_at"]; ok {
		return h.policies.DismissPolicy(ctx, entry.RowID)
	}

	h.scope.Warn("no mutation for patched fields, dropping", "rowId", entry.RowID, "fields", fieldNames(entry.Data))
	return nil
}
