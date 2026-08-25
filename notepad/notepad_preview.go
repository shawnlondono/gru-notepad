package notepad

import (
	"strings"

	"github.com/ledocorp/gru/ui"
	"github.com/ledocorp/gru/ui/preview"
)

const (
	notepadPreviewDebounce    = float32(0.3)
	notepadPreviewEditorRatio = float32(0.52)
	notepadPreviewMaxBytes    = 256 * 1024 // cap compile on huge buffers (paste / big files)
)

func (s *notepadScene) previewEditorSource() string {
	if s.editor == nil {
		return ""
	}
	source := notepadNormalizeText(s.editor.Text.Get())
	if len(source) > notepadPreviewMaxBytes {
		source = source[:notepadPreviewMaxBytes] + "\n\n… preview truncated"
	}
	return source
}

func (s *notepadScene) applyPreviewLayout() {
	if s.split == nil || s.split.Ratio == nil {
		return
	}
	if s.useMobilePanelSheets() {
		// Notes sheet wins mainSplit; preview only when notes is closed.
		if s.showPreview && !s.showSession {
			s.split.MinFirst = 0
			s.split.MinSecond = 0
			s.split.Ratio.Set(0)
			if s.previewVP != nil {
				s.previewVP.Show()
			}
			if s.mdPreview != nil {
				s.mdPreview.Show()
			}
		} else {
			s.split.MinSecond = 0
			s.split.Ratio.Set(1)
			if s.previewVP != nil {
				s.previewVP.Hide()
			}
			if s.mdPreview != nil {
				s.mdPreview.Hide()
			}
		}
		s.split.MarkDirty()
		s.split.MarkDrawDirty()
		if s.split.Parent() != nil {
			s.split.Parent().MarkDirty()
			s.split.Parent().MarkDrawDirty()
		}
		// Editor was ~0 width while preview sheet was open (or vice versa).
		// Settle next idle frame — same fix a manual resize provides.
		s.paneSettlePending = true
		return
	}
	if s.showPreview {
		s.split.MinSecond = 120
		s.split.Ratio.Set(notepadPreviewEditorRatio)
		if s.previewVP != nil {
			s.previewVP.Show()
		}
		if s.mdPreview != nil {
			s.mdPreview.Show()
		}
	} else {
		s.split.MinSecond = 0
		s.split.Ratio.Set(1)
		if s.previewVP != nil {
			s.previewVP.Hide()
		}
		if s.mdPreview != nil {
			s.mdPreview.Hide()
		}
	}
	s.split.MarkDirty()
	s.split.MarkDrawDirty()
	if s.split.Parent() != nil {
		s.split.Parent().MarkDirty()
		s.split.Parent().MarkDrawDirty()
	}
}

func (s *notepadScene) schedulePreview() {
	if !s.showPreview || s.mdPreview == nil {
		return
	}
	s.previewPending = true
	s.previewTimer = notepadPreviewDebounce
}

// refreshPreviewForDocument runs after Open or switching notes while preview may
// already be visible. It relayouts the split (fixes clip bounds) and rebuilds the
// preview immediately so we do not show the previous note's render for a frame.
func (s *notepadScene) refreshPreviewForDocument() {
	s.previewPending = false
	s.previewTimer = 0
	s.applyPreviewLayout()
	if s.mdPreview == nil || s.uiDoc == nil {
		return
	}
	if s.uiDoc.Root != nil {
		s.uiDoc.Root.Layout()
	}
	if s.showPreview {
		s.rebuildPreviewFull()
	}
	if s.previewVP != nil {
		s.previewVP.MarkDirty()
	}
	if s.editorVP != nil {
		s.editorVP.MarkDirty()
	}
	if s.split != nil {
		s.split.MarkDirty()
	}
}

func (s *notepadScene) tickPreview(dt float32) {
	if !s.showPreview || !s.previewPending {
		return
	}
	if ui.ChromeWindowMoving() {
		s.previewTimer = 0.12
		return
	}
	// Session 14: hold rebuild while typing so backspace stays on the editor.
	if ui.TypingGestureActive() {
		s.previewTimer = notepadPreviewDebounce
		return
	}
	s.previewTimer -= dt
	if s.previewTimer > 0 {
		return
	}
	s.previewPending = false
	s.rebuildPreview()
}

// syncClosedPreviewSplit keeps the editor full-width while preview is off (avoids
// a blank second pane during window resize).
func (s *notepadScene) syncClosedPreviewSplit() {
	if s.showPreview || s.split == nil || s.split.Ratio == nil {
		return
	}
	if s.split.Ratio.Get() >= 0.999 && (s.previewVP == nil || s.previewVP.IsHidden()) {
		return
	}
	s.applyPreviewLayout()
}

