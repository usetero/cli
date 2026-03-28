package browser

import (
	"fmt"
	"os/exec"
	"runtime"
)

// Open launches the user's default browser for the given URL.
func Open(url string) error {
	if url == "" {
		return fmt.Errorf("browser url is required")
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return cmd.Start()
}
