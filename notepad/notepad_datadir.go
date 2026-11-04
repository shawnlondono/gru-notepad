//go:build !android

package notepad

import (
	"os"
	"path/filepath"
)

// notepadDataDir is where Gru Notepad stores notepad.db for this install.
//
// Release/portable layout (dict/en_US beside the exe): {exe}/data/
// so each extracted zip gets its own DB — not dev %AppData%.
//
// Dev runs without ship layout fall back to %AppData%/gru-notepad/.
func notepadDataDir() (string, error) {
	if portable := notepadPortableDataDir(); portable != "" {
		if err := os.MkdirAll(portable, 0o755); err != nil {
			return "", err
		}
		return portable, nil
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	appDir := filepath.Join(dir, "gru-notepad")
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		return "", err
	}
	return appDir, nil
}

// notepadPortableDataDir returns {exe}/data when the ship zip layout is detected.
func notepadPortableDataDir() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	return portableDataDirForRoot(filepath.Dir(exe))
}

func portableDataDirForRoot(exeDir string) string {
	aff := filepath.Join(exeDir, "dict", "en_US", "en_US.aff")
	if st, err := os.Stat(aff); err != nil || st.IsDir() {
		return ""
	}
	return filepath.Join(exeDir, "data")
}

func notepadPendingOpenPath() string {
	if portable := notepadPortableDataDir(); portable != "" {
		return filepath.Join(portable, "pending_open.txt")
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "gru-notepad", "pending_open.txt")
}
