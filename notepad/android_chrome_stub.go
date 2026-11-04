//go:build !android

package notepad

import (
	"github.com/ledocorp/gru/examples"
	"github.com/ledocorp/gru/ui"
)

// SyncHostChrome is a no-op on desktop hosts.
func SyncHostChrome(scene examples.Scene, doc *ui.Document) {}
