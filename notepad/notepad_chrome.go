// Gru Notepad — Phase 4 chrome: menubar vs ribbon toolbar.
package notepad

import (
	"time"

	"github.com/ledocorp/gru/ui"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	notepadChromeMenubar = "menubar"
	notepadChromeRibbon  = "ribbon"
)

const notepadRibbonH float32 = 128
const notepadMobileRibbonH float32 = 84

func (s *notepadScene) preloadNotepadRibbonIcons() {
	icons := []string{
		ui.PhosphorFile,
		ui.PhosphorFolderOpen,
		ui.PhosphorSave2,
		ui.PhosphorDownload,
		ui.PhosphorCopy,
		ui.PhosphorClipboard,
		ui.PhosphorScissorsCut,
		ui.PhosphorTrash,
		ui.PhosphorTray,
		ui.PhosphorPencilSimple,
		ui.PhosphorMagnifyingGlass,
		ui.PhosphorFindReplace,
		ui.PhosphorMarkdownLine,
		ui.PhosphorSettings4,
		ui.PhosphorSettings3,
		ui.PhosphorListSettings,
		ui.PhosphorTextWrap,
		ui.PhosphorTextUnderline,
		ui.PhosphorInfoI,
		ui.PhosphorArrowGoBack,
		ui.PhosphorArrowGoForward,
		ui.PhosphorList,
		ui.PhosphorCheck,
		ui.PhosphorCodeView,
		ui.PhosphorCodeBlock,
		ui.PhosphorZoomIn,
		ui.PhosphorZoomOut,
		ui.PhosphorFileClose,
		ui.PhosphorCaretRight,
		ui.PhosphorCaretLeft,
		ui.PhosphorCaretDown,
		ui.PhosphorCaretUp,
		ui.PhosphorSun,
		ui.PhosphorResize,
		ui.PhosphorHome2,
		ui.PhosphorEditBox,
		ui.PhosphorBookRead,
		ui.PhosphorSearch,
		ui.PhosphorCodeBox,
		ui.PhosphorDotsThreeVertical,
	}
	// Warm chrome icons. Prefer EnsureLoadedMany when available on newer GRU;
	// v0.7.1 only has EnsureLoaded (per-glyph atlas rebuild).
	for _, name := range icons {
		ui.Phosphor.EnsureLoaded(name, ui.PhosphorRegular)
	}
}

