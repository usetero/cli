package main

import (
	"github.com/usetero/cli/internal/interfaces/cli"
)

// version is set via ldflags during build
var version = "dev"

func main() {
	cli.Execute(version)
}
