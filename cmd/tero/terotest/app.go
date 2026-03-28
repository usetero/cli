package terotest

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/creack/pty"
)

const (
	defaultStopTimeout = 2 * time.Second
	defaultWaitStep    = 25 * time.Millisecond
	defaultWidth       = 120
	defaultHeight      = 40
)

type Key string

const (
	KeyEnter Key = "\r"
	KeyTab   Key = "\t"
	KeyUp    Key = "\x1b[A"
	KeyDown  Key = "\x1b[B"
	KeyCtrlC Key = "\x03"
)

type App struct {
	tb          testing.TB
	cmd         *exec.Cmd
	pty         *os.File
	output      *outputBuffer
	done        chan error
	stopTimeout time.Duration
}

type outputBuffer struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (b *outputBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.b.Write(p)
}

func (b *outputBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.b.String()
}

func Start(t testing.TB, binary string, opts Options) *App {
	t.Helper()

	if opts.HomeDir == "" {
		opts.HomeDir = tempHomeDir(t)
	}
	if opts.StopTimeout <= 0 {
		opts.StopTimeout = defaultStopTimeout
	}
	if opts.WindowWidth == 0 {
		opts.WindowWidth = defaultWidth
	}
	if opts.WindowHeight == 0 {
		opts.WindowHeight = defaultHeight
	}

	cmd := exec.CommandContext(context.Background(), binary, opts.Args...)
	cmd.Dir = cmdDir(t)
	cmd.Env = buildEnv(opts)

	ptmx, err := pty.StartWithSize(cmd, &pty.Winsize{
		Cols: opts.WindowWidth,
		Rows: opts.WindowHeight,
	})
	if err != nil {
		t.Fatalf("start tero app: %v", err)
	}

	out := &outputBuffer{}
	done := make(chan error, 1)
	go func() {
		_, _ = io.Copy(out, ptmx)
	}()
	go func() {
		done <- cmd.Wait()
	}()

	return &App{
		tb:          t,
		cmd:         cmd,
		pty:         ptmx,
		output:      out,
		done:        done,
		stopTimeout: opts.StopTimeout,
	}
}

func (a *App) WaitFor(text string, timeout time.Duration) {
	a.tb.Helper()

	if matched, ok := a.waitFor(timeout, text); ok {
		_ = matched
		return
	}
	a.tb.Fatalf("timed out waiting for %q\n%s", text, a.Snapshot())
}

func (a *App) WaitForAny(timeout time.Duration, texts ...string) string {
	a.tb.Helper()

	if len(texts) == 0 {
		a.tb.Fatal("WaitForAny requires at least one text")
	}
	matched, ok := a.waitFor(timeout, texts...)
	if !ok {
		a.tb.Fatalf("timed out waiting for one of %q\n%s", texts, a.Snapshot())
	}
	return matched
}

func (a *App) waitFor(timeout time.Duration, texts ...string) (string, bool) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		snapshot := a.Snapshot()
		for _, text := range texts {
			if strings.Contains(snapshot, text) {
				return text, true
			}
		}

		select {
		case err := <-a.done:
			a.tb.Fatalf("tero exited before rendering %q: %v\n%s", texts, err, snapshot)
		default:
		}

		time.Sleep(defaultWaitStep)
	}
	return "", false
}

func (a *App) Type(text string) {
	a.tb.Helper()

	if _, err := io.WriteString(a.pty, text); err != nil {
		a.tb.Fatalf("write to tero app: %v", err)
	}
}

func (a *App) Send(data []byte) {
	a.tb.Helper()

	if _, err := a.pty.Write(data); err != nil {
		a.tb.Fatalf("send to tero app: %v", err)
	}
}

func (a *App) Press(key Key) {
	a.tb.Helper()
	a.Send([]byte(key))
}

func (a *App) Snapshot() string {
	return ansi.Strip(a.output.String())
}

func (a *App) Stop() {
	a.tb.Helper()

	if a.cmd.Process == nil {
		return
	}

	select {
	case <-a.done:
		_ = a.pty.Close()
		return
	default:
	}

	a.Press(KeyCtrlC)

	select {
	case <-time.After(a.stopTimeout):
		_ = a.cmd.Process.Signal(syscall.SIGKILL)
		<-a.done
	case <-a.done:
	}

	_ = a.pty.Close()
}

func buildEnv(opts Options) []string {
	envMap := map[string]string{}
	for _, entry := range os.Environ() {
		key, value, ok := strings.Cut(entry, "=")
		if !ok {
			continue
		}
		envMap[key] = value
	}

	envMap["HOME"] = opts.HomeDir
	envMap["TERM"] = "xterm-256color"
	envMap["COLORTERM"] = "truecolor"

	for key, value := range opts.Env {
		envMap[key] = value
	}

	env := make([]string, 0, len(envMap))
	for key, value := range envMap {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}
	return env
}

func tempHomeDir(t testing.TB) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "tero-home-*")
	if err != nil {
		t.Fatalf("create temp home dir: %v", err)
	}
	return dir
}
