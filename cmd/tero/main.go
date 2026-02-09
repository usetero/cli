package main

import (
	"github.com/usetero/cli/internal/cmd"
)

// version is set via ldflags during build
var version = "dev"

func main() {
	cmd.Execute(version)
}
