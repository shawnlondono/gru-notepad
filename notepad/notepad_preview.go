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
		s.rebuildPreview()
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
	s.previewTimer -= dt
	if s.previewTimer > 0 {
		return
	}
	s.previewPending = false
	s.rebuildPreview()
}

// softResetPreviewOnHostResize nudges MarkdownView when the document width changes.
// Studio's Markdown Preview demo gets this via examples.FinishShellMount + simple full-width
// layout; Notepad's SplitView pane often needs an explicit SimulateResizeReflow so
// RichText/tables remeasure. Scoped to notepadScene — does not touch ui/preview demos.
func (s *notepadScene) softResetPreviewOnHostResize(doc *ui.Document) {
	if !s.showPreview || s.mdPreview == nil || doc == nil {
		return
	}
	w := float32(doc.Width)
	if w <= 1 {
		return
	}
	if s.lastPreviewHostW > 1 && (w > s.lastPreviewHostW+0.5 || w < s.lastPreviewHostW-0.5) {
		s.mdPreview.SimulateResizeReflow()
		if s.uiDoc != nil && s.uiDoc.Root != nil {
			s.uiDoc.Root.MarkDirty()
		}
	}
	s.lastPreviewHostW = w
}

func (s *notepadScene) rebuildPreview() {
	if s.mdPreview == nil || s.uiDoc == nil {
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
	// Full widget build only for the visible active note; other open notes warm
	// parse cache via warmInactiveMarkdownCaches (background, no EnsureFullyBuilt).
	s.mdPreview.SetMarkdown(source)
	if s.showPreview {
		s.mdPreview.EnsureFullyBuilt()
	}
	if s.uiDoc != nil && s.uiDoc.Root != nil {
		s.uiDoc.Root.Layout()
	}
	if s.mdPreview != nil {
		s.mdPreview.RelayoutAfterContentChange()
	}
	if s.previewVP != nil {
		s.previewVP.MarkDirty()
	}
	if s.mdPreview != nil {
		s.mdPreview.MarkIdleStable()
	}
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
		if s.uiDoc != nil {
			s.uiDoc.QueueMain(func() { s.warmInactiveMarkdownCaches() })
		}
	} else {
		s.previewPending = false
		s.applyPreviewLayout()
	}
	s.syncRibbonToggleState()
}
