package onboardingtest

import (
	"context"

	"github.com/usetero/cli/internal/domains/tenancy"
)

type AccountService struct {
	ListValue    []tenancy.Account
	CreatedValue tenancy.AccountID
}

func (s *AccountService) Create(context.Context, string) (tenancy.AccountID, error) {
	if s.CreatedValue == "" {
		s.CreatedValue = "acct_new"
	}
	return s.CreatedValue, nil
}
func (s *AccountService) Delete(context.Context, tenancy.AccountID) error { return nil }
func (s *AccountService) List(context.Context) ([]tenancy.Account, error) { return s.ListValue, nil }
