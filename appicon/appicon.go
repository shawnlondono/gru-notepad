// Package appicon applies OS/title-bar icons for the public Gru Notepad host.
//
// The private monorepo uses github.com/ledocorp/gru/internal/appicon (not importable
// from another module). This package is the public-boundary stand-in — stubs today;
// embed Notepad PNGs here before a polished release if taskbar/title icons matter.
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
