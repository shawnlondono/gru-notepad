// Package notepad is the Gru Notepad product scene (MASTER_SEPARATION_PLAN Phase 2.4).
//
// Thin host: go run -tags notepad ./cmd/grunotepad
// Studio dogfood: blank-import this package; GORY_SCENE="Notepad (Go)"
package notepad

import "github.com/ledocorp/gru/examples"

// SceneTitle is the registered examples.Scene title and GORY_SCENE value.
// Kept equal to examples.NotepadSceneTitle (shared string for non-notepad builds).
const SceneTitle = examples.NotepadSceneTitle

// NewScene returns a fresh Notepad scene (preferred by the thin host).
func NewScene() examples.Scene {
	return &notepadScene{}
}

func init() {
	examples.Register(func() examples.Scene { return NewScene() })
}
