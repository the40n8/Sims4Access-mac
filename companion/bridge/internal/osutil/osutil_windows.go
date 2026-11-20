package osutil

import (
	"os"
	"os/exec"
	"syscall"
)

const createNoWindow = 0x08000000

// OpenURL opens a page in the default browser.
func OpenURL(url string) error {
	cmd := exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	HideWindow(cmd)
	return cmd.Start()
}

// HideWindow keeps a console program (cloudflared) from opening a window
// over the game.
func HideWindow(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: createNoWindow}
}

// ComputerName is this PC's name.
func ComputerName() string {
	if n, err := os.Hostname(); err == nil && n != "" {
		return n
	}
	return "PC"
}