func (s *notepadScene) buildNotepadRibbon() *ui.Toolbar {
	tb := ui.NewToolbar("notepad-ribbon", 0, 0, 0, notepadRibbonH)
	tb.SetStyle("toolbar-ribbon")
	tb.SetZIndex(100) // above work area so preview content cannot paint over chrome
	tb.Overflow = true
	tb.OverflowKind = ui.ToolbarOverflowScroll

	s.ribbonNotesChecked = ui.NewSignal(s.showSession)
	s.ribbonPreviewChecked = ui.NewSignal(s.showPreview)
	s.ribbonWrapChecked = ui.NewSignal(s.wordWrap)
	s.ribbonStatusChecked = ui.NewSignal(s.showStatus)
	s.ribbonSpellChecked = ui.NewSignal(s.spellCheck)
	s.ribbonSyntaxChecked = ui.NewSignal(s.syntaxHighlight)

	home := tb.AddRibbonTab("Home")
	file := tb.AddSection(home, "notepad-ribbon-file", "File")
	tb.AddPhosphorIconButton(file.ID, "new", ui.PhosphorFile, "New", "New note (Ctrl+N)", func() { s.doNew() })
	tb.AddPhosphorIconButton(file.ID, "open", ui.PhosphorFolderOpen, "Open", "Open file (Ctrl+O)", func() { s.pickOpen() })
	tb.AddPhosphorIconButton(file.ID, "save", ui.PhosphorSave2, "Save", "Save (Ctrl+S)", func() { s.doSave(false) })
	tb.AddPhosphorIconButton(file.ID, "save-as", ui.PhosphorCopy, "Save As", "Save As (Ctrl+Shift+S)", func() { s.doSave(true) })
	tb.AddPhosphorIconButton(file.ID, "save-all", ui.PhosphorDownload, "Save All", "Save all open notes", func() { s.doSaveAll() })
	tb.AddPhosphorIconButton(file.ID, "close", ui.PhosphorFileClose, "Close", "Close note", func() { s.doCloseNote() })
	tb.AddPhosphorIconButton(file.ID, "recent", ui.PhosphorTray, "Recent", "Open recent file", func() {
		items := s.recentMenuItems()
		if len(items) == 0 {
			ui.ShowToast("No recent files", ui.ToastInfo, 1500*time.Millisecond)
			return
		}
		if s.ribbonRecentBtn == nil {
			return
		}
		r := s.ribbonRecentBtn.Bounds()
		ui.ShowContextMenu(items, r.X, r.Y+r.Height)
	})
	tb.AddPhosphorIconButton(file.ID, "settings", ui.PhosphorSettings3, "Settings", "Settings", func() { s.showSettingsPage() })
	// CaretLeft — distinct from Edit→Undo (ArrowGoBack) so Settings exit is obvious.
	tb.AddPhosphorIconButton(file.ID, "back", ui.PhosphorCaretLeft, "Back", "Back to editor", func() { s.closeSettingsPage() })

	editTab := tb.AddRibbonTab("Edit")
	clipboard := tb.AddSection(editTab, "notepad-ribbon-clipboard", "Clipboard")
	tb.AddPhosphorIconButton(clipboard.ID, "undo", ui.PhosphorArrowGoBack, "Undo", "Undo (Ctrl+Z)", func() {
		if s.editor != nil {
			s.editor.Undo()
		}
	})
	tb.AddPhosphorIconButton(clipboard.ID, "redo", ui.PhosphorArrowGoForward, "Redo", "Redo (Ctrl+Y)", func() {
		if s.editor != nil {
			s.editor.Redo()
		}
	})
	tb.AddPhosphorIconButton(clipboard.ID, "cut", ui.PhosphorScissorsCut, "Cut", "Cut (Ctrl+X)", func() {
		if s.editor != nil {
			s.editor.Cut()
		}
	})
	tb.AddPhosphorIconButton(clipboard.ID, "copy", ui.PhosphorCopy, "Copy", "Copy (Ctrl+C)", func() {
		if s.editor != nil {
			s.editor.Copy()
		}
	})
	tb.AddPhosphorIconButton(clipboard.ID, "paste", ui.PhosphorClipboard, "Paste", "Paste (Ctrl+V)", func() {
		if s.editor != nil {
			s.editor.Paste()
		}
	})
	tb.AddPhosphorIconButton(clipboard.ID, "delete", ui.PhosphorTrash, "Delete", "Delete selection (Del)", func() {
		if s.editor != nil && s.editor.HasSelection() {
			s.editor.Delete()
		}
	})
	tb.AddPhosphorIconButton(clipboard.ID, "select-all", ui.PhosphorCheck, "All", "Select all (Ctrl+A)", func() {
		if s.editor != nil {
			s.editor.SelectAll()
		}
	})

	proof := tb.AddSection(editTab, "notepad-ribbon-proof", "Proofing")
	tb.AddPhosphorRibbonToggle(proof.ID, "spell", ui.PhosphorTextUnderline, "Spell", "Spell check", s.ribbonSpellChecked, func() { s.toggleSpellCheck() })

	searchTab := tb.AddRibbonTab("Search")
	s.ribbonTabSearch = searchTab
	find := tb.AddSection(searchTab, "notepad-ribbon-find", "Find")
	s.ribbonFindBar = ui.NewTextInput("notepad-ribbon-find-bar", s.findQuery, 0, 0, 0, 36)
	s.ribbonFindBar.SetStyle("toolbar-menu")
	// Match settings row titles — type size only (theme owns ink on light/dark).
	if tile := ui.GetThemeStyle("list-tile"); tile.FontSize > 0 {
		s.ribbonFindBar.SetStyleOverrides(ui.Style{
			FontSize:      tile.FontSize,
			FontDensity:   tile.FontDensity,
			NoCaptionBold: tile.NoCaptionBold,
		})
	}
	s.ribbonFindBar.Placeholder = "Find…"
	s.ribbonFindBar.PreferredWidth = ribbonFindBarW
	s.ribbonFindBar.MinWidth = ribbonFindBarMinW
	s.ribbonFindBar.MaxWidth = ribbonFindBarMaxW
	s.wireRibbonFindBar(s.ribbonFindBar)
	tb.AddCustom(find.ID, "find-bar", s.ribbonFindBar)
	tb.AddPhosphorIconButton(find.ID, "find", ui.PhosphorMagnifyingGlass, "Find", "Find (Ctrl+F)", func() { s.showFindDialog() })
	tb.AddPhosphorIconButton(find.ID, "find-next", ui.PhosphorCaretDown, "Next", "Find next (F3)", func() { s.findNextOrPrompt() })
	tb.AddPhosphorIconButton(find.ID, "find-prev", ui.PhosphorCaretUp, "Prev", "Find previous (Shift+F3)", func() { s.findPreviousOrPrompt() })
	tb.AddPhosphorIconButton(find.ID, "replace", ui.PhosphorFindReplace, "Replace", "Replace (Ctrl+H)", func() { s.showReplaceDialog() })
	tb.AddPhosphorIconButton(find.ID, "goto", ui.PhosphorCodeView, "Line", "Go to line (Ctrl+G)", func() { s.showGoToDialog() })
	s.syncRibbonFindMode()

	view := tb.AddRibbonTab("View")
	panes := tb.AddSection(view, "notepad-ribbon-panes", "Panes")
	tb.AddPhosphorRibbonToggle(panes.ID, "notes", ui.PhosphorList, "Notes", "Open notes panel", s.ribbonNotesChecked, func() { s.toggleSessionPanel() })
	tb.AddPhosphorRibbonToggle(panes.ID, "preview", ui.PhosphorMarkdownLine, "Preview", "Markdown preview", s.ribbonPreviewChecked, func() { s.toggleMarkdownPreview() })
	tb.AddPhosphorRibbonToggle(panes.ID, "wrap", ui.PhosphorTextWrap, "Wrap", "Word wrap", s.ribbonWrapChecked, func() { s.toggleWordWrap() })
	tb.AddPhosphorRibbonToggle(panes.ID, "status", ui.PhosphorInfoI, "Status", "Status bar", s.ribbonStatusChecked, func() { s.toggleStatusBar() })

	zoom := tb.AddSection(view, "notepad-ribbon-zoom", "Zoom")
	tb.AddPhosphorIconButton(zoom.ID, "zoom-in", ui.PhosphorZoomIn, "In", "Zoom in (Ctrl++)", func() { s.zoomBy(0.1) })
	tb.AddPhosphorIconButton(zoom.ID, "zoom-out", ui.PhosphorZoomOut, "Out", "Zoom out (Ctrl+-)", func() { s.zoomBy(-0.1) })
	tb.AddPhosphorIconButton(zoom.ID, "zoom-reset", ui.PhosphorMagnifyingGlass, "100%", "Reset zoom (Ctrl+0)", func() { s.setZoom(1) })

	syntaxTab := tb.AddRibbonTab("Syntax")
	syntax := tb.AddSection(syntaxTab, "notepad-ribbon-syntax", "Syntax")
	tb.AddPhosphorRibbonToggle(syntax.ID, "highlight", ui.PhosphorCodeView, "Highlight", "Syntax highlighting", s.ribbonSyntaxChecked, func() { s.toggleSyntaxHighlight() })
	tb.AddPhosphorIconButton(syntax.ID, "lang", ui.PhosphorCodeBlock, "Language", "Pick syntax language", func() {
		if s.ribbonLangBtn == nil {
			return
		}
		r := s.ribbonLangBtn.Bounds()
		ui.ShowContextMenu(s.syntaxLangMenuItems(), r.X, r.Y+r.Height)
	})
	chromaOpts := notepadChromaStyleOptions()
	tb.AddMenuButton(syntax.ID, "chroma-theme", "Theme", chromaOpts, func(i int) {
		if i < 0 || i >= len(chromaOpts) {
			return
		}
		s.chromaStyleName = chromaOpts[i]
		if s.chromaStyleIdx != nil {
			s.chromaStyleIdx.Set(i)
		}
		s.applyNotepadThemePrefs()
		s.persistQuickPrefs()
	})
	if w := tb.Widget("chroma-theme"); w != nil {
		if dd, ok := w.(*ui.Dropdown); ok {
			s.ribbonChromaDD = dd
			dd.PreferredWidth = 128
			dd.MaxWidth = 160
			dd.SelectedIndex.Set(notepadChromaStyleIndex(s.chromaStyleName))
		}
	}

	if w := tb.Widget("recent"); w != nil {
		s.ribbonRecentBtn, _ = w.(*ui.IconButton)
	}
	if w := tb.Widget("lang"); w != nil {
		s.ribbonLangBtn, _ = w.(*ui.IconButton)
	}
	tb.SetItemSuppressed("back", true)
	s.syncRibbonFileButtons()

	return tb
}

