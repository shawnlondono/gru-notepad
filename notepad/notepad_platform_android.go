//go:build android

package notepad

import (
	"github.com/ledocorp/gru/examples"
	"github.com/ledocorp/gru/ui"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func notepadWindowFocused() bool { return ui.PlatformWindowFocused() }

func notepadArmIME() { ui.ArmIMEForUserGesture() }

func notepadShowSoftInput() { ui.PlatformShowSoftInput() }

func notepadToolStripHeight() float32 { return ui.Dp(ui.PhoneToolStripHDp) }

func notepadBottomNavHeight() float32 { return ui.Dp(ui.PhoneBottomNavHDp) }

func notepadStartTextFileOpen(onDone func(path string, err error)) {
	examples.StartTextFileOpen(onDone)
}

func notepadStartTextFileSave(defaultName string, content []byte, onDone func(path string, err error)) {
	examples.StartTextFileSave(defaultName, content, onDone)
}

func notepadPinNodeBounds(n ui.Node, r rl.Rectangle) {
	if n == nil {
		return
	}
	n.PinScreenBounds(r)
}
