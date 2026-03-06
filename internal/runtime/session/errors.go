package session

import "fmt"

var ErrStorageNotScopeAware = fmt.Errorf("session storage must support organization scoping")
