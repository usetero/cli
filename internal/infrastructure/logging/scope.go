package logging

// Scope wraps a logger with a hierarchical scope path.
type Scope struct {
	base  Logger
	attrs []any
	path  string
}

// RootScope creates the root scope.
func RootScope(logger Logger) Scope {
	return Scope{base: logger}
}

// Child returns a new nested scope.
func (s Scope) Child(name string) Scope {
	path := name
	if s.path != "" {
		path = s.path + "/" + name
	}
	return Scope{
		base:  s.base,
		attrs: append([]any(nil), s.attrs...),
		path:  path,
	}
}

// Path returns the scope path.
func (s Scope) Path() string {
	return s.path
}

// With adds context fields to the logger.
func (s Scope) With(args ...any) Scope {
	nextAttrs := append([]any(nil), s.attrs...)
	nextAttrs = append(nextAttrs, args...)
	return Scope{
		base:  s.base,
		attrs: nextAttrs,
		path:  s.path,
	}
}

func (s Scope) logger() Logger {
	l := s.base
	if l == nil {
		return NewWithWriter(nil, LevelInfo)
	}
	if len(s.attrs) > 0 {
		l = l.With(s.attrs...)
	}
	if s.path != "" {
		l = l.With("scope", s.path)
	}
	return l
}

func (s Scope) Debug(msg string, args ...any) { s.logger().Debug(msg, args...) }
func (s Scope) Info(msg string, args ...any)  { s.logger().Info(msg, args...) }
func (s Scope) Warn(msg string, args ...any)  { s.logger().Warn(msg, args...) }
func (s Scope) Error(msg string, args ...any) { s.logger().Error(msg, args...) }
