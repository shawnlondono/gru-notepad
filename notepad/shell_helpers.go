package notepad

import "github.com/ledocorp/gru/ui"

// Bottom band that reveals auto-hidden mobile chrome (matches examples.shellFooterLift).
const shellFooterLift = float32(96)

func focusEditableAt(d *ui.Document) {
	ui.RouteScenePointerFocus(d)
}
