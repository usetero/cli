package version

// Version is the current CLI version.
// This should be updated manually before each release.
// In the future, this can be overridden at build time via ldflags:
//
//	go build -ldflags "-X github.com/usetero/cli/internal/version.Version=0.0.1"
const Version = "0.0.1"
