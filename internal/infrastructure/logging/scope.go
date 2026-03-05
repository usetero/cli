package logging

// Scope wraps a logger with a hierarchical scope path.
type Scope struct {
	logger Logger
	path   string
}

// RootScope creates the root scope.
func RootScope(logger Logger) Scope {
	return Scope{logger: logger}
}

// Child returns a new nested scope.
func (s Scope) Child(name string) Scope {
	path := name
	if s.path != "" {
		path = s.path + "/" + name
	}
	return Scope{logger: s.logger.With("scope", path), path: path}
}

// Path returns the scope path.
func (s Scope) Path() string {
	return s.path
}

// With adds context fields to the logger.
func (s Scope) With(args ...any) Scope {
	return Scope{logger: s.logger.With(args...), path: s.path}
}

func (s Scope) Debug(msg string, args ...any) { s.logger.Debug(msg, args...) }
func (s Scope) Info(msg string, args ...any)  { s.logger.Info(msg, args...) }
func (s Scope) Warn(msg string, args ...any)  { s.logger.Warn(msg, args...) }
func (s Scope) Error(msg string, args ...any) { s.logger.Error(msg, args...) }
