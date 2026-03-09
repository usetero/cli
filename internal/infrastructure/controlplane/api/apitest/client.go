package apitest

import (
	"context"

	controlplane "github.com/usetero/cli/internal/infrastructure/controlplane/api"
)

// Client is a configurable test double for control-plane API calls.
type Client struct {
	ListOrganizationsFn                   func(ctx context.Context) ([]controlplane.Organization, error)
	CreateOrganizationAndBootstrapFn      func(ctx context.Context, name string) (controlplane.OrganizationBootstrap, error)
	DeleteOrganizationFn                  func(ctx context.Context, organizationID controlplane.OrganizationID) error
	ListAccountsFn                        func(ctx context.Context, organizationID controlplane.OrganizationID) ([]controlplane.Account, error)
	CreateAccountFn                       func(ctx context.Context, organizationID controlplane.OrganizationID, name string) (controlplane.Account, error)
	DeleteAccountFn                       func(ctx context.Context, accountID controlplane.AccountID) error
	ListWorkspacesFn                      func(ctx context.Context, accountID controlplane.AccountID) ([]controlplane.Workspace, error)
	DeleteWorkspaceFn                     func(ctx context.Context, workspaceID controlplane.WorkspaceID) error
	GetAccountDatadogAccountFn            func(ctx context.Context, accountID controlplane.AccountID) (*controlplane.DatadogAccount, error)
	ValidateDatadogAPIKeyFn               func(ctx context.Context, apiKey string, site controlplane.DatadogSite) (bool, string, error)
	CreateDatadogAccountWithCredentialsFn func(ctx context.Context, input controlplane.CreateDatadogAccountInput) (controlplane.DatadogAccount, error)
	GetDatadogAccountStatusFn             func(ctx context.Context, datadogAccountID controlplane.DatadogAccountID) (*controlplane.DatadogAccountStatus, error)
}

var _ interface {
	ListOrganizations(ctx context.Context) ([]controlplane.Organization, error)
	CreateOrganizationAndBootstrap(ctx context.Context, name string) (controlplane.OrganizationBootstrap, error)
	DeleteOrganization(ctx context.Context, organizationID controlplane.OrganizationID) error
	ListAccounts(ctx context.Context, organizationID controlplane.OrganizationID) ([]controlplane.Account, error)
	CreateAccount(ctx context.Context, organizationID controlplane.OrganizationID, name string) (controlplane.Account, error)
	DeleteAccount(ctx context.Context, accountID controlplane.AccountID) error
	ListWorkspaces(ctx context.Context, accountID controlplane.AccountID) ([]controlplane.Workspace, error)
	DeleteWorkspace(ctx context.Context, workspaceID controlplane.WorkspaceID) error
	GetAccountDatadogAccount(ctx context.Context, accountID controlplane.AccountID) (*controlplane.DatadogAccount, error)
	ValidateDatadogAPIKey(ctx context.Context, apiKey string, site controlplane.DatadogSite) (bool, string, error)
	CreateDatadogAccountWithCredentials(ctx context.Context, input controlplane.CreateDatadogAccountInput) (controlplane.DatadogAccount, error)
	GetDatadogAccountStatus(ctx context.Context, datadogAccountID controlplane.DatadogAccountID) (*controlplane.DatadogAccountStatus, error)
} = (*Client)(nil)

func (c *Client) ListOrganizations(ctx context.Context) ([]controlplane.Organization, error) {
	if c.ListOrganizationsFn == nil {
		return nil, nil
	}
	return c.ListOrganizationsFn(ctx)
}

func (c *Client) CreateOrganizationAndBootstrap(ctx context.Context, name string) (controlplane.OrganizationBootstrap, error) {
	if c.CreateOrganizationAndBootstrapFn == nil {
		return controlplane.OrganizationBootstrap{}, nil
	}
	return c.CreateOrganizationAndBootstrapFn(ctx, name)
}

func (c *Client) DeleteOrganization(ctx context.Context, organizationID controlplane.OrganizationID) error {
	if c.DeleteOrganizationFn == nil {
		return nil
	}
	return c.DeleteOrganizationFn(ctx, organizationID)
}

func (c *Client) ListAccounts(ctx context.Context, organizationID controlplane.OrganizationID) ([]controlplane.Account, error) {
	if c.ListAccountsFn == nil {
		return nil, nil
	}
	return c.ListAccountsFn(ctx, organizationID)
}

func (c *Client) CreateAccount(ctx context.Context, organizationID controlplane.OrganizationID, name string) (controlplane.Account, error) {
	if c.CreateAccountFn == nil {
		return controlplane.Account{}, nil
	}
	return c.CreateAccountFn(ctx, organizationID, name)
}

func (c *Client) DeleteAccount(ctx context.Context, accountID controlplane.AccountID) error {
	if c.DeleteAccountFn == nil {
		return nil
	}
	return c.DeleteAccountFn(ctx, accountID)
}

func (c *Client) ListWorkspaces(ctx context.Context, accountID controlplane.AccountID) ([]controlplane.Workspace, error) {
	if c.ListWorkspacesFn == nil {
		return nil, nil
	}
	return c.ListWorkspacesFn(ctx, accountID)
}

func (c *Client) DeleteWorkspace(ctx context.Context, workspaceID controlplane.WorkspaceID) error {
	if c.DeleteWorkspaceFn == nil {
		return nil
	}
	return c.DeleteWorkspaceFn(ctx, workspaceID)
}

func (c *Client) GetAccountDatadogAccount(ctx context.Context, accountID controlplane.AccountID) (*controlplane.DatadogAccount, error) {
	if c.GetAccountDatadogAccountFn == nil {
		return nil, nil
	}
	return c.GetAccountDatadogAccountFn(ctx, accountID)
}

func (c *Client) ValidateDatadogAPIKey(ctx context.Context, apiKey string, site controlplane.DatadogSite) (bool, string, error) {
	if c.ValidateDatadogAPIKeyFn == nil {
		return false, "", nil
	}
	return c.ValidateDatadogAPIKeyFn(ctx, apiKey, site)
}

func (c *Client) CreateDatadogAccountWithCredentials(ctx context.Context, input controlplane.CreateDatadogAccountInput) (controlplane.DatadogAccount, error) {
	if c.CreateDatadogAccountWithCredentialsFn == nil {
		return controlplane.DatadogAccount{}, nil
	}
	return c.CreateDatadogAccountWithCredentialsFn(ctx, input)
}

func (c *Client) GetDatadogAccountStatus(ctx context.Context, datadogAccountID controlplane.DatadogAccountID) (*controlplane.DatadogAccountStatus, error) {
	if c.GetDatadogAccountStatusFn == nil {
		return nil, nil
	}
	return c.GetDatadogAccountStatusFn(ctx, datadogAccountID)
}
