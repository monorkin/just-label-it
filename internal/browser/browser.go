package browser

import (
	"fmt"
	"os/exec"
	"runtime"
)

// Open opens a URL in the user's default browser.
func Open(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "linux":
		cmd = "xdg-open"
		args = []string{url}
	case "darwin":
		cmd = "open"
		args = []string{url}
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start", url}
	default:
		return fmt.Errorf("unsupported platform %q", runtime.GOOS)
	}

	return exec.Command(cmd, args...).Start()
}
