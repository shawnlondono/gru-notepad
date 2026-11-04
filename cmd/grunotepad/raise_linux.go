//go:build linux

package main

import (
	"github.com/ledocorp/gru/x11util"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// raiseRunningWindow matches linux-experience platformWindowFocus: GLFW's
// GetWindowHandle is a GLFWwindow*, so raise by title over a side X connection.
func raiseRunningWindow() {
	if !rl.IsWindowReady() {
		return
	}
	_ = x11util.RaiseNotepadWindow()
}
