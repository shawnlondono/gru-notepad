//go:build !android

package notepad

import (
	"github.com/ledocorp/gru/examples"
	"github.com/ledocorp/gru/ui"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func notepadWindowFocused() bool { return rl.IsWindowFocused() }

func notepadArmIME() {}

func notepadShowSoftInput() {}

func notepadToolStripHeight() float32 { return notepadMobileRibbonH }

func notepadBottomNavHeight() float32 { return 64 }

func notepadStartTextFileOpen(onDone func(path string, err error)) {
	path, err := examples.PickTextFileOpen()
	if onDone != nil {
		onDone(path, err)
	}
}

func notepadStartTextFileSave(defaultName string, content []byte, onDone func(path string, err error)) {
	_ = content
	path, err := examples.PickTextFileSave(defaultName)
	if onDone != nil {
		onDone(path, err)
	}
}

func notepadPinNodeBounds(n ui.Node, r rl.Rectangle) {
	if n == nil {
		return
	}
	n.SetBounds(r)
}
