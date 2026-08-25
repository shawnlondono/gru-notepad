// Gru Notepad — in-shell settings (desktop settings surface).
package notepad

import (
	"github.com/ledocorp/gru/ui"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func (s *notepadScene) buildNotepadSettingsPage() *ui.Container {
	page := ui.NewContainer("notepad-settings-page", 0, 0, 0, 0)
	page.LayoutType = ui.LayoutFlex
	page.FlexDirection = ui.FlexColumn
	page.SetStyle("settings-flat-band")
	page.SetFlexGrow(1)

	vp := ui.NewViewport("notepad-settings-vp", 0, 0, 0, 0)
	vp.SetStyle("settings-scroll")
	vp.FlexDirection = ui.FlexColumn
	vp.SetFlexGrow(1)
	vp.Gap = 20

	titleRow, back := settingsPageTitleRow("notepad-settings-page-title", "Settings", func() {
		s.closeSettingsPage()
	})
	s.settingsBack = back
	vp.AddChild(titleRow)

	intro := ui.NewLabel("notepad-settings-intro",
		"Saved defaults apply on startup and update this session when changed.",
		0, 0, 0, 0)
	intro.SetStyle("form-value")
	intro.Align = ui.LabelAlignLeft
	intro.Wrap = true
	vp.AddChild(intro)

	s.setWrapTog = ui.NewToggle("notepad-set-wrap", s.wordWrap, 0, 0, 52, 28)
	s.setStatusTog = ui.NewToggle("notepad-set-status", s.showStatus, 0, 0, 52, 28)
	s.setSyntaxTog = ui.NewToggle("notepad-set-syntax", s.syntaxHighlight, 0, 0, 52, 28)
	s.setSpellTog = ui.NewToggle("notepad-set-spell", s.spellCheck, 0, 0, 52, 28)
	s.setSessionTog = ui.NewToggle("notepad-set-session", s.showSession, 0, 0, 52, 28)
	s.setChromeTog = ui.NewToggle("notepad-set-chrome", s.chromeMode == notepadChromeRibbon, 0, 0, 52, 28)
	s.setMobileSheetsTog = ui.NewToggle("notepad-set-mobile-sheets", s.mobilePanelSheets, 0, 0, 52, 28)
	s.setMobileNavTog = ui.NewToggle("notepad-set-mobile-nav", s.mobileBottomNav, 0, 0, 52, 28)
	s.setFindModalTog = ui.NewToggle("notepad-set-find-modal", s.findUseModal, 0, 0, 52, 28)

	appearanceOpts := []string{"System", "Light", "Dark"}
	appearanceDD := ui.NewDropdown("notepad-set-appearance", appearanceOpts, notepadAppearanceIndex(s.appearanceMode), 0, 0, 168, 36)
	s.settingsAppearanceDD = appearanceDD
	s.appearanceIdx = appearanceDD.SelectedIndex
	chromaOpts := notepadChromaStyleOptions()
	chromaDD := ui.NewDropdown("notepad-set-chroma", chromaOpts, notepadChromaStyleIndex(s.chromaStyleName), 0, 0, 168, 36)
	s.chromaStyleIdx = chromaDD.SelectedIndex

	vp.AddChild(settingsSection("notepad-settings-sec-editor", "Editor"))
	vp.AddChild(settingsFlatToggleRow("notepad-set-wrap-row", "Word wrap", "Wrap long lines", ui.PhosphorTextWrap, s.setWrapTog))
	vp.AddChild(settingsFlatToggleRow("notepad-set-spell-row", "Spell check", "Underline misspellings", ui.PhosphorTextUnderline, s.setSpellTog))

	vp.AddChild(settingsSection("notepad-settings-sec-view", "View"))
	vp.AddChild(settingsFlatToggleRow("notepad-set-status-row", "Status bar", "Line and column info", ui.PhosphorInfoI, s.setStatusTog))
	vp.AddChild(settingsFlatToggleRow("notepad-set-session-row", "Open notes panel", "Sidebar for open notes", ui.PhosphorList, s.setSessionTog))
	vp.AddChild(settingsFlatToggleRow("notepad-set-mobile-sheets-row", "Mobile full-width panels", "Full-screen notes or preview", ui.PhosphorResize, s.setMobileSheetsTog))
	vp.AddChild(settingsFlatToggleRow("notepad-set-mobile-nav-row", "Mobile bottom navigation", "Icon bar below 1024px width", ui.PhosphorList, s.setMobileNavTog))
	vp.AddChild(settingsFlatToggleRow("notepad-set-chrome-row", "Ribbon toolbar", "Ribbon on wide screens", ui.PhosphorListSettings, s.setChromeTog))
	vp.AddChild(settingsFlatToggleRow("notepad-set-find-modal-row", "Find dialog", "Modal vs ribbon find", ui.PhosphorMagnifyingGlass, s.setFindModalTog))

	vp.AddChild(settingsSection("notepad-settings-sec-syntax", "Syntax"))
	vp.AddChild(settingsFlatToggleRow("notepad-set-syntax-row", "Syntax highlight", "Colorize source code", ui.PhosphorCodeView, s.setSyntaxTog))
	vp.AddChild(settingsFlatDropdownRow("notepad-set-chroma-row", "Syntax theme", "Chroma editor colors", ui.PhosphorCodeBlock, chromaDD))

	vp.AddChild(settingsSection("notepad-settings-sec-appearance", "Appearance"))
	vp.AddChild(settingsFlatDropdownRow("notepad-set-appearance-row", "Theme", "Light, dark, or system", ui.PhosphorSun, appearanceDD))

	bottomPad := ui.NewLabel("notepad-settings-bottom-pad", "", 0, 0, 0, 96)
	bottomPad.SetStyle("transparent")
	vp.AddChild(bottomPad)

	s.wireSettingsToggles()

	page.AddChild(vp)
	page.Hide()
	return page
}

func (s *notepadScene) wireSettingsToggles() {
	var applying bool
	apply := func() {
		if s.setWrapTog == nil || applying || s.settingsSyncing {
			return
		}
		applying = true
		defer func() { applying = false }()

		prevAppearance := s.appearanceMode
		prevChroma := s.chromaStyleName

		s.wordWrap = s.setWrapTog.Value.Get()
		s.showStatus = s.setStatusTog.Value.Get()
		s.syntaxHighlight = s.setSyntaxTog.Value.Get()
		s.spellCheck = s.setSpellTog.Value.Get()
		s.showSession = s.setSessionTog.Value.Get()
		s.mobilePanelSheets = s.setMobileSheetsTog.Value.Get()
		s.mobileBottomNav = s.setMobileNavTog.Value.Get()
		s.findUseModal = s.setFindModalTog.Value.Get()
		if s.appearanceIdx != nil {
			newAppearance := notepadAppearanceFromIndex(s.appearanceIdx.Get())
			if newAppearance != s.appearanceMode {
				s.appearanceMode = newAppearance
				s.chromaStyleName = notepadDefaultChromaStyle(newAppearance)
				if s.chromaStyleIdx != nil {
					s.chromaStyleIdx.Set(notepadChromaStyleIndex(s.chromaStyleName))
				}
			} else {
				s.appearanceMode = newAppearance
			}
		}
		if s.chromaStyleIdx != nil {
			opts := notepadChromaStyleOptions()
			i := s.chromaStyleIdx.Get()
			if i >= 0 && i < len(opts) {
				s.chromaStyleName = opts[i]
			}
		}
		if s.setChromeTog.Value.Get() {
			s.chromeMode = notepadChromeRibbon
		} else {
			s.chromeMode = notepadChromeMenubar
		}
		if s.editor != nil {
			s.editor.SetWordWrap(s.wordWrap)
		}
		s.toggleStatusBarApply(s.showStatus)
		s.applyEditorSyntax()
		s.applyEditorSpellCheck()
		s.applySessionPanelLayout()
		s.applyResponsivePanelLayout()
		s.applyChromeLayout()
		s.syncRibbonFindMode()
		s.persistQuickPrefs()
		s.refreshChrome()

		themeChanged := s.appearanceMode != prevAppearance || s.chromaStyleName != prevChroma
		if themeChanged && s.uiDoc != nil {
			if s.settingsAppearanceDD != nil && s.settingsAppearanceDD.IsOpen() {
				s.settingsAppearanceDD.ClearOverlayPointerState()
			}
			doc := s.uiDoc
			doc.QueueMain(func() { s.applyNotepadThemePrefs() })
		}
	}
	s.setWrapTog.Value.Subscribe(func() { apply() })
	s.setStatusTog.Value.Subscribe(func() { apply() })
	s.setSyntaxTog.Value.Subscribe(func() { apply() })
	s.setSpellTog.Value.Subscribe(func() { apply() })
	s.setSessionTog.Value.Subscribe(func() { apply() })
	s.setMobileSheetsTog.Value.Subscribe(func() { apply() })
	s.setMobileNavTog.Value.Subscribe(func() { apply() })
	s.setFindModalTog.Value.Subscribe(func() { apply() })
	s.setChromeTog.Value.Subscribe(func() { apply() })
	if s.appearanceIdx != nil {
		s.appearanceIdx.Subscribe(func() { apply() })
	}
	if s.chromaStyleIdx != nil {
		s.chromaStyleIdx.Subscribe(func() { apply() })
	}
}

func (s *notepadScene) syncSettingsToggles() {
	if s.setWrapTog == nil {
		return
	}
	s.settingsSyncing = true
	defer func() { s.settingsSyncing = false }()
	s.setWrapTog.Value.Set(s.wordWrap)
	s.setStatusTog.Value.Set(s.showStatus)
	s.setSyntaxTog.Value.Set(s.syntaxHighlight)
	s.setSpellTog.Value.Set(s.spellCheck)
	s.setSessionTog.Value.Set(s.showSession)
	s.setMobileSheetsTog.Value.Set(s.mobilePanelSheets)
	s.setMobileNavTog.Value.Set(s.mobileBottomNav)
	s.setFindModalTog.Value.Set(s.findUseModal)
	s.setChromeTog.Value.Set(s.chromeMode == notepadChromeRibbon)
	if s.appearanceIdx != nil {
		s.appearanceIdx.Set(notepadAppearanceIndex(s.appearanceMode))
	}
	if s.chromaStyleIdx != nil {
		s.chromaStyleIdx.Set(notepadChromaStyleIndex(s.chromaStyleName))
	}
}

func (s *notepadScene) showSettingsPage() {
	if s.uiDoc == nil || s.settingsPage == nil {
		return
	}
	ui.CloseContextMenu()
	s.syncSettingsToggles()
	s.inSettings = true
	s.syncRibbonFileButtons()
	s.applyChromeLayout()
	s.workArea.Hide()
	s.settingsPage.Show()
	if s.statusBar != nil {
		s.statusBar.Hide()
	}
	if s.settingsPage.Parent() != nil {
		s.settingsPage.Parent().MarkDirty()
	}
}

func (s *notepadScene) closeSettingsPage() {
	if s.settingsPage == nil {
		return
	}
	s.inSettings = false
	s.syncRibbonFileButtons()
	s.settingsPage.Hide()
	s.workArea.Show()
	s.toggleStatusBarApply(s.showStatus)
	s.applyChromeLayout()
	if s.settingsThemePending && s.showPreview {
		s.settingsThemePending = false
		s.rebuildPreview()
	}
	if s.workArea.Parent() != nil {
		s.workArea.Parent().MarkDirty()
	}
	if s.uiDoc != nil && s.editor != nil {
		s.uiDoc.SetFocus(s.editor)
	}
}

func (s *notepadScene) toggleStatusBarApply(show bool) {
	s.showStatus = show
	if s.statusBar == nil || s.inSettings {
		return
	}
	if show {
		s.statusBar.Show()
	} else {
		s.statusBar.Hide()
	}
	s.refreshStatus()
}

func (s *notepadScene) handleSettingsPointer() {
	// Keep File tab selected so ribbon Back stays visible if the user clicked
	// another ribbon tab while settings is open.
	if s.ribbon != nil {
		s.ribbon.SelectRibbonTabForItem("back")
	}
	clearDrops := func() {
		if s.settingsAppearanceDD != nil && s.settingsAppearanceDD.IsOpen() {
			s.settingsAppearanceDD.ClearOverlayPointerState()
		}
	}
	tryBack := func(btn *ui.IconButton) bool {
		if btn == nil || btn.IsHidden() {
			return false
		}
		b := btn.Bounds()
		if b.Width < 1 || b.Height < 1 {
			return false
		}
		mouse := rl.GetMousePosition()
		pending := ui.PointerClickPending()
		pos := ui.PointerClickPosition()
		hit := rl.CheckCollisionPointRec(mouse, b)
		if pending {
			hit = hit || rl.CheckCollisionPointRec(pos, b)
		}
		if !hit {
			return false
		}
		// Prefer latched click (deep idle FPS). Also accept raw press so Back works
		// when ScenePointerBlocked left IconButton OnClick silent.
		consumed := ui.PointerClickConsume(b)
		pressed := rl.IsMouseButtonPressed(rl.MouseLeftButton) && rl.CheckCollisionPointRec(mouse, b)
		if !consumed && !pressed {
			if ui.ScenePointerBlocked() && pending && rl.CheckCollisionPointRec(pos, b) {
				ui.PointerClickMarkUsed()
			} else {
				return false
			}
		}
		clearDrops()
		s.closeSettingsPage()
		return true
	}
	if tryBack(s.settingsBack) {
		return
	}
	_ = tryBack(s.ribbonBack)
}

func (s *notepadScene) handleSettingsKeys() {
	if !s.inSettings {
		return
	}
	if rl.IsKeyPressed(rl.KeyEscape) {
		s.closeSettingsPage()
	}
}
