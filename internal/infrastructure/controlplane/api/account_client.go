package api

import (
	"context"
	"fmt"

	"github.com/usetero/cli/internal/infrastructure/controlplane/api/gen"
)

// ListWorkspaces fetches workspaces for the bound account.
func (c *AccountClient) ListWorkspaces(ctx context.Context) ([]Workspace, error) {
	client, err := c.base.gqlForAccount(ctx, c.accountID)
	if err != nil {
		return nil, err
	}

	resp, err := gen.ListWorkspaces(ctx, client, c.accountID.String())
	if err != nil {
		return nil, err
	}

	edges := resp.Workspaces.Edges
	out := make([]Workspace, 0, len(edges))
	for _, edge := range edges {
		if edge == nil || edge.Node == nil {
			continue
		}
		node := edge.Node
		out = append(out, Workspace{
			ID:        WorkspaceID(node.Id),
			Name:      node.Name,
			CreatedAt: node.CreatedAt,
		})
	}
	return out, nil
}

// DeleteWorkspace deletes a workspace by ID.
func (c *AccountClient) DeleteWorkspace(ctx context.Context, workspaceID WorkspaceID) error {
	if workspaceID == "" {
		return fmt.Errorf("workspace id is required")
	}

	client, err := c.base.gqlForAccount(ctx, c.accountID)
	if err != nil {
		return err
	}

	resp, err := gen.DeleteWorkspace(ctx, client, workspaceID.String(), "DELETE")
	if err != nil {
		return err
	}
	if !resp.DeleteWorkspace {
		return fmt.Errorf("delete workspace returned false")
	}
	return nil
}

// GetDatadogAccount fetches Datadog integration metadata for the bound account.
func (c *AccountClient) GetDatadogAccount(ctx context.Context) (*DatadogAccount, error) {
	client, err := c.base.gqlForAccount(ctx, c.accountID)
	if err != nil {
		return nil, err
	}

	resp, err := gen.GetAccount(ctx, client, c.accountID.String())
	if err != nil {
		return nil, err
	}

	for _, edge := range resp.Accounts.Edges {
		if edge == nil || edge.Node == nil {
			continue
		}
		if edge.Node.DatadogAccount == nil {
			return nil, nil
		}
		dd := edge.Node.DatadogAccount
		return &DatadogAccount{ID: DatadogAccountID(dd.Id), Name: dd.Name, Site: DatadogSite(dd.Site)}, nil
	}

	return nil, nil
}

// ValidateDatadogAPIKey asks control plane to validate a Datadog API key.
func (c *AccountClient) ValidateDatadogAPIKey(ctx context.Context, apiKey string, site DatadogSite) (bool, string, error) {
	if apiKey == "" {
		return false, "", fmt.Errorf("api key is required")
	}
	if !site.Valid() {
		return false, "", fmt.Errorf("datadog site is required")
	}

	client, err := c.base.gqlForAccount(ctx, c.accountID)
	if err != nil {
		return false, "", err
	}

	resp, err := gen.ValidateDatadogApiKey(ctx, client, gen.ValidateDatadogApiKeyInput{
		ApiKey: apiKey,
		Site:   gen.DatadogAccountSite(site),
	})
	if err != nil {
		return false, "", err
	}

	out := resp.ValidateDatadogApiKey
	if out.Valid {
		return true, "", nil
	}

	message := "invalid api key"
	if out.Error != nil && *out.Error != "" {
		message = *out.Error
	}
	return false, message, nil
}

// CreateDatadogAccountWithCredentials creates Datadog integration with credentials.
func (c *AccountClient) CreateDatadogAccountWithCredentials(ctx context.Context, input CreateDatadogAccountInput) (DatadogAccount, error) {
	if input.Name == "" {
		return DatadogAccount{}, fmt.Errorf("name is required")
	}
	if !input.Site.Valid() {
		return DatadogAccount{}, fmt.Errorf("datadog site is required")
	}
	if input.APIKey == "" {
		return DatadogAccount{}, fmt.Errorf("api key is required")
	}
	if input.AppKey == "" {
		return DatadogAccount{}, fmt.Errorf("app key is required")
	}

	client, err := c.base.gqlForAccount(ctx, c.accountID)
	if err != nil {
		return DatadogAccount{}, err
	}

	resp, err := gen.CreateDatadogAccountWithCredentials(ctx, client, gen.CreateDatadogAccountWithCredentialsInput{
		Attributes: gen.CreateDatadogAccountInput{
			AccountID: c.accountID.String(),
			Name:      input.Name,
			Site:      gen.DatadogAccountSite(input.Site),
		},
		Credentials: gen.CreateDatadogCredentialsInput{
			ApiKey: input.APIKey,
			AppKey: input.AppKey,
		},
	})
	if err != nil {
		return DatadogAccount{}, err
	}

	created := resp.CreateDatadogAccount
	return DatadogAccount{ID: DatadogAccountID(created.Id), Name: created.Name, Site: DatadogSite(created.Site)}, nil
}

// GetDatadogAccountStatus returns status cache for Datadog account, or nil when unavailable.
func (c *AccountClient) GetDatadogAccountStatus(ctx context.Context, datadogAccountID DatadogAccountID) (*DatadogAccountStatus, error) {
	if datadogAccountID == "" {
		return nil, fmt.Errorf("datadog account id is required")
	}

	client, err := c.base.gqlForAccount(ctx, c.accountID)
	if err != nil {
		return nil, err
	}

	resp, err := gen.GetDatadogAccountStatus(ctx, client, datadogAccountID.String())
	if err != nil {
		return nil, err
	}

	for _, edge := range resp.DatadogAccounts.Edges {
		if edge == nil || edge.Node == nil || edge.Node.Status == nil {
			continue
		}
		s := edge.Node.Status
		return &DatadogAccountStatus{
			Health:                        DatadogAccountHealth(s.Health),
			ReadyForUse:                   s.ReadyForUse,
			ServiceCount:                  s.LogServiceCount,
			ActiveServices:                s.LogActiveServices,
			OKServices:                    s.OkServices,
			DisabledServices:              s.DisabledServices,
			InactiveServices:              s.InactiveServices,
			EventCount:                    s.LogEventCount,
			AnalyzedCount:                 s.LogEventAnalyzedCount,
			PreviewLogEventCount:          s.PreviewLogEventCount,
			EffectiveLogEventCount:        s.EffectiveLogEventCount,
			CurrentEventsPerHour:          s.CurrentEventsPerHour,
			CurrentBytesPerHour:           s.CurrentBytesPerHour,
			CurrentTotalUSDPerHour:        s.CurrentTotalUsdPerHour,
			PreviewSavedEventsPerHour:     s.PreviewSavedEventsPerHour,
			PreviewSavedBytesPerHour:      s.PreviewSavedBytesPerHour,
			PreviewSavedTotalUSDPerHour:   s.PreviewSavedTotalUsdPerHour,
			EffectiveSavedEventsPerHour:   s.EffectiveSavedEventsPerHour,
			EffectiveSavedBytesPerHour:    s.EffectiveSavedBytesPerHour,
			EffectiveSavedTotalUSDPerHour: s.EffectiveSavedTotalUsdPerHour,
			RefreshedAt:                   s.RefreshedAt,
		}, nil
	}

	return nil, nil
}
