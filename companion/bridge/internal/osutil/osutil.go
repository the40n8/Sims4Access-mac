// Package osutil holds what differs between the Mac and Windows: opening
// the browser, child process windows, executable names, and the words and
// shortcuts the pages use.
package osutil

import (
	"os"
	"path/filepath"
	"runtime"
)

// Platform is "mac" or "windows" (the phone app's wording follows it).
func Platform() string {
	if runtime.GOOS == "windows" {
		return "windows"
	}
	return "mac"
}

// Computer is how the pages name this computer: "this Mac", "this PC".
func Computer() string {
	if runtime.GOOS == "windows" {
		return "PC"
	}
	return "Mac"
}

// MenuKeys is the shortcut for the mod's accessibility menu, as spoken.
func MenuKeys() string {
	if runtime.GOOS == "windows" {
		return "Control Alt H"
	}
	return "Option Command A"
}

// Exe adds ".exe" on Windows.
func Exe(name string) string {
	if runtime.GOOS == "windows" {
		return name + ".exe"
	}
	return name
}

// NextToMe is a file in the folder of the running program.
func NextToMe(name string) string {
	exe, _ := os.Executable()
	return filepath.Join(filepath.Dir(exe), name)
}

// Runnable reports whether path is a program this computer can start.
func Runnable(path string) bool {
	st, err := os.Stat(path)
	if err != nil || st.IsDir() {
		return false
	}
	if runtime.GOOS == "windows" {
		return filepath.Ext(path) == ".exe"
	}
	return st.Mode()&0o111 != 0
}

// DataDir is the companion's settings folder: ~/Library/Application
// Support/SimSightCompanion on the Mac, %AppData%\SimSightCompanion on Windows.
func DataDir() string {
	if dir, err := os.UserConfigDir(); err == nil {
		return filepath.Join(dir, "SimSightCompanion")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Library", "Application Support", "SimSightCompanion")
}
