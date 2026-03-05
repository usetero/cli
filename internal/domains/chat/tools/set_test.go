package tools

import (
	"context"
	"encoding/json"
	"path/filepath"
	"slices"
	"testing"

	"github.com/usetero/cli/internal/infrastructure/sqlite"
)

func TestToolset_DefinitionsAndRun(t *testing.T) {
	t.Parallel()

	sqlite.SetExtensionPath("")
	db, err := sqlite.OpenBare(context.Background(), filepath.Join(t.TempDir(), "tools.sqlite"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()
	if _, err := db.Exec(context.Background(), "CREATE TABLE things (id TEXT PRIMARY KEY, name TEXT NOT NULL)"); err != nil {
		t.Fatalf("create table: %v", err)
	}
	if _, err := db.Exec(context.Background(), "INSERT INTO things (id, name) VALUES ('thing_1', 'Thing')"); err != nil {
		t.Fatalf("insert thing: %v", err)
	}

	tools := Toolset{
		Query: NewQueryTool(db),
		Show:  NewShowTool(db),
		EnableService: NewEnableServiceTool(
			func(context.Context, ServiceID) error { return nil },
			func(context.Context, ServiceID) error { return nil },
		),
		ApprovePolicy: NewApprovePolicyTool(func(context.Context, PolicyID) error { return nil }),
	}
	defs := tools.Definitions()
	if len(defs) != 4 {
		t.Fatalf("expected 4 definitions, got %d", len(defs))
	}
	gotNames := []Name{defs[0].Name, defs[1].Name, defs[2].Name, defs[3].Name}
	wantNames := []Name{QueryToolName, ShowToolName, EnableServiceToolName, ApprovePolicyToolName}
	if !slices.Equal(gotNames, wantNames) {
		t.Fatalf("definitions names = %v, want %v", gotNames, wantNames)
	}

	out, err, ok := tools.Run(context.Background(), QueryToolName, json.RawMessage(`{"sql":"select 1 as n"}`))
	if err != nil || !ok {
		t.Fatalf("expected query tool run success, ok=%v err=%v", ok, err)
	}
	var parsed QueryResult
	if err := json.Unmarshal(out, &parsed); err != nil {
		t.Fatalf("parse output: %v", err)
	}

	_, err, ok = tools.Run(context.Background(), ShowToolName, json.RawMessage(`{"entity":"thing","id":"thing_1"}`))
	if err != nil || !ok {
		t.Fatalf("expected show tool run success, ok=%v err=%v", ok, err)
	}

	_, err, ok = tools.Run(context.Background(), EnableServiceToolName, json.RawMessage(`{"service_id":"svc_1","enabled":true}`))
	if err != nil || !ok {
		t.Fatalf("expected enable service tool run success, ok=%v err=%v", ok, err)
	}

	_, err, ok = tools.Run(context.Background(), ApprovePolicyToolName, json.RawMessage(`{"policy_id":"pol_1"}`))
	if err != nil || !ok {
		t.Fatalf("expected approve policy tool run success, ok=%v err=%v", ok, err)
	}

	_, _, ok = tools.Run(context.Background(), Name("unknown_tool"), json.RawMessage(`{}`))
	if ok {
		t.Fatalf("expected unknown tool to be missing")
	}
}
