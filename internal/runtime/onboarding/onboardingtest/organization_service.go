package onboardingtest

import (
	"context"

	"github.com/usetero/cli/internal/domains/tenancy"
)

type OrganizationService struct {
	ListValue      []tenancy.Organization
	BootstrapValue tenancy.OrganizationBootstrap
}

func (s *OrganizationService) List(context.Context) ([]tenancy.Organization, error) {
	return s.ListValue, nil
}

func (s *OrganizationService) Create(context.Context, string) (tenancy.OrganizationBootstrap, error) {
	return s.BootstrapValue, nil
}
