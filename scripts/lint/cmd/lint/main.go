package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type lintStep struct {
	name string
	cmd  string
	args []string
}

func main() {
	suite := flag.String("suite", "all", "lint suite: all|repo|analyzers|scripts")
	flag.Parse()

	steps, err := buildSteps(*suite)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	for _, step := range steps {
		if err := runStep(step); err != nil {
			os.Exit(1)
		}
	}
}

func buildSteps(suite string) ([]lintStep, error) {
	concurrency := fmt.Sprintf("--concurrency=%d", runtime.NumCPU())

	analyzers := []lintStep{
		{
			name: "go analyzers",
			cmd:  "go",
			args: []string{"run", "./scripts/lint/analyzers", "./..."},
		},
	}

	scriptChecks := []lintStep{
		{
			name: "shellcheck",
			cmd:  "sh",
			args: []string{"-c", "find scripts -type f -name '*.sh' -print0 | xargs -0 shellcheck"},
		},
		{
			name: "event naming",
			cmd:  "./scripts/lint/check-event-naming.sh",
		},
		{
			name: "local message naming",
			cmd:  "./scripts/lint/check-local-msg-naming.sh",
		},
		{
			name: "event ownership",
			cmd:  "./scripts/lint/check-event-ownership.sh",
		},
		{
			name: "event loop safety",
			cmd:  "./scripts/lint/check-event-loop-safety.sh",
		},
		{
			name: "tui child routing",
			cmd:  "./scripts/lint/check-tui-child-routing.sh",
		},
	}

	all := []lintStep{
		{
			name: "golangci-lint",
			cmd:  "golangci-lint",
			args: []string{"run", concurrency, "./..."},
		},
	}
	all = append(all, analyzers...)
	all = append(all, scriptChecks...)

	switch strings.ToLower(strings.TrimSpace(suite)) {
	case "all":
		return all, nil
	case "repo":
		repo := append([]lintStep{}, analyzers...)
		repo = append(repo, scriptChecks...)
		return repo, nil
	case "analyzers":
		return analyzers, nil
	case "scripts":
		return scriptChecks, nil
	default:
		return nil, errors.New("invalid --suite value: expected all|repo|analyzers|scripts")
	}
}

func runStep(step lintStep) error {
	fmt.Printf("lint: %s\n", step.name)
	command := exec.CommandContext(context.Background(), step.cmd, step.args...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Stdin = os.Stdin
	if err := command.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "lint step failed: %s\n", step.name)
		return err
	}
	return nil
}
