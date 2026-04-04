package core

// Page describes shell-facing page metadata.
type Page struct {
	Title string
}

// PageProvider exposes shell-facing page metadata.
type PageProvider interface {
	Page() Page
}
