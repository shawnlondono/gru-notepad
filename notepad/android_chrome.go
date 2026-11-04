//go:build android

package notepad

import (
	"github.com/ledocorp/gru/examples"
	"github.com/ledocorp/gru/ui"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// SyncHostChrome pins BottomNav to the ChromeTop content band before hit-testing.
// Hosts call this after Layout and before Root.Update.
func SyncHostChrome(scene examples.Scene, doc *ui.Document) {
	s, ok := scene.(*notepadScene)
	if !ok || s == nil {
		return
	}
	s.syncAndroidChromeBand(doc)
}

// syncAndroidChromeBand keeps the edge-to-edge shell and BottomNav on the
// ChromeTop content band. Android-only (uses Document.EnsureContentBand).
func (s *notepadScene) syncAndroidChromeBand(doc *ui.Document) {
	if !ui.AndroidGLES() || doc == nil || doc.Root == nil {
		return
	}
	if doc.Width > 0 && doc.Height > 0 {
		fullH := int32(doc.ChromeTop()) + doc.Height + int32(doc.ChromeBottom())
		if fullH < doc.Height {
			fullH = doc.Height
		}
		doc.EnsureContentBand(doc.Width, fullH)
	} else {
		band := rl.NewRectangle(0, doc.ChromeTop(), float32(doc.Width), float32(doc.Height))
		for _, ch := range doc.Root.Children() {
			b := ch.Bounds()
			if abs32(b.Y-band.Y) > 0.5 || abs32(b.Height-band.Height) > 0.5 ||
				abs32(b.Width-band.Width) > 0.5 {
				ch.SetBounds(band)
			}
		}
	}
	s.pinMobileBottomNavBounds(doc)
}
