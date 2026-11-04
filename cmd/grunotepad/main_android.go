//go:build android

// grunotepad Android NativeActivity host — fullscreen App Shell style (no title bar).
package main

import (
	"fmt"
	"os"
	"runtime"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/ledocorp/gru/examples"
	"github.com/ledocorp/gru/ui"
	"github.com/shawnlondono/gru-notepad/notepad"
)

func isAndroid() bool { return runtime.GOOS == "android" }

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

// main is required by the Go runtime; NativeActivity enters via SetCallbackFunc(runApp).
func main() {}

func runApp() {
	rl.SetTraceLogLevel(rl.LogWarning)
	rl.SetConfigFlags(uint32(rl.FlagVsyncHint))
	// 0×0 → full ANativeWindow (same as hello / grudemo).
	rl.InitWindow(0, 0, "Gru Notepad")
	defer rl.CloseWindow()
	rl.SetExitKey(0)

	ui.EnsureAndroidInsetFlags()
	windowW := int32(rl.GetScreenWidth())
	windowH := int32(rl.GetScreenHeight())
	if windowW > windowH {
		windowW, windowH = windowH, windowW
	}
	if windowW < 1 {
		windowW = 360
	}
	if windowH < 1 {
		windowH = 640
	}

	ui.ApplyUIOptimizations()
	ui.RefreshDisplayScale()
	ui.RefreshDensity()
	ui.RefreshTypeScaleFromWindow(windowW, windowH)
	ui.InitDisplayAwareAtlases()
	defer ui.UnloadSDFFonts()
	defer ui.Icons.UnloadAll()

	// SSAA off on Android (texture size + GLES).
	ui.SetTextEngineMode(ui.TextEngineShaped)

	ui.Toasts.SetWindowSize(windowW, windowH)
	ui.NotificationCenterMgr.SetWindowSize(windowW, windowH)
	ui.Tooltips.SetWindowSize(windowW, windowH)
	defer ui.Toasts.Unload()

	fmt.Fprintf(os.Stderr, "grunotepad android: %dx%d dpi=%.2f rem=%.0f backend=%s\n",
		windowW, windowH, ui.DisplayScale, ui.RootFontSize, ui.UIFontBackend())

	scene := notepad.NewScene()
	defer scene.Destroy()

	doc := ui.NewDocument(windowW, windowH)
	// Match grudemo Android mount: insets + Resize. Never MountFullscreenDocument —
	// that zeros ChromeTop and puts chrome under the status bar.
	ui.ApplyAndroidSafeInsets(doc, 0, windowW, windowH)
	scene.Build(doc)
	doc.Resize(windowW, windowH)
	examples.FinishShellMount(doc)
	ui.ActiveBreakpoint.Set(ui.CurrentBreakpoint(float32(windowW)))
	ui.SetOverlayChromeInsets(doc.ChromeTop(), doc.ChromeBottom())

	bg := rl.NewColor(228, 229, 235, 255)
	rl.SetTargetFPS(60)
	settleFrames := 8
	shouldClose := false
	diagFrames := 0

	fmt.Fprintln(os.Stderr, "grunotepad android: notepad scene mounted")

	for !rl.WindowShouldClose() && !shouldClose {
		if rl.IsKeyPressed(rl.KeyBack) {
			shouldClose = true
		}

		ui.PreparePointerInput()
		ui.SetActiveDocument(doc)
		dt := rl.GetFrameTime()
		doc.DrainQueue()

		nw, nh := int32(rl.GetScreenWidth()), int32(rl.GetScreenHeight())
		if nw > nh {
			nw, nh = nh, nw
		}
		if nw < 1 {
			nw = windowW
		}
		if nh < 1 {
			nh = windowH
		}
		if nw != windowW || nh != windowH {
			windowW, windowH = nw, nh
			ui.ApplyAndroidSafeInsets(doc, 0, windowW, windowH)
			doc.Resize(windowW, windowH)
			ui.Toasts.SetWindowSize(windowW, windowH)
			ui.Tooltips.SetWindowSize(windowW, windowH)
			ui.NotificationCenterMgr.SetWindowSize(windowW, windowH)
			ui.RefreshTypeScaleFromWindow(windowW, windowH)
			ui.ActiveBreakpoint.Set(ui.CurrentBreakpoint(float32(windowW)))
			ui.SetOverlayChromeInsets(doc.ChromeTop(), doc.ChromeBottom())
			settleFrames = 8
		}

		if settleFrames > 0 {
			if ui.ApplyAndroidSafeInsets(doc, 0, windowW, windowH) {
				doc.Resize(windowW, windowH)
				ui.SetOverlayChromeInsets(doc.ChromeTop(), doc.ChromeBottom())
			}
			doc.ForceFullLayout()
			examples.FinishShellMount(doc)
			settleFrames--
		}

		// Re-apply insets every frame: content-rect often arrives after settleFrames
		// (ChromeTop stuck at 0 → BottomNav hit Y missing chromeTop). Soft-IME
		// height (adjustNothing) is reported separately and lifts ChromeBottom.
		ui.EnsureAndroidInsetFlags()
		if ui.ApplyAndroidSafeInsets(doc, ui.IMEInsetHeight(), windowW, windowH) {
			doc.Resize(windowW, windowH)
			ui.SetOverlayChromeInsets(doc.ChromeTop(), doc.ChromeBottom())
		}
		doc.EnsureContentBand(windowW, windowH)

		ui.PrepareOverlayHitTest(doc.Root)
		ui.PrepareWheelScroll(doc.Root)
		if doc.Root.IsDirty() || ui.SubtreeLayoutDirty(doc.Root) {
			doc.Root.Layout()
		}
		notepad.SyncHostChrome(scene, doc)

		ui.ModalMgr.Update(dt)
		ui.DrawerMgr.Update(dt)
		ui.BottomSheetMgr.Update(dt)

		block := ui.OverlayBlocksSceneInput()
		ui.SetScenePointerBlocked(block)
		if !block {
			ui.ProcessSwitchListTilePointers(doc.Root, dt)
		}
		ui.BeginFrameCursor()
		doc.Root.Update(dt)
		ui.ContextMenuMgr.Update(dt)
		scene.OnUpdate(doc, dt)
		ui.SetScenePointerBlocked(false)

		ui.Tooltips.Update(dt)
		ui.Toasts.Update(dt)
		ui.NotificationCenterMgr.Update(dt)

		if doc.Root.IsDirty() || ui.SubtreeLayoutDirty(doc.Root) {
			doc.Root.Layout()
		}
		notepad.SyncHostChrome(scene, doc)
		if !ui.OverlayBlocksSceneInput() {
			ui.RouteScenePointerFocus(doc)
		}

		diagFrames++
		if ui.PointerClickPending() || rl.IsMouseButtonDown(rl.MouseLeftButton) {
			m := rl.GetMousePosition()
			if ui.PointerClickPending() {
				m = ui.PointerClickPosition()
			}
			rl.TraceLog(rl.LogWarning, "grunotepad tap: mouse=(%.0f,%.0f) down=%d pressed=%d pend=%d block=%d",
				m.X, m.Y,
				boolToInt(rl.IsMouseButtonDown(rl.MouseLeftButton)),
				boolToInt(rl.IsMouseButtonPressed(rl.MouseLeftButton)),
				boolToInt(ui.PointerClickPending()),
				boolToInt(block))
		}
		if diagFrames == 60 {
			rl.TraceLog(rl.LogWarning, "grunotepad chrome top=%.0f bottom=%.0f focus=%d",
				doc.ChromeTop(), doc.ChromeBottom(), boolToInt(ui.PlatformWindowFocused()))
		}

		rl.BeginDrawing()
		rl.ClearBackground(bg)
		doc.Root.Draw()
		ui.DrawAnimationOverlays(doc.Root)
		ui.DrawInteractionOverlays(doc.Root)
		ui.Tooltips.Draw()
		ui.ModalMgr.Draw()
		ui.DrawerMgr.Draw()
		ui.BottomSheetMgr.Draw()
		ui.ContextMenuMgr.Draw()
		ui.NotificationCenterMgr.Draw()
		ui.Toasts.Draw()
		ui.DrawOpenMenuPopups(doc.Root)
		rl.EndDrawing()
	}

	os.Exit(0)
}
