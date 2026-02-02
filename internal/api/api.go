package api

import (
	"errors"
	"strings"

	"github.com/usetero/cli/internal/log"
)

// Sentinel errors for API responses.
var (
	// ErrNotFound indicates the requested resource does not exist.
	ErrNotFound = errors.New("not found")

	// ErrAlreadyExists indicates the resource already exists.
	ErrAlreadyExists = errors.New("already exists")
)

// IsNotFound returns true if the error indicates a resource was not found.
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrNotFound) {
		return true
	}
	// Check error message for common "not found" patterns from GraphQL
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "not found") || strings.Contains(msg, "not_found")
}

// IsAlreadyExists returns true if the error indicates a resource already exists.
func IsAlreadyExists(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrAlreadyExists) {
		return true
	}
	// Check error message for common "already exists" patterns from GraphQL
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "already exists") || strings.Contains(msg, "already_exists") ||
		strings.Contains(msg, "duplicate") || strings.Contains(msg, "conflict")
}

// API bundles all control plane API services.
// It provides a single entry point for all API-related operations.
type API struct {
	Organizations   *OrganizationService
	Accounts        *AccountService
	DatadogAccounts *DatadogAccountService
	Services        *ServiceService
}

// New creates a new API with all services initialized.
// Requires an authenticated API client.
func New(client Client, logger log.Logger) *API {
	return &API{
		Organizations:   NewOrganizationService(client, logger),
		Accounts:        NewAccountService(client, logger),
		DatadogAccounts: NewDatadogAccountService(client, logger),
		Services:        NewServiceService(client, logger),
	}
}
