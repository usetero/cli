package session

import (
	"context"
	"errors"
	"testing"

	"github.com/usetero/cli/internal/domains/tenancy"
	"github.com/usetero/cli/internal/infrastructure/logging/logtest"
	pssyncer "github.com/usetero/cli/internal/infrastructure/powersync/syncer"
	"github.com/usetero/cli/internal/infrastructure/sqlite"
	"github.com/usetero/cli/internal/runtime/session/sessiontest"
)

type scopedStorage struct {
	path sqlite.DatabasePath
	org  tenancy.OrganizationID
}

func (s *scopedStorage) SetOrganizationID(organizationID tenancy.OrganizationID) {
	s.org = organizationID
}

func (s *scopedStorage) DatabasePath(accountID sqlite.AccountID) (sqlite.DatabasePath, error) {
	if s.org == "" {
		return "", ErrStorageNotScopeAware
	}
	return s.path, accountID.Validate()
}

func TestService_Ensure_StartNoopAndSwitch(t *testing.T) {
	t.Parallel()

	storage := &scopedStorage{path: sqlite.DatabasePath(t.TempDir() + "/session.sqlite")}
	syncerOne := &sessiontest.Syncer{Ready: false, StateValue: &pssyncer.Connecting{}}
	syncerTwo := &sessiontest.Syncer{Ready: true, StateValue: &pssyncer.Ready{}}
	syncerThree := &sessiontest.Syncer{Ready: true, StateValue: &pssyncer.Ready{}}
	syncers := []*sessiontest.Syncer{syncerOne, syncerTwo, syncerThree}
	syncerIdx := 0

	svc := NewService(
		storage,
		func() (Syncer, error) {
			next := syncers[syncerIdx]
			syncerIdx++
			return next, nil
		},
		func(_ *sqlite.DB, _ interface{ NotifyUploadCompleted(context.Context) error }) (Uploader, error) {
			return sessiontest.NewUploader(), nil
		},
		logtest.NewScope(t),
	)
	svc.setOpenDB(openBareDB(t))

	scope1 := Scope{OrganizationID: "org_1", AccountID: "acc_1"}
	if err := svc.Ensure(context.Background(), scope1); err != nil {
		t.Fatalf("ensure scope1: %v", err)
	}
	if err := svc.Ensure(context.Background(), scope1); err != nil {
		t.Fatalf("ensure scope1 noop: %v", err)
	}
	if syncerIdx != 1 {
		t.Fatalf("expected one syncer start after noop ensure, got %d", syncerIdx)
	}

	scope2 := Scope{OrganizationID: "org_1", AccountID: "acc_2"}
	if err := svc.Ensure(context.Background(), scope2); err != nil {
		t.Fatalf("ensure scope2: %v", err)
	}
	if syncerIdx != 2 {
		t.Fatalf("expected scope switch to restart session, got starts=%d", syncerIdx)
	}

	scope3 := Scope{OrganizationID: "org_2", AccountID: "acc_2"}
	if err := svc.Ensure(context.Background(), scope3); err != nil {
		t.Fatalf("ensure scope3: %v", err)
	}
	if syncerIdx != 3 {
		t.Fatalf("expected organization switch to restart session, got starts=%d", syncerIdx)
	}

	status := svc.Status()
	if !status.Running {
		t.Fatalf("expected running status")
	}
	if status.Scope != scope3 {
		t.Fatalf("unexpected status scope: %+v", status.Scope)
	}
	if !status.Ready {
		t.Fatalf("expected ready status")
	}

	if err := svc.Stop(); err != nil {
		t.Fatalf("stop: %v", err)
	}
}

func TestService_Ensure_RequiresScopedStorage(t *testing.T) {
	t.Parallel()

	svc := NewService(
		sessiontest.Storage{Path: sqlite.DatabasePath(t.TempDir() + "/session.sqlite")},
		func() (Syncer, error) { return &sessiontest.Syncer{}, nil },
		func(_ *sqlite.DB, _ interface{ NotifyUploadCompleted(context.Context) error }) (Uploader, error) {
			return sessiontest.NewUploader(), nil
		},
		logtest.NewScope(t),
	)
	if err := svc.Ensure(context.Background(), Scope{OrganizationID: "org_1", AccountID: "acc_1"}); !errors.Is(err, ErrStorageNotScopeAware) {
		t.Fatalf("expected ErrStorageNotScopeAware, got %v", err)
	}
}
