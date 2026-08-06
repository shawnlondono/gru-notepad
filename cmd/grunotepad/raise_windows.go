//go:build windows

package main

import (
	"syscall"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var (
	user32                  = syscall.NewLazyDLL("user32.dll")
	procSetForegroundWindow = user32.NewProc("SetForegroundWindow")
)

func raiseRunningWindow() {
	hwnd := uintptr(rl.GetWindowHandle())
	if hwnd == 0 {
		return
	}
	procSetForegroundWindow.Call(hwnd)
}
