package core

// NoticeLevel is the semantic tone for inline shell feedback.
type NoticeLevel uint8

const (
	NoticeInfo NoticeLevel = iota
	NoticeError
	NoticeSuccess
)

// Notice describes inline non-blocking feedback shown near active shell input.
type Notice struct {
	Level   NoticeLevel
	Message string
}

// NoticeProvider exposes the current inline notice for a model.
type NoticeProvider interface {
	Notice() *Notice
}
