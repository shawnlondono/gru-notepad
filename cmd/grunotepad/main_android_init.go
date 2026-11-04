//go:build android

package main

import rl "github.com/gen2brain/raylib-go/raylib"

// NativeActivity entry (raylib android_main → callback).
func init() {
	rl.SetCallbackFunc(runApp)
}
