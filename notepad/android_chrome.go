//go:build android

package notepad

import (
	"github.com/ledocorp/gru/examples"
	"github.com/ledocorp/gru/ui"
)

// SyncHostChrome pins BottomNav to the ChromeTop content band before hit-testing.
// Hosts call this after Layout and before Root.Update.
func SyncHostChrome(scene examples.Scene, doc *ui.Document) {
	s, ok := scene.(*notepadScene)
	if !ok || s == nil {
		return
	}
	s.SyncAndroidChromeBand(doc)
}
