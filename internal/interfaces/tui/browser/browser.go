package browser

import (
	"fmt"
	"os/exec"
	"runtime"

	tea "charm.land/bubbletea/v2"
)

// OpenRequestedMsg asks the shell to open a URL in the user's browser.
type OpenRequestedMsg struct {
	URL string
}

// OpenCompletedMsg reports the browser open result back to the shell.
type OpenCompletedMsg struct {
	URL string
	Err error
}

// OpenURLCmd opens the provided URL using the local operating system.
func OpenURLCmd(url string) tea.Cmd {
	return func() tea.Msg {
		if url == "" {
			return OpenCompletedMsg{Err: fmt.Errorf("browser url is required")}
		}

		cmd, err := command(url)
		if err != nil {
			return OpenCompletedMsg{URL: url, Err: err}
		}
		if err := cmd.Run(); err != nil {
			return OpenCompletedMsg{URL: url, Err: err}
		}
		return OpenCompletedMsg{URL: url}
	}
}

func command(url string) (*exec.Cmd, error) {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", url), nil
	case "linux":
		return exec.Command("xdg-open", url), nil
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url), nil
	default:
		return nil, fmt.Errorf("unsupported platform %q", runtime.GOOS)
	}
}
