package bootstrap

import (
	"testing"

	"github.com/usetero/cli/internal/auth"
	"github.com/usetero/cli/internal/domain"
)

func TestCompleteOnboarding(t *testing.T) {
	t.Parallel()

	user := &auth.User{ID: "user-1"}
	org := &domain.Organization{ID: "org-1", Name: "Org 1"}
	account := &domain.Account{ID: "acc-1", Name: "Account 1"}

	got, ok := CompleteOnboarding(State{
		User:    user,
		Org:     org,
		Account: account,
	})
	if !ok {
		t.Fatal("expected completion payload")
	}
	if got.User == nil || got.User.ID != "user-1" {
		t.Fatalf("user = %#v, want user-1", got.User)
	}
	if got.Org.ID != "org-1" {
		t.Fatalf("org = %#v, want org-1", got.Org)
	}
	if got.Account.ID != "acc-1" {
		t.Fatalf("account = %#v, want acc-1", got.Account)
	}
}

func TestCompleteOnboardingMissingRequirements(t *testing.T) {
	t.Parallel()

	base := State{
		User:    &auth.User{ID: "user-1"},
		Org:     &domain.Organization{ID: "org-1"},
		Account: &domain.Account{ID: "acc-1"},
	}

	cases := []struct {
		name  string
		state State
	}{
		{name: "missing user", state: State{Org: base.Org, Account: base.Account}},
		{name: "missing org", state: State{User: base.User, Account: base.Account}},
		{name: "missing account", state: State{User: base.User, Org: base.Org}},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if _, ok := CompleteOnboarding(tt.state); ok {
				t.Fatalf("expected no completion payload for %s", tt.name)
			}
		})
	}
}
