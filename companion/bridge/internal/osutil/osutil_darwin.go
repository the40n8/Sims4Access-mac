package osutil

import (
	"os/exec"
	"strings"
)

// OpenURL opens a page in the default browser.
func OpenURL(url string) error {
	return exec.Command("/usr/bin/open", url).Run()
}

// HideWindow does nothing on the Mac (programs have no console window).
func HideWindow(cmd *exec.Cmd) {}

// ComputerName is the name people gave this computer (System Settings).
func ComputerName() string {
	if out, err := exec.Command("/usr/sbin/scutil", "--get", "ComputerName").Output(); err == nil {
		if n := strings.TrimSpace(string(out)); n != "" {
			return n
		}
	}
	return "Mac"
}
