package datadog

import "github.com/pkg/browser"

// openBrowser is a package-level function for opening URLs in a browser.
// It can be replaced in tests to prevent actual browser windows from opening.
var openBrowser = browser.OpenURL
