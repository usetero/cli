package extension

import "errors"

// ErrNoActiveIteration indicates control-plane received an event when no sync
// iteration is currently active.
var ErrNoActiveIteration = errors.New("powersync extension: no active iteration")
