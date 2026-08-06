// Package notepad is the Gru Notepad product scene.
//
// Thin host: go run ./cmd/grunotepad
package notepad

import "github.com/ledocorp/gru/examples"

// SceneTitle is the registered examples.Scene title.
// Kept equal to examples.NotepadSceneTitle.
const SceneTitle = examples.NotepadSceneTitle

// NewScene returns a fresh Notepad scene (preferred by the thin host).
func NewScene() examples.Scene {
	return &notepadScene{}
}

func init() {
	examples.Register(func() examples.Scene { return NewScene() })
}
