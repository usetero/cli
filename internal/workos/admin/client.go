// Package admin wraps the WorkOS management API for internal admin operations.
// The parent workos package handles OAuth (device flow, token refresh) using a public client ID.
// This package uses a secret API key for server-side management operations.
package admin

import (
	"context"
	"fmt"

	"github.com/workos/workos-go/v4/pkg/usermanagement"
)

// Client wraps the WorkOS management API, exposing only what we need.
type Client struct {
	um *usermanagement.Client
}

// NewClient creates a new admin client authenticated with a WorkOS API key.
func NewClient(apiKey string) *Client {
	return &Client{
		um: usermanagement.NewClient(apiKey),
	}
}

// Membership represents a user's membership in an organization.
type Membership struct {
	ID             string
	UserID         string
	OrganizationID string
	Status         string
}

// CreateMembership adds a user to an organization.
func (c *Client) CreateMembership(ctx context.Context, userID, orgID string) (Membership, error) {
	m, err := c.um.CreateOrganizationMembership(ctx, usermanagement.CreateOrganizationMembershipOpts{
		UserID:         userID,
		OrganizationID: orgID,
	})
	if err != nil {
		return Membership{}, fmt.Errorf("create membership: %w", err)
	}
	return toMembership(m), nil
}

// DeleteMembership removes a user from an organization.
func (c *Client) DeleteMembership(ctx context.Context, membershipID string) error {
	if err := c.um.DeleteOrganizationMembership(ctx, usermanagement.DeleteOrganizationMembershipOpts{
		OrganizationMembership: membershipID,
	}); err != nil {
		return fmt.Errorf("delete membership: %w", err)
	}
	return nil
}

// FindMembership finds a user's membership in an organization.
// Returns the membership if found, or an error if not found.
func (c *Client) FindMembership(ctx context.Context, userID, orgID string) (Membership, error) {
	resp, err := c.um.ListOrganizationMemberships(ctx, usermanagement.ListOrganizationMembershipsOpts{
		UserID:         userID,
		OrganizationID: orgID,
	})
	if err != nil {
		return Membership{}, fmt.Errorf("list memberships: %w", err)
	}
	if len(resp.Data) == 0 {
		return Membership{}, fmt.Errorf("no membership found for user %s in org %s", userID, orgID)
	}
	return toMembership(resp.Data[0]), nil
}

func toMembership(m usermanagement.OrganizationMembership) Membership {
	return Membership{
		ID:             m.ID,
		UserID:         m.UserID,
		OrganizationID: m.OrganizationID,
		Status:         string(m.Status),
	}
}