// syncRibbonFileButtons mirrors menubar-only File affordances onto ribbon cells.
func (s *notepadScene) syncRibbonFileButtons() {
	if s.ribbon == nil {
		return
	}
	in := s.inSettings
	fileIDs := []string{"new", "open", "save", "save-as", "save-all", "close", "recent", "settings"}
	for _, id := range fileIDs {
		s.ribbon.SetItemSuppressed(id, in)
	}
	s.ribbon.SetItemSuppressed("back", !in)
	if in {
		// Back lives on the File tab; switch there if Settings was opened from
		// another tab. Compatible with GRU v0.7.1 (no SelectRibbonTabForItem).
		s.ribbon.ActiveGroup.Set(0)
		s.ribbon.SetItemSuppressed("back", true)
		s.ribbon.SetItemSuppressed("back", false)
	}
	s.ribbon.MarkDirty()
}

// syncRibbonToggleState mirrors pane/wrap prefs onto ribbon toggle cells.
func (s *notepadScene) syncRibbonToggleState() {
	if s.ribbonNotesChecked != nil {
		s.ribbonNotesChecked.Set(s.showSession)
	}
	if s.ribbonPreviewChecked != nil {
		s.ribbonPreviewChecked.Set(s.showPreview)
	}
	if s.ribbonWrapChecked != nil {
		s.ribbonWrapChecked.Set(s.wordWrap)
	}
	if s.ribbonStatusChecked != nil {
		s.ribbonStatusChecked.Set(s.showStatus)
	}
	if s.ribbonSpellChecked != nil {
		s.ribbonSpellChecked.Set(s.spellCheck)
	}
	if s.ribbonSyntaxChecked != nil {
		s.ribbonSyntaxChecked.Set(s.syntaxHighlight)
	}
}

