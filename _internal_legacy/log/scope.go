package log

// Scope wraps a Logger with hierarchical path tracking.
// Pass scopes down through component constructors to build a hierarchy of nested components.
// Unlike OpenTelemetry spans, scopes are long-lived (component lifetime) not per-operation.
type Scope struct {
	logger Logger
	path   string
}

// RootScope creates a root scope from a logger. It has no name until Child is called.
func RootScope(logger Logger) Scope {
	return Scope{
		logger: logger,
		path:   "",
	}
}

// Child creates a child scope with the given name appended to the path.
func (s Scope) Child(name string) Scope {
	var newPath string
	if s.path == "" {
		newPath = name
	} else {
		newPath = s.path + "/" + name
	}
	return Scope{
		logger: s.logger.With("scope", newPath),
		path:   newPath,
	}
}

// Path returns the full scope path.
func (s Scope) Path() string {
	return s.path
}

// Debug logs at debug level.
func (s Scope) Debug(msg string, args ...any) {
	s.logger.Debug(msg, args...)
}

// Info logs at info level.
func (s Scope) Info(msg string, args ...any) {
	s.logger.Info(msg, args...)
}

// Warn logs at warn level.
func (s Scope) Warn(msg string, args ...any) {
	s.logger.Warn(msg, args...)
}

// Error logs at error level.
func (s Scope) Error(msg string, args ...any) {
	s.logger.Error(msg, args...)
}

// With returns a new scope with additional context.
func (s Scope) With(args ...any) Scope {
	return Scope{
		logger: s.logger.With(args...),
		path:   s.path,
	}
}
