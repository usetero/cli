package main

import (
	"github.com/usetero/cli/internal/cmd"
)

// version is set via ldflags during build
var version = "0.0.1" // TODO: revert to "dev"

func main() {
	cmd.Execute(version)
}
