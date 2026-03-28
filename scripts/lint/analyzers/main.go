package main

import (
	"os"

	"github.com/usetero/cli/scripts/lint/analyzers/archdeps"
	"github.com/usetero/cli/scripts/lint/analyzers/ctxfirst"
	"github.com/usetero/cli/scripts/lint/analyzers/teacmdpurity"
	"golang.org/x/tools/go/analysis/multichecker"
)

func main() {
	if len(os.Args) < 2 {
		os.Exit(0)
	}

	multichecker.Main(
		archdeps.Analyzer,
		ctxfirst.Analyzer,
		teacmdpurity.Analyzer,
	)
}
