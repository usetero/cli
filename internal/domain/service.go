package domain

// ServiceID is a unique identifier for a service.
type ServiceID string

func (id ServiceID) String() string { return string(id) }
