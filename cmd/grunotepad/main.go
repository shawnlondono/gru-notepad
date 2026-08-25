// grunotepad is the thin window host for Gru Notepad.
package main

import (
	"fmt"
	"os"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/shawnlondono/gru-notepad/notepad"
	"github.com/ledocorp/gru/examples"
	"github.com/ledocorp/gru/examples/appinstance"
	"github.com/shawnlondono/gru-notepad/appicon"
	"github.com/ledocorp/gru/ui"
)

const (
	initW int32 = 1440
	initH int32 = 900
)

func main() {
	exit, stopInstance := notepad.StartupInstance(examples.StartupOpenFilePath())
	if exit {
		fmt.Fprintln(os.Stderr, "grunotepad: another Gru Notepad is already running (forwarded open, exiting)")
		return
	}
	defer stopInstance()

	rl.SetTraceLogLevel(rl.LogWarning)
	rl.SetConfigFlags(uint32(rl.FlagMsaa4xHint | rl.FlagVsyncHint |
		rl.FlagWindowResizable | rl.FlagWindowUndecorated | rl.FlagWindowHidden))
	rl.InitWindow(initW, initH, "Gru Notepad")
	appinstance.RaiseRunningInstance = raiseRunningWindow
	defer rl.CloseWindow()
	rl.SetExitKey(0)
	rl.SetWindowMinSize(int(ui.MinClientWidth), 320)

	ui.InitWebViewHostPolicy(true)
	ui.ApplyUIOptimizations()
	ui.RefreshDisplayScale()
	ui.InitDisplayAwareAtlases()
	ui.SetTextEngineMode(ui.TextEngineShaped)
	defer ui.UnloadSDFFonts()
	defer ui.Icons.UnloadAll()

	appicon.ApplyWindowIcon()

	windowW, windowH := initW, initH
	ui.InitSupersampling(windowW, windowH)
	ui.ApplyDisplayAwareSupersampling(windowW, windowH)
	defer ui.UnloadSupersampling()
	ui.RefreshTypeScaleFromWindow(windowW, windowH)
	ui.Toasts.SetWindowSize(windowW, windowH)
	ui.NotificationCenterMgr.SetWindowSize(windowW, windowH)
	ui.Tooltips.SetWindowSize(windowW, windowH)
	defer ui.Toasts.Unload()

	appicon.ReapplyWindowIcon()
	fmt.Fprintf(os.Stderr, "Gru Notepad render: text=%s icons=%s ssaa=%.1fx dpi=%.2f\n",
		ui.TextEngineBackendName(), ui.Phosphor.IconFontSummary(),
		ui.EffectiveSupersamplingScale(), ui.DisplayScale)

	scene := notepad.NewScene()
	defer scene.Destroy()
	defer ui.DestroyAllWebViewHosts()

	doc := ui.NewDocument(windowW, windowH)
	doc.SetChromeTop(ui.TitleBarHeight)
	scene.Build(doc)
	// Settle after Resize-equivalent mount so MarkdownView soft-reset sees real widths
	// (FinishShellMount → settlePreviewMarkdown). Do not rely on Build-time FinishShellMount alone.
	ui.MountBorderlessDocument(doc, windowW, windowH)
	examples.FinishShellMount(doc)
	if rl.IsWindowReady() {
		doc.SetPlatformWindowHandle(uintptr(rl.GetWindowHandle()))
	}
	ui.SyncWebViewHosts(doc)
	doc.EnableUIRenderTexture(true)
	ui.ActiveBreakpoint.Set(ui.CurrentBreakpoint(float32(windowW)))

	shouldClose := false
	titleBar := ui.NewTitleBar("Gru Notepad", ui.TitleBarStyleDark,
		func() { shouldClose = true },
		func() { rl.MinimizeWindow() },
		func() {
			if rl.IsWindowMaximized() {
				rl.RestoreWindow()
			} else {
				rl.MaximizeWindow()
			}
		})
	titleBar.HandleResize = true
	titleBar.DrawLeadingIcon = appicon.DrawTitleBarIcon
	titleBar.SetSize(windowW, windowH)
	examples.ConfigureTitleBar(titleBar, scene)
	rl.SetWindowTitle("Gru Notepad")

	bg := ui.WindowChromeFill()
	borderless := true
	windowShown := false
	rl.SetTargetFPS(60)
	ui.SetOverlayChromeInsets(ui.TitleBarHeight, 0)
	ui.ApplyNativeBorderlessRoundedCorners(true)

	ui.PresentHook = func() {
		if doc == nil || !rl.IsWindowReady() {
			return
		}
		rl.BeginDrawing()
		ui.BeginDrawingBorderless(ui.WindowChromeFill(), borderless, windowW, windowH)
		if ui.SuperTargetDrawable() {
			ui.BlitToScreenBorderless(windowW, windowH, borderless)
		}
		ui.DrawAnimationOverlays(doc.Root)
		drawScreenOverlays(windowW, windowH, doc.Root)
		rl.EndDrawing()
	}

	fmt.Fprintln(os.Stderr, "grunotepad: thin host (edge shell)")

	setBreakpointIfChanged := func(windowW int32) {
		if bp := ui.CurrentBreakpoint(float32(windowW)); ui.ActiveBreakpoint.Get() != bp {
			ui.ActiveBreakpoint.Set(bp)
		}
	}

	var resizeHold ui.ResizeHoldTracker
	var captionHoverMask int
	for !rl.WindowShouldClose() && !shouldClose {
		frameStart := time.Now()
		ui.PreparePointerInput()
		ui.SetActiveDocument(doc)
		dt := rl.GetFrameTime()
		doc.DrainQueue()
		bg = ui.WindowChromeFill()

		nw, nh := int32(rl.GetScreenWidth()), int32(rl.GetScreenHeight())
		if nw < 1 {
			nw = windowW
		}
		if nh < 1 {
			nh = windowH
		}
		frameResized := nw != windowW || nh != windowH
		// Live dimension change: mark moving before ANY Resize→Root.Layout this frame.
		if frameResized || titleBar.IsResizing() || ui.FillClientChromeResizing() {
			ui.SetChromeWindowMoving(true)
		}
		if frameResized {
			windowW, windowH = nw, nh
			resizeHold.NoteDimensionChange(frameStart)
			// Defer RT + shell layout (S1: ForceFullLayout only if rem changes — inside Sync).
			ui.SyncBorderlessClientSize(doc, titleBar, windowW, windowH)
			setBreakpointIfChanged(windowW)
		}

		// Frame order: overlay hit rects → title bar → wheel →
		// layout-before-pointer → overlay mgrs → blocked scene input → Update.
		ui.PrepareOverlayHitTest(doc.Root)
		titleBar.Update(windowW, windowH)
		prevCaptionMask := captionHoverMask
		captionHoverMask = titleBar.CaptionHoverMask()
		// Post-blit paints hover on cache hits; one invalidate on hover exit clears stale plates.
		if prevCaptionMask != 0 && captionHoverMask == 0 &&
			!titleBar.IsResizing() && !titleBar.IsDragging() &&
			!ui.PresentSizeMismatch(windowW, windowH) {
			doc.InvalidatePaint()
		}
		ui.SyncBorderlessChromeFrame(titleBar)
		// Chrome drag / wheel band — set BEFORE SyncBorderlessClientIfChanged so a
		// title-bar edge drag Resize does not remeasure preview RichText mid-drag.
		if titleBar != nil {
			ui.SetChromeTitleBarDragging(titleBar.IsDragging() || titleBar.IsTitleClickPending())
			// Do not OR frameResized here — that delayed preview settle one extra frame
			// after mouse-up. Mid-drag is already gated by the early SetChromeWindowMoving
			// before Sync + IsResizing() while the edge is held.
			resizingChrome := titleBar.IsDragging() || titleBar.IsTitleClickPending() ||
				titleBar.IsResizing() || ui.FillClientChromeResizing()
			ui.SetChromeWindowMoving(resizingChrome)
			if titleBar.IsResizing() {
				resizeHold.NoteGesture(frameStart)
			}
			ui.SetWheelSuppressBandY(ui.TitleBarHeight)
		}
		if w, h, ok := ui.SyncBorderlessClientIfChanged(doc, titleBar, windowW, windowH); ok {
			windowW, windowH = w, h
			setBreakpointIfChanged(windowW)
			frameResized = true
			ui.SetChromeWindowMoving(true)
		}
		if borderless && titleBar.IsDragging() {
			ui.ResetWheelScrollGesture()
		}
		ui.PrepareWheelScroll(doc.Root)
		if doc.Root.IsDirty() || ui.SubtreeLayoutDirty(doc.Root) {
			doc.Root.Layout()
		}

		ui.ModalMgr.Update(dt)
		ui.DrawerMgr.Update(dt)
		ui.BottomSheetMgr.Update(dt)

		mouse := rl.GetMousePosition()
		titleBarBlocks := borderless && mouse.Y < ui.TitleBarHeight
		resizeInputBlock := borderless && (titleBar.IsResizing() || titleBar.IsDragging() || ui.FillClientChromeResizing())
		ui.SetScenePointerBlocked(ui.OverlayBlocksSceneInput() || resizeInputBlock || titleBarBlocks)
		if !ui.ScenePointerBlocked() {
			ui.ProcessSwitchListTilePointers(doc.Root, dt)
		}
		ui.BeginFrameCursor()
		doc.Root.Update(dt)
		// Dismiss context menu before focus routing.
		ui.ContextMenuMgr.Update(dt)
		scene.OnUpdate(doc, dt)
		if rl.IsWindowReady() {
			doc.SetPlatformWindowHandle(uintptr(rl.GetWindowHandle()))
		}
		ui.SetScenePointerBlocked(false)
		titleBar.ApplyResizeCursor(windowW, windowH)

		ui.Tooltips.Update(dt)
		ui.Toasts.Update(dt)
		ui.NotificationCenterMgr.Update(dt)

		if doc.Root.IsDirty() || ui.SubtreeLayoutDirty(doc.Root) {
			doc.Root.Layout()
		}
		ui.DrainFillClientChromeResize()
		if w, h, ok := ui.SyncBorderlessClientIfChanged(doc, titleBar, windowW, windowH); ok {
			windowW, windowH = w, h
			setBreakpointIfChanged(windowW)
			frameResized = true
			ui.SetChromeWindowMoving(true)
			if doc.Root.IsDirty() || ui.SubtreeLayoutDirty(doc.Root) {
				doc.Root.Layout()
			}
		}
		ui.SyncWebViewHosts(doc)
		focusBlock := resizeInputBlock
		if titleBarBlocks {
			if pos, ok := ui.PeekFocusHandoffClick(); ok && pos.Y >= ui.TitleBarHeight {
				// content click — allow focus routing
			} else {
				focusBlock = true
			}
		}
		if !ui.OverlayBlocksSceneInput() && !focusBlock {
			ui.RouteScenePointerFocus(doc)
		}

		// RT / client mismatch: defer recreate while dragging; never stretch-blit.
		if ui.PresentSizeMismatch(windowW, windowH) {
			resizingNow := frameResized || resizeHold.KeepActiveFPS(frameStart) || titleBar.IsResizing()
			if resizingNow {
				ui.DeferWindowTextures(windowW, windowH)
			} else {
				ui.FlushPendingWindowTextures()
				if ui.PresentSizeMismatch(windowW, windowH) {
					ui.ResizeWindowTextures(windowW, windowH)
				}
			}
			frameResized = true
		}

		fullRedraw := false
		if frameResized || titleBar.IsResizing() {
			doc.InvalidatePaint()
		}

		// ARCHIVE 2026-08-21: while RT/client mismatch, skip SSAA pass — one direct
		// draw at present. Filling a stale RT then blitting was double work + lag.
		sizeMismatch := ui.PresentSizeMismatch(windowW, windowH)
		needSuper := !sizeMismatch && (doc.NeedsRedraw() || ui.SuperTargetNeedsFill())
		if needSuper && ui.SuperTargetDrawable() {
			fullRedraw = true
			ui.BeginSuperFrame(bg, borderless, windowW, windowH)
			doc.Root.Draw()
			titleBar.Draw()
			ui.EndSuperFrame()
			ui.RecordCacheMiss()
			ui.ClearDrawDirtySubtree(doc.Root)
		} else if needSuper {
			ui.RecordCacheMiss()
			doc.Root.MarkDrawDirty()
		} else if !sizeMismatch {
			ui.RecordCacheHit()
			ui.ClearDrawDirtySubtree(doc.Root)
		}

		rl.BeginDrawing()
		ui.BeginDrawingBorderless(bg, borderless, windowW, windowH)
		canBlit := !sizeMismatch && ui.SuperTargetDrawable() && !ui.SuperTargetNeedsFill()
		if canBlit {
			ui.BlitToScreenBorderless(windowW, windowH, borderless)
		} else {
			// Mismatch / RT not ready — live-client direct draw (never stretch).
			doc.Root.Draw()
			titleBar.Draw()
			ui.ClearDrawDirtySubtree(doc.Root)
			ui.RecordCacheMiss()
			fullRedraw = true
		}
		if borderless && !fullRedraw && titleBar.ChromeNeedsPostBlit() {
			titleBar.DrawCaptionHoverPlates()
		}
		ui.DrawAnimationOverlays(doc.Root)
		drawInteractionOverlay(windowW, windowH, doc.Root)
		drawScreenOverlays(windowW, windowH, doc.Root)
		rl.EndDrawing()

		frameEnd := time.Now()
		if borderless && titleBar.IsResizing() {
			resizeHold.NoteGesture(frameEnd)
		}

		// Settle: recreate deferred RT once chrome is idle.
		ui.SettleBorderlessClientTextures(titleBar)

		if !windowShown {
			if rl.IsWindowHidden() {
				rl.ClearWindowState(uint32(rl.FlagWindowHidden))
			}
			windowShown = true
			raiseRunningWindow()
			appicon.ReapplyWindowIcon()
		}
	}
}

