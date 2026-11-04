//go:build android

package notepad

import (
	"os"
	"path/filepath"
)

// notepadDataDir stores notepad.db under the app files tree on NativeActivity.
// Package id must match packaging/android/notepad applicationId.
func notepadDataDir() (string, error) {
	base := "/data/data/org.ledocorp.gru.notepad/files"
	if home := os.Getenv("HOME"); home != "" {
		base = home
	}
	appDir := filepath.Join(base, "gru-notepad")
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		// Fallback: world-writable cache (emulator / restricted installs).
		appDir = filepath.Join(os.TempDir(), "gru-notepad")
		if err2 := os.MkdirAll(appDir, 0o755); err2 != nil {
			return "", err
		}
	}
	return appDir, nil
}

func notepadPendingOpenPath() string {
	dir, err := notepadDataDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "pending_open.txt")
}