// softResetPreviewOnHostResize coalesces preview reflow during host resize.
// Session 14 contract (smooth mid-drag):
//   - while resizing: set previewResizePending only
//   - never SimulateResizeReflow / applyBreakpointLayout mid-gesture
//   - on chrome idle: settle immediately (same frame — no QueueMain lag)
func (s *notepadScene) softResetPreviewOnHostResize(doc *ui.Document) {
	if doc == nil {
		return
	}
	w := float32(doc.Width)
	if w <= 1 {
		return
	}
	prevFrame := s.prevFramePreviewHostW
	if prevFrame <= 1 {
		s.prevFramePreviewHostW = w
		s.lastPreviewHostW = w
		return
	}
	fdw := w - prevFrame
	if fdw < 0 {
		fdw = -fdw
	}
	s.prevFramePreviewHostW = w

	if ui.ChromeWindowMoving() {
		if fdw > 0.5 && s.showPreview && s.mdPreview != nil && !s.inSettings {
			s.previewResizePending = true
		}
		return
	}
	// ── Gesture ended — settle ASAP ──────────────────────────────────────────
	if !s.showPreview {
		s.syncClosedPreviewSplit()
	}
	if s.breakpointLayoutPending {
		s.breakpointLayoutPending = false
		s.applyBreakpointLayout()
	}
	if s.showPreview && s.mdPreview != nil && !s.inSettings && s.previewResizePending {
		s.flushPreviewSettleReflow()
		return
	}
	if !s.previewResizePending {
		s.lastPreviewHostW = w
	}
}

func (s *notepadScene) tickPreviewReflow(dt float32) {
	if !s.previewResizePending || s.mdPreview == nil || s.uiDoc == nil {
		return
	}
	if ui.ChromeWindowMoving() {
		return
	}
	_ = dt
	s.flushPreviewSettleReflow()
}

// flushPreviewSettleReflow runs the 1px preview alignment pass immediately on release.
func (s *notepadScene) flushPreviewSettleReflow() {
	if s.mdPreview == nil || s.uiDoc == nil {
		return
	}
	if ui.ChromeWindowMoving() {
		s.previewResizePending = true
		return
	}
	if !s.showPreview || s.inSettings {
		s.previewResizePending = false
		s.previewReflowDelay = 0
		return
	}
	s.previewResizePending = false
	s.previewReflowDelay = 0
	s.lastPreviewHostW = float32(s.uiDoc.Width)
	s.mdPreview.SimulateResizeReflow()
	s.settleNotesPreviewIdle()
}

func (s *notepadScene) rebuildPreview() {
	if s.mdPreview == nil || s.uiDoc == nil {
		return
	}
	if ui.ChromeWindowMoving() || ui.TypingGestureActive() {
		s.previewPending = true
		s.previewTimer = notepadPreviewDebounce
		return
	}
	source := s.previewEditorSource()
	if strings.TrimSpace(source) == "" {
		s.mdPreview.SetMarkdown("")
		s.mdPreview.MarkDirty()
		if s.previewVP != nil {
			s.previewVP.MarkDirty()
		}
		return
	}
	s.mdPreview.SetDocument(s.uiDoc)
	// Live edits: async parse + incremental mount (Session 14). Sync EnsureFullyBuilt
	// on every debounce was the typing/backspace stall with preview open.
	s.mdPreview.SetMarkdownAsync(s.uiDoc, source)
	if s.previewVP != nil {
		s.previewVP.MarkDirty()
	}
}

// rebuildPreviewFull synchronously materializes preview blocks (note open/switch).
func (s *notepadScene) rebuildPreviewFull() {
	if s.mdPreview == nil || s.uiDoc == nil {
		return
	}
	if ui.ChromeWindowMoving() {
		s.previewPending = true
		s.previewTimer = 0.12
		return
	}
	source := s.previewEditorSource()
	if strings.TrimSpace(source) == "" {
		s.mdPreview.SetMarkdown("")
		s.mdPreview.MarkDirty()
		if s.previewVP != nil {
			s.previewVP.MarkDirty()
		}
		return
	}
	s.mdPreview.SetDocument(s.uiDoc)
	s.mdPreview.SetMarkdown(source)
	// Note open/switch: above-fold sync so the pane is usable; rest mounts incrementally.
	s.mdPreview.EnsureAboveFoldBuilt()
	if s.uiDoc.Root != nil {
		s.uiDoc.Root.Layout()
	}
	s.mdPreview.RelayoutAfterContentChange()
	if s.previewVP != nil {
		s.previewVP.MarkDirty()
	}
	s.mdPreview.MarkIdleStable()
	s.settleNotesPreviewIdle()
}