func drawScreenOverlays(windowW, windowH int32, root ui.Node) {
	if ui.SupersamplingActive() && ui.OverlayTargetDrawable() {
		ui.BeginOverlaySuperFrame()
		ui.Tooltips.Draw()
		ui.ModalMgr.Draw()
		ui.DrawerMgr.Draw()
		ui.BottomSheetMgr.Draw()
		ui.ContextMenuMgr.Draw()
		ui.NotificationCenterMgr.Draw()
		ui.Toasts.Draw()
		ui.DrawOpenMenuPopups(root)
		ui.EndOverlaySuperFrame()
		ui.BlitOverlayToScreen(windowW, windowH)
		return
	}
	ui.Tooltips.Draw()
	ui.ModalMgr.Draw()
	ui.DrawerMgr.Draw()
	ui.BottomSheetMgr.Draw()
	ui.ContextMenuMgr.Draw()
	ui.NotificationCenterMgr.Draw()
	ui.Toasts.Draw()
	ui.DrawOpenMenuPopups(root)
}

func drawInteractionOverlay(windowW, windowH int32, root ui.Node) {
	if root == nil {
		return
	}
	if ui.SupersamplingActive() && ui.OverlayTargetDrawable() {
		ui.BeginOverlaySuperFrame()
		ui.DrawInteractionOverlays(root)
		ui.EndOverlaySuperFrame()
		ui.BlitOverlayToScreen(windowW, windowH)
		return
	}
	ui.DrawInteractionOverlays(root)
}
