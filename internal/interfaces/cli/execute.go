package cli

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/alecthomas/kong"
	"github.com/usetero/cli/internal/infrastructure/logging"
	"github.com/usetero/cli/internal/interfaces/cli/config"
)

// CLI is the Kong command model for tero.
type CLI struct {
	Config config.RuntimeConfig `embed:""`

	MCP MCP `cmd:"" help:"Run MCP mode."`
}

type runner struct {
	cfg   config.RuntimeConfig
	scope logging.Scope
}

// Execute runs the CLI entrypoint.
func Execute(version string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "Fatal error: %v\n%s", r, string(debug.Stack()))
			os.Exit(1)
		}
	}()

	cli := &CLI{}
	kctx, err := kong.New(cli,
		kong.Name("tero"),
		kong.Description("Tero CLI"),
		kong.Vars{"version": version},
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	parseCtx, err := kctx.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	cfg, err := config.Resolve(cli.Config)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	scope := logging.RootScope(logging.New(string(cfg.Env), cfg.Logging.Level)).Child("cli")
	parseCtx.Bind(&runner{cfg: cfg, scope: scope})
	if err := parseCtx.Run(); err != nil {
		scope.Error("command failed", "error", err)
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
