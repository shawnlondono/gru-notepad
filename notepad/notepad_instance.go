package notepad

import "github.com/ledocorp/gru/examples/appinstance"

var onInstanceOpenFile func(string)

func init() {
	appinstance.SetPendingOpenPath = notepadPendingOpenPath
}

// StartupInstance forwards duplicate launches to the running process (Windows/Linux).
// forwardPath is argv/GORY_OPEN_FILE for this launch; exit=true means this process should quit.
func StartupInstance(forwardPath string) (exit bool, stop func()) {
	return appinstance.Startup(forwardPath, func(path string) {
		if onInstanceOpenFile != nil {
			onInstanceOpenFile(path)
		}
	})
}

// SetInstanceOpenFileHandler receives paths from a second Gru Notepad launch.
func SetInstanceOpenFileHandler(fn func(string)) {
	onInstanceOpenFile = fn
}
