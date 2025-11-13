package organization

// DefaultOrgSaver defines the interface for saving and retrieving default organization preferences.
type DefaultOrgSaver interface {
	GetDefaultOrgID() string
	SetDefaultOrgID(orgID string) error
}
