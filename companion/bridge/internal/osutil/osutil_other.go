//go:build !darwin && !windows

package osutil

import (
	"os"
	"os/exec"
)

// OpenURL opens a page with xdg-open (development on Linux only).
func OpenURL(url string) error { return exec.Command("xdg-open", url).Start() }

func HideWindow(cmd *exec.Cmd) {}

func ComputerName() string {
	n, _ := os.Hostname()
	return n
}
