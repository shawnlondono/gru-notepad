// Package appicon applies OS/title-bar icons for the Gru Notepad host.
// Assets may be embedded here; stubs are fine until PNGs are added.
package appicon

import rl "github.com/gen2brain/raylib-go/raylib"

// SetPreferDarkIcon selects light- vs dark-chrome icon sets when assets exist.
func SetPreferDarkIcon(dark bool) {}

// ApplyWindowIcon sets the OS window icon (no-op until assets are embedded).
func ApplyWindowIcon() {}

// ReapplyWindowIcon refreshes the OS window icon after show/theme changes.
func ReapplyWindowIcon() {}

// DrawTitleBarIcon paints the leading title-bar mark (no-op until assets are embedded).
func DrawTitleBarIcon(rect rl.Rectangle) {}