func (s *notepadScene) effectiveChromeMode(windowW float32) string {
	if ui.CurrentBreakpoint(windowW) < ui.BreakpointMD {
		return notepadChromeRibbon
	}
	if s.chromeMode == notepadChromeRibbon {
		return notepadChromeRibbon
	}
	return notepadChromeMenubar
}

func (s *notepadScene) applyChromeLayout() {
	s.reconcileChrome()
}

// reconcileChrome applies menubar/ribbon/mobile-nav visibility and ribbon compact
// mode from prefs + window width. Call after settings apply, close settings,
// breakpoint changes, and mobile-nav pref toggles so desktop chrome never stays stale.
func (s *notepadScene) reconcileChrome() {
	mobile := s.useMobileBottomNav()

	s.applyMobileRibbonChrome()

	if !mobile && s.bottomNav != nil && !s.bottomNav.IsHidden() {
		s.bottomNav.Hide()
	}

	winW := float32(rl.GetScreenWidth())
	useRibbon := s.effectiveChromeMode(winW) == notepadChromeRibbon

	if mobile {
		if s.menuBar != nil {
			s.menuBar.Hide()
		}
		if s.ribbon != nil {
			s.ribbon.Show()
		}
	} else {
		if s.menuBar != nil {
			if useRibbon {
				s.menuBar.Hide()
			} else {
				s.menuBar.Show()
			}
		}
		if s.ribbon != nil {
			if useRibbon {
				s.ribbon.Show()
			} else {
				s.ribbon.Hide()
			}
		}
	}

	s.syncMobileBottomChrome()
	s.syncRibbonFindMode()

	// Layout visible chrome only; avoid pinning SSAA dirty on hidden ribbon/menubar.
	if s.ribbon != nil && !s.ribbon.IsHidden() {
		s.ribbon.Layout()
	}
	if s.menuBar != nil && !s.menuBar.IsHidden() {
		s.menuBar.Layout()
	}
}

func (s *notepadScene) wireChromeResponsive() {
	ui.ActiveBreakpoint.Subscribe(func() {
		s.applyChromeLayout()
		s.applyResponsivePanelLayout()
	})
}