// notepadDocWarmMarkdown reports whether an open buffer is worth a light parse warm.
func notepadDocWarmMarkdown(d *notepadDoc) bool {
	if d == nil || strings.TrimSpace(d.Text) == "" {
		return false
	}
	lang := notepadSyntaxLanguage(d.Path)
	if lang == "markdown" {
		return true
	}
	if lang != "" {
		return false
	}
	t := strings.TrimSpace(d.Text)
	return strings.HasPrefix(t, "#") ||
		strings.Contains(t, "\n## ") ||
		strings.Contains(t, "```")
}

// warmInactiveMarkdownCaches pre-parses non-active markdown notes on workers.
// Does not build MarkdownView widgets — only the goldmark AST cache.
func (s *notepadScene) warmInactiveMarkdownCaches() {
	if s == nil || s.sess == nil {
		return
	}
	activeID := s.sess.activeID
	for _, d := range s.sess.docs {
		if d == nil || d.ID == activeID || !notepadDocWarmMarkdown(d) {
			continue
		}
		src := d.Text
		if len(src) > notepadPreviewMaxBytes {
			src = src[:notepadPreviewMaxBytes]
		}
		preview.WarmMarkdownCache(src)
	}
}

// settleNotesPreviewIdle clears draw-only dirty on notes/preview chrome after layout
// so nested SplitViews can reach deep idle with both panes visible.
func (s *notepadScene) settleNotesPreviewIdle() {
	if s.openNotes != nil && s.openNotes.Root != nil {
		ui.ClearDrawDirtySubtree(s.openNotes.Root)
	}
	if s.previewVP != nil {
		ui.ClearDrawDirtySubtree(s.previewVP)
	}
	if s.mainSplit != nil {
		ui.ClearDrawDirtySubtree(s.mainSplit)
	}
	if s.split != nil {
		ui.ClearDrawDirtySubtree(s.split)
	}
}

// settleEditorAfterShellMount remasures AutoHeight editor at real pane width and
// clamps scroll. Early mount could wrap at a narrow width and leave a giant empty
// region under the text (ScrollY stuck near the old bottom).
func (s *notepadScene) settleEditorAfterShellMount() {
	s.settleEditorAndPreviewPanes()
}

// settleEditorAndPreviewPanes remasures editor (+ preview when open) after shell
// mount, note switch, or mobile sheet transitions. Same path a window resize uses.
func (s *notepadScene) settleEditorAndPreviewPanes() {
	if s.editor != nil {
		s.editor.InvalidateAutoHeightMeasure()
		s.editor.Layout()
		if s.editorVP != nil {
			s.editorVP.ScrollY = 0
			s.editorVP.MarkDirty()
			s.editorVP.Layout()
			s.editorVP.MarkDrawDirty()
		}
		s.editor.MarkDrawDirty()
	}
	if s.showPreview && s.mdPreview != nil && !s.mdPreview.IsHidden() {
		s.mdPreview.SimulateResizeReflow()
		if s.previewVP != nil {
			s.previewVP.ScrollY = 0
			s.previewVP.MarkDirty()
			s.previewVP.Layout()
			s.previewVP.MarkDrawDirty()
		}
	}
}

func (s *notepadScene) flushPaneSettleIfPending() {
	if !s.paneSettlePending {
		return
	}
	if ui.ChromeWindowMoving() {
		return
	}
	s.paneSettlePending = false
	if s.uiDoc != nil && s.uiDoc.Root != nil {
		s.uiDoc.Root.Layout()
	}
	s.settleEditorAndPreviewPanes()
}

func (s *notepadScene) toggleMarkdownPreview() {
	s.showPreview = !s.showPreview
	if s.sess != nil && s.sess.active() != nil {
		s.sess.active().ShowPreview = s.showPreview
	}
	if s.showPreview {
		if s.useMobilePanelSheets() {
			s.applySessionPanelLayout()
		}
		s.refreshPreviewForDocument()
		s.paneSettlePending = true
		if s.uiDoc != nil {
			s.uiDoc.QueueMain(func() { s.warmInactiveMarkdownCaches() })
		}
	} else {
		s.previewPending = false
		s.applyPreviewLayout()
		s.paneSettlePending = true
	}
	s.syncRibbonToggleState()
}
