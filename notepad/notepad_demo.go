// Notepad (Go) is the Phase 1–2 reference app: plain-text editing, native file I/O,
// desktop menu bar, status bar, ui.TextEditor, and a debounced Markdown preview pane.
// Keyboard shortcuts and menu hints are in OnUpdate + menu builders; preview in notepad_preview.go.
package notepad

import (
	"errors"
	"fmt"
	"github.com/ledocorp/gru/examples"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/shawnlondono/gru-notepad/appicon"
	"github.com/ledocorp/gru/ui"
	"github.com/ledocorp/gru/ui/preview"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type notepadScene struct {
	examples.BaseScene
	deferSpellInit       bool
	editor               *ui.TextEditor
	editorVP             *ui.Viewport
	split                *ui.SplitView
	previewVP            *ui.Viewport
	mdPreview            *preview.MarkdownView
	mainSplit            *ui.SplitView
	openNotes            *examples.ListPane
	workArea             *ui.Container
	settingsPage         *ui.Container
	inSettings           bool
	mobilePanelSheets    bool
	mobileBottomNav      bool
	setWrapTog           *ui.Toggle
	setStatusTog         *ui.Toggle
	setSyntaxTog         *ui.Toggle
	setSpellTog          *ui.Toggle
	setSessionTog        *ui.Toggle
	setChromeTog         *ui.Toggle
	setMobileSheetsTog   *ui.Toggle
	setMobileNavTog      *ui.Toggle
	setFindModalTog      *ui.Toggle
	appearanceIdx        *ui.Signal[int]
	chromaStyleIdx       *ui.Signal[int]
	bottomNav            *ui.BottomNavigationBar
	menuBar              *ui.MenuBar
	ribbon               *ui.Toolbar
	ribbonNotesChecked   *ui.Signal[bool]
	ribbonPreviewChecked *ui.Signal[bool]
	ribbonWrapChecked    *ui.Signal[bool]
	ribbonStatusChecked  *ui.Signal[bool]
	ribbonSpellChecked   *ui.Signal[bool]
	ribbonSyntaxChecked  *ui.Signal[bool]
	ribbonRecentBtn      *ui.IconButton
	ribbonLangBtn        *ui.IconButton
	ribbonChromaDD       *ui.Dropdown
	ribbonFindBar        *ui.TextInput
	ribbonTabSearch      int // ribbon ActiveGroup index for Search tab
	chromeMode           string
	uiDoc                *ui.Document
	statusBar            *ui.StatusBar
	store                *NotepadStore
	sess                 *notepadSession
	path                 string
	savedText            string
	dirty                bool
	wordWrap             bool
	syntaxHighlight      bool
	spellCheck           bool
	syntaxLangPick       string // "auto", "plain", or Chroma lexer name (go, markdown, …)
	appearanceMode       string // light, dark, system
	chromaStyleName      string // Chroma styles registry name
	showStatus           bool
	showPreview          bool
	showSession          bool
	previewPending       bool
	previewTimer         float32
	lastPreviewHostW     float32 // soft-reset MarkdownView when host width changes (Notepad only)
	statusDocTitle       *ui.Label
	statusDocState       *ui.Label
	statusLn             *ui.Label
	statusChars          *ui.Label
	statusType           *ui.Label
	statusZoom           *ui.Label
	statusEnc            *ui.Label
	findQuery            string
	replaceWith          string
	findUseModal         bool
	findMatchCase        bool
	noteDisplayName      string // active doc title in open notes (may differ from path)
	persistTimer         float32
	lastStatusDoc        string
	lastStatusState      string
	lastStatusLn         string
	lastStatusChars      string
	lastStatusType       string
	lastStatusZoom       string
	lastStatusEnc        string
}

func (s *notepadScene) Title() string { return SceneTitle }

func (s *notepadScene) TitleBarShowTitle() bool    { return false }
func (s *notepadScene) TitleBarShowAppIcon() bool  { return true }
func (s *notepadScene) TitleBarCenterText() string { return "" }

func (s *notepadScene) HideDemoNav() bool { return true }

func (s *notepadScene) Destroy() {
	if s.store != nil {
		s.persistOpenDocs()
		_ = s.store.Close()
		s.store = nil
	}
}

func notepadStatusLabel(id, text string) *ui.Label {
	l := ui.NewLabel(id, text, 0, 0, 0, 0)
	l.SetStyle("statusbar-label")
	l.Align = ui.LabelAlignLeft
	return l
}

func (s *notepadScene) applyEditorSyntax() {
	if s.editor == nil {
		return
	}
	lang := ""
	if s.syntaxHighlight {
		autoLang := notepadSyntaxLanguage(s.path)
		switch s.syntaxLangPick {
		case "", "auto":
			lang = autoLang
		case "plain":
			lang = ""
		default:
			lang = s.syntaxLangPick
		}
		// .md buffers always use fence-aware markdown highlighting (VS Code-style).
		// A persisted manual SQL/Go pick must not paint prose outside ```lang blocks.
		if autoLang == "markdown" && lang != "" && lang != "markdown" {
			lang = "markdown"
		}
	}
	s.editor.SetSyntaxLanguage(lang)
	s.editor.FlushSyntaxHighlight()
}

func (s *notepadScene) setSyntaxLangPick(pick string) {
	s.syntaxLangPick = pick
	s.applyEditorSyntax()
	s.refreshStatus()
}

func (s *notepadScene) applyEditorSpellCheck() {
	if s.editor == nil {
		return
	}
	if s.spellCheck {
		s.editor.SetSpellChecker(NotepadSpellChecker())
	} else {
		s.editor.SetSpellChecker(nil)
	}
	s.editor.SetSpellCheckEnabled(s.spellCheck)
	s.editor.SetSpellAutoCorrect(s.spellCheck)
	s.editor.SetSpellAutoCorrectTable(notepadSpellAutoCorrect)
	if s.spellCheck {
		s.editor.FlushSpellCheck()
	}
}

func (s *notepadScene) toggleSpellCheck() {
	s.spellCheck = !s.spellCheck
	s.applyEditorSpellCheck()
	s.syncRibbonToggleState()
	s.persistQuickPrefs()
}

func (s *notepadScene) Build(doc *ui.Document) {
	s.uiDoc = doc
	// Product host: pack Notepad icons once (avoid the full demo WarmGlyph set).
	s.preloadNotepadRibbonIcons()
	s.store, _ = OpenNotepadStore()
	prefs := defaultNotepadPrefs()
	if s.store != nil {
		prefs = s.store.LoadPrefs()
	}
	s.sess = newNotepadSession()
	sessionRestored := s.tryRestoreOpenDocs()
	if !sessionRestored {
		// Empty DB / cleared session — install welcome pack in-memory (also persisted).
		if s.store != nil {
			_ = s.store.EnsureWelcomePack()
			sessionRestored = s.tryRestoreOpenDocs()
		}
	}
	if !sessionRestored {
		s.path = ""
		s.savedText = ""
		s.dirty = false
	}
	s.wordWrap = prefs.WordWrap
	s.syntaxHighlight = prefs.SyntaxHighlight
	s.spellCheck = prefs.SpellCheck
	s.syntaxLangPick = prefs.SyntaxLangPick
	s.showStatus = prefs.ShowStatus
	if !sessionRestored {
		// Session restore already applied active.ShowPreview; new sessions start preview off.
		s.showPreview = false
	}
	s.showSession = prefs.ShowSessionPanel
	s.mobilePanelSheets = prefs.MobilePanelSheets
	s.mobileBottomNav = prefs.MobileBottomNav
	s.chromeMode = prefs.ChromeMode
	if s.chromeMode == "" {
		s.chromeMode = notepadChromeMenubar
	}
	s.appearanceMode = prefs.AppearanceMode
	s.chromaStyleName = prefs.ChromaStyle
	s.findUseModal = prefs.FindUseModal
	s.applyNotepadThemePrefs()

	root := examples.MountEdgeToEdgeRoot(doc, "notepad", false)
	root.Gap = 0

	s.menuBar = ui.NewMenuBar("notepad-menubar", []ui.MenuBarMenu{
		{Label: "File", ItemsFunc: s.fileMenuItems},
		{Label: "Edit", ItemsFunc: s.editMenuItems},
		{Label: "View", ItemsFunc: s.viewMenuItems},
		{Label: "Syntax", ItemsFunc: s.syntaxMenuItems},
	}, 0, 0, 0, 0)
	s.ribbon = s.buildNotepadRibbon()
	s.ribbon.Hide()

	body := ui.NewContainer("notepad-body", 0, 0, 0, 0)
	body.LayoutType = ui.LayoutFlex
	body.FlexDirection = ui.FlexColumn
	body.SetStyle("transparent")
	body.SetFlexGrow(1)
	body.Gap = 0

	editorVP := ui.NewViewport("notepad-editor-vp", 0, 0, 0, 0)
	editorVP.WheelScrollStep = 80 // ~2× default; closer to Notepad fast wheel
	editorVP.SetFlexGrow(1)
	editorVP.SetStyle("editor-scroll")
	editorVP.Gap = 0
	s.editorVP = editorVP

	s.editor = ui.NewTextEditor("notepad-editor", "", 0, 0, 0, 0)
	s.editor.SetDocument(doc)
	edStyle := ui.GetThemeStyle("text-editor")
	edStyle.Mono = true
	edStyle.FontSize = 17
	edStyle.MinFontSize = 17
	s.editor.SetStyleOverrides(edStyle)
	s.editor.Placeholder = "Type or open a file. Syntax menu: Highlight + language (Auto uses extension)."
	s.editor.AutoHeight = true
	s.editor.WordWrap = false
	s.editor.OnChange = func() {
		if s.editor == nil {
			return
		}
		was := s.dirty
		s.dirty = s.editor.Text.Get() != s.savedText
		if s.sess != nil && s.sess.active() != nil {
			s.sess.active().Text = s.editor.Text.Get()
		}
		if was != s.dirty {
			s.refreshChrome()
			s.rebuildSessionList()
		} else {
			s.refreshStatus()
		}
		s.schedulePreview()
		s.persistTimer = 1.5
	}
	if prefs.DefaultZoom > 0 {
		z := prefs.DefaultZoom
		if z > 0.96 && z < 1.04 {
			z = 1
		}
		s.editor.SetZoom(z)
	}
	s.editor.SetWordWrap(s.wordWrap)
	editorVP.AddChild(s.editor)

	s.previewVP = ui.NewViewport("notepad-preview-vp", 0, 0, 0, 0)
	s.previewVP.SetStyle("editor-scroll")
	s.previewVP.Gap = 0
	s.previewVP.ContentClipBleed = 0

	s.mdPreview = preview.NewMarkdownView("notepad-md-preview", 0, 0, 0, 0)
	s.mdPreview.SetDocument(doc)
	s.mdPreview.SetFlexGrow(1)
	s.mdPreview.Hide()
	s.previewVP.AddChild(s.mdPreview)

	s.split = ui.NewSplitView("notepad-split", ui.SplitHorizontal, editorVP, s.previewVP, 0, 0, 0, 0)
	s.split.SetFlexGrow(1)
	s.split.MinSecond = 120

	s.openNotes = examples.NewListPane(doc, examples.ListPaneOptions{
		ID:             "notepad-open-notes",
		Title:          "Open Notes",
		PreferredWidth: 260,
		MinWidth:       180,
		ShowCollapse:   true,
		OnCollapse:     func() { s.toggleSessionPanel() },
	})

	editorArea := ui.NewContainer("notepad-editor-area", 0, 0, 0, 0)
	editorArea.LayoutType = ui.LayoutFlex
	editorArea.FlexDirection = ui.FlexColumn
	editorArea.SetStyle("transparent")
	editorArea.SetFlexGrow(1)
	editorArea.Gap = 0
	editorArea.AddChild(s.split)

	s.mainSplit = ui.NewSplitView("notepad-main-split", ui.SplitHorizontal, s.openNotes.Root, editorArea, 0, 0, 0, 0)
	s.mainSplit.SetFlexGrow(1)
	s.mainSplit.MinFirst = 160
	s.mainSplit.MinSecond = 200

	s.workArea = ui.NewContainer("notepad-work", 0, 0, 0, 0)
	s.workArea.LayoutType = ui.LayoutFlex
	s.workArea.FlexDirection = ui.FlexColumn
	s.workArea.SetStyle("transparent")
	s.workArea.SetFlexGrow(1)
	s.workArea.AddChild(s.mainSplit)

	s.settingsPage = s.buildNotepadSettingsPage()

	body.AddChild(s.workArea)
	body.AddChild(s.settingsPage)

	s.statusDocTitle = notepadStatusLabel("notepad-status-doc", "Untitled")
	s.statusDocState = notepadStatusLabel("notepad-status-state", "")
	s.statusLn = notepadStatusLabel("notepad-status-ln", "Ln 1, Col 1")
	s.statusChars = notepadStatusLabel("notepad-status-chars", "0 characters")
	s.statusType = notepadStatusLabel("notepad-status-type", "Plain Text")
	s.statusZoom = notepadStatusLabel("notepad-status-zoom", "100%")
	s.statusEnc = notepadStatusLabel("notepad-status-enc", "UTF-8")

	s.statusBar = ui.NewStatusBar("notepad-status", 0, 0, 0, 0)
	s.statusBar.SetColumns([]ui.StatusBarColumn{
		{Weight: 28, Align: ui.LabelAlignLeft, Nodes: []ui.Node{s.statusDocTitle}},
		{Weight: 10, Align: ui.LabelAlignLeft, Nodes: []ui.Node{s.statusDocState}},
		{Weight: 16, Align: ui.LabelAlignLeft, Nodes: []ui.Node{s.statusLn}},
		{Weight: 14, Align: ui.LabelAlignLeft, Nodes: []ui.Node{s.statusChars}},
		{Weight: 12, Align: ui.LabelAlignLeft, Nodes: []ui.Node{s.statusType}},
		{Weight: 10, Align: ui.LabelAlignRight, Nodes: []ui.Node{s.statusZoom}},
		{Weight: 10, Align: ui.LabelAlignRight, Nodes: []ui.Node{s.statusEnc}},
	})

	root.AddChild(s.menuBar)
	root.AddChild(s.ribbon)
	root.AddChild(body)
	root.AddChild(s.statusBar)
	s.bottomNav = s.buildMobileBottomNav()
	root.AddChild(s.bottomNav)
	s.wireChromeResponsive()
	s.applyChromeLayout()
	root.MarkDirty()

	if sessionRestored && s.sess != nil {
		s.sess.loadActive(s)
		if d := s.sess.active(); d != nil {
			s.dirty = d.isDirty()
		}
		s.ensureOpenNotesVisible()
	}
	doc.SetFocus(s.editor)
	s.applyEditorSyntax()
	if s.spellCheck {
		s.deferSpellInit = true
	} else {
		s.applyEditorSpellCheck()
	}
	s.applyPreviewLayout()
	s.applySessionPanelLayout()
	s.applyResponsivePanelLayout()
	s.rebuildSessionList()
	s.refreshChrome()
	if !s.showStatus {
		s.toggleStatusBarApply(false)
	}
	examples.FinishShellMount(doc)
	// Host may Resize after Build; soft-reset preview once widths are real
	// (thin host mirrors FinishShellMount after Resize).
	if s.showPreview && s.mdPreview != nil {
		s.rebuildPreview()
		s.mdPreview.SimulateResizeReflow()
	}
	// After first frame: load Inter (preview) + mono, then warm inactive MD parse caches.
	// Avoids paying font load on the first Markdown Preview toggle.
	doc.QueueMain(func() {
		ui.EnsurePreviewUIFont(0)
		ui.EnsureUIFontAux(0)
		s.warmInactiveMarkdownCaches()
	})
	SetInstanceOpenFileHandler(func(path string) {
		if s.uiDoc != nil {
			openPath := path
			s.uiDoc.QueueMain(func() { s.doOpen(openPath) })
		}
	})
	if path := examples.StartupOpenFilePath(); path != "" {
		openPath := path
		doc.QueueMain(func() { s.doOpen(openPath) })
	}
}

func (s *notepadScene) fileMenuItems() []ui.ContextMenuItem {
	items := []ui.ContextMenuItem{
		{Label: "New", Shortcut: "Ctrl+N", Action: func() { s.doNew() }},
		{Label: "Open…", Shortcut: "Ctrl+O", Action: func() { s.pickOpen() }},
		{Label: "Save", Shortcut: "Ctrl+S", Action: func() { s.doSave(false) }},
		{Label: "Save As…", Shortcut: "Ctrl+Shift+S", Action: func() { s.doSave(true) }},
		{Label: "Save All", Action: func() { s.doSaveAll() }},
		{Divider: true},
		{Label: "Close Note", Action: func() { s.doCloseNote() }},
		{Divider: true},
	}
	if s.inSettings {
		items = append(items, ui.ContextMenuItem{
			Label:  "Back to editor",
			Action: func() { s.closeSettingsPage() },
		}, ui.ContextMenuItem{Divider: true})
	}
	recent := s.recentMenuItems()
	if len(recent) > 0 {
		items = append(items, ui.ContextMenuItem{Label: "Recent", SubItems: recent})
		items = append(items, ui.ContextMenuItem{Divider: true})
	}
	items = append(items,
		ui.ContextMenuItem{Label: "Settings", Action: func() { s.showSettingsPage() }},
		ui.ContextMenuItem{Divider: true},
		ui.ContextMenuItem{Label: "Exit", Action: func() { rl.CloseWindow() }},
	)
	return items
}

func (s *notepadScene) recentMenuItems() []ui.ContextMenuItem {
	if s.store == nil {
		return nil
	}
	paths, err := s.store.RecentFiles()
	if err != nil || len(paths) == 0 {
		return nil
	}
	out := make([]ui.ContextMenuItem, 0, len(paths))
	for _, p := range paths {
		path := p
		label := filepath.Base(path)
		if label == "" {
			label = path
		}
		out = append(out, ui.ContextMenuItem{
			Label: label,
			Action: func() {
				s.doOpen(path)
			},
		})
	}
	return out
}

func (s *notepadScene) doNew() {
	if s.editor == nil || s.sess == nil {
		return
	}
	s.sess.addUntitled(s)
}

func (s *notepadScene) doCloseNote() {
	if s.sess == nil {
		return
	}
	if len(s.sess.docs) <= 1 {
		ui.ShowToast("Cannot close the only open note", ui.ToastInfo, 2*time.Second)
		return
	}
	if s.dirty {
		ui.ShowToast("Save or discard changes before closing", ui.ToastWarning, 2*time.Second)
		return
	}
	if s.sess.removeActive(s) {
		s.persistOpenDocs()
	}
}

func (s *notepadScene) pickOpen() {
	path, err := examples.PickTextFileOpen()
	if errors.Is(err, examples.ErrFileDialogCancelled) || errors.Is(err, examples.ErrFileDialogUnavailable) {
		return
	}
	if err != nil {
		ui.ShowToast("Open failed: "+err.Error(), ui.ToastError, 3*time.Second)
		return
	}
	// Defer past the native dialog return so the next frame drains layout before open.
	if s.uiDoc != nil {
		openPath := path
		s.uiDoc.QueueMain(func() { s.doOpen(openPath) })
		return
	}
	s.doOpen(path)
}

func notepadNormalizeText(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	return s
}

func (s *notepadScene) doOpen(path string) {
	path = filepath.Clean(path)
	if s.sess != nil {
		if existing := s.sess.findByPath(path); existing != nil {
			s.activateOpenDoc(existing.ID)
			if s.store != nil {
				_ = s.store.AddRecent(path)
			}
			return
		}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		ui.ShowToast("Open failed: "+err.Error(), ui.ToastError, 3*time.Second)
		return
	}
	if !utf8.Valid(data) {
		ui.ShowToast("File is not valid UTF-8 text", ui.ToastWarning, 3*time.Second)
		return
	}
	if s.store != nil {
		_ = s.store.AddRecent(path)
	}
	s.openDocument(path, notepadNormalizeText(string(data)))
}

func (s *notepadScene) openDocument(path, text string) {
	if s.sess == nil {
		return
	}
	path = filepath.Clean(path)
	if existing := s.sess.findByPath(path); existing != nil {
		s.activateOpenDoc(existing.ID)
		return
	}
	s.sess.flushActive(s)
	if s.shouldOpenInNewDoc() {
		d := s.sess.newDoc()
		s.sess.docs = append(s.sess.docs, d)
		s.sess.activeID = d.ID
	}
	d := s.sess.active()
	if d == nil {
		return
	}
	d.Path = path
	d.Text = text
	d.SavedText = text
	d.ShowPreview = false
	s.sess.loadActive(s)
	s.syntaxLangPick = "auto"
	s.applyEditorSyntax()
	s.applyEditorSpellCheck()
	s.ensureOpenNotesVisible()
	s.applySessionPanelLayout()
	s.applyPreviewLayout()
	s.rebuildSessionList()
	s.refreshOpenNotesChrome()
	s.refreshChrome()
	s.persistOpenDocs()
}

func (s *notepadScene) shouldOpenInNewDoc() bool {
	if s.sess == nil {
		return false
	}
	d := s.sess.active()
	if d == nil {
		return false
	}
	if d.Path != "" {
		return true
	}
	return strings.TrimSpace(d.Text) != ""
}

func (s *notepadScene) doSave(saveAs bool) {
	if s.editor == nil {
		return
	}
	if s.sess != nil {
		s.sess.flushActive(s)
	}
	if s.path == "" || saveAs {
		path, err := examples.PickTextFileSave(s.saveDefaultFilename(saveAs))
		if errors.Is(err, examples.ErrFileDialogCancelled) || errors.Is(err, examples.ErrFileDialogUnavailable) {
			return
		}
		if err != nil {
			ui.ShowToast("Save failed: "+err.Error(), ui.ToastError, 3*time.Second)
			return
		}
		s.path = path
	}
	s.writeFile()
}

func (s *notepadScene) writeFile() {
	if s.editor == nil || s.path == "" {
		return
	}
	if s.sess != nil {
		s.sess.flushActive(s)
	}
	text := s.editor.Text.Get()
	if err := os.WriteFile(s.path, []byte(text), 0o644); err != nil {
		ui.ShowToast("Save failed: "+err.Error(), ui.ToastError, 3*time.Second)
		return
	}
	s.syncDisplayNameAfterSave(s.path)
	s.savedText = text
	s.dirty = false
	if s.sess != nil && s.sess.active() != nil {
		d := s.sess.active()
		d.SavedText = text
		d.Text = text
	}
	if s.store != nil {
		_ = s.store.AddRecent(s.path)
	}
	s.refreshChrome()
	s.rebuildSessionList()
	s.applyEditorSyntax()
	s.applyEditorSpellCheck()
}

func (s *notepadScene) refreshChrome() {
	s.syncRibbonToggleState()
	s.syncRibbonChromaTheme()
	s.syncRibbonFileButtons()
	s.syncRibbonFindMode()
	s.refreshStatus()
	rl.SetWindowTitle(s.windowTitle())
}

func (s *notepadScene) windowTitle() string {
	name := s.activeDocTitle()
	if s.dirty {
		name += " *"
	}
	return fmt.Sprintf("Gru — %s — %s", SceneTitle, name)
}

func (s *notepadScene) statusDocSuffix() string {
	switch {
	case s.dirty && s.path == "":
		return "Unsaved"
	case s.dirty:
		return "Modified"
	default:
		return ""
	}
}

func (s *notepadScene) refreshStatus() {
	title := s.activeDocTitle()
	if title != s.lastStatusDoc {
		s.lastStatusDoc = title
		s.statusDocTitle.Text.Set(title)
	}
	state := s.statusDocSuffix()
	if state != s.lastStatusState {
		s.lastStatusState = state
		s.statusDocState.Text.Set(state)
	}

	text := ""
	if s.editor != nil {
		text = s.editor.Text.Get()
	}
	line, col := 1, 1
	if s.editor != nil {
		line, col = s.editor.CursorLineCol()
	}
	lineCol := fmt.Sprintf("Ln %d, Col %d", line, col)
	if lineCol != s.lastStatusLn {
		s.lastStatusLn = lineCol
		s.statusLn.Text.Set(lineCol)
	}
	charLine := fmt.Sprintf("%d characters", utf8.RuneCountInString(text))
	if charLine != s.lastStatusChars {
		s.lastStatusChars = charLine
		s.statusChars.Text.Set(charLine)
	}
	typeLine := notepadDocType(s.path, s.editorSyntaxLang())
	if typeLine != s.lastStatusType {
		s.lastStatusType = typeLine
		s.statusType.Text.Set(typeLine)
	}
	zoom := float32(1)
	if s.editor != nil && s.editor.Zoom > 0 {
		zoom = s.editor.Zoom
	}
	zoomLine := fmt.Sprintf("%.0f%%", zoom*100)
	if zoomLine != s.lastStatusZoom {
		s.lastStatusZoom = zoomLine
		s.statusZoom.Text.Set(zoomLine)
	}
	if s.lastStatusEnc != "UTF-8" {
		s.lastStatusEnc = "UTF-8"
		s.statusEnc.Text.Set("UTF-8")
	}
}

func (s *notepadScene) editMenuItems() []ui.ContextMenuItem {
	ed := s.editor
	hasSel := ed != nil && ed.HasSelection()
	canPaste := ed != nil && normalizeEditorClipboardForMenu(rl.GetClipboardText()) != ""
	return []ui.ContextMenuItem{
		{Label: "Undo", Shortcut: "Ctrl+Z", Disabled: ed == nil || !ed.CanUndo(), Action: func() {
			if ed != nil {
				ed.Undo()
			}
		}},
		{Label: "Redo", Shortcut: "Ctrl+Y", Disabled: ed == nil || !ed.CanRedo(), Action: func() {
			if ed != nil {
				ed.Redo()
			}
		}},
		{Divider: true},
		{Label: "Cut", Shortcut: "Ctrl+X", Disabled: !hasSel, Action: func() {
			if ed != nil {
				ed.Cut()
			}
		}},
		{Label: "Copy", Shortcut: "Ctrl+C", Disabled: ed == nil || (!hasSel && ed.Text.Get() == ""), Action: func() {
			if ed != nil {
				ed.Copy()
			}
		}},
		{Label: "Paste", Shortcut: "Ctrl+V", Disabled: !canPaste, Action: func() {
			if ed != nil {
				ed.Paste()
			}
		}},
		{Label: "Delete", Shortcut: "Del", Disabled: !hasSel, Action: func() {
			if ed != nil {
				ed.Delete()
			}
		}},
		{Divider: true},
		{Label: "Select All", Shortcut: "Ctrl+A", Action: func() {
			if ed != nil {
				ed.SelectAll()
			}
		}},
		{Divider: true},
		{Label: "Spell Check", Checked: s.spellCheck, Action: func() { s.toggleSpellCheck() }},
		{Divider: true},
		{Label: "Find…", Shortcut: "Ctrl+F", Action: func() { s.showFindDialog() }},
		{Label: "Find Next", Shortcut: "F3", Action: func() { s.findNextOrPrompt() }},
		{Label: "Find Previous", Shortcut: "Shift+F3", Action: func() { s.findPreviousOrPrompt() }},
		{Label: "Replace…", Shortcut: "Ctrl+H", Action: func() { s.showReplaceDialog() }},
		{Label: "Go To…", Shortcut: "Ctrl+G", Action: func() { s.showGoToDialog() }},
	}
}

func normalizeEditorClipboardForMenu(s string) string {
	s = strings.TrimRight(s, "\x00")
	return s
}

func (s *notepadScene) syntaxMenuItems() []ui.ContextMenuItem {
	items := []ui.ContextMenuItem{
		{Label: "Highlight", Checked: s.syntaxHighlight, Action: func() { s.toggleSyntaxHighlight() }},
		{Divider: true},
	}
	items = append(items, s.syntaxLanguageMenuItems(true)...)
	return items
}

// syntaxLangMenuItems is the language submenu for ribbon (checked marks, like menubar).
func (s *notepadScene) syntaxLangMenuItems() []ui.ContextMenuItem {
	return s.syntaxLanguageMenuItems(true)
}

func (s *notepadScene) viewMenuItems() []ui.ContextMenuItem {
	items := []ui.ContextMenuItem{
		{Label: "Open Notes", Checked: s.showSession, Action: func() { s.toggleSessionPanel() }},
		{Divider: true},
		{Label: "Word Wrap", Checked: s.wordWrap, Action: func() { s.toggleWordWrap() }},
		{Label: "Status Bar", Checked: s.showStatus, Action: func() { s.toggleStatusBar() }},
		{Label: "Markdown Preview", Checked: s.showPreview, Action: func() { s.toggleMarkdownPreview() }},
	}
	items = append(items,
		ui.ContextMenuItem{Divider: true},
		ui.ContextMenuItem{Label: "Zoom In", Shortcut: "Ctrl++", Action: func() { s.zoomBy(0.1) }},
		ui.ContextMenuItem{Label: "Zoom Out", Shortcut: "Ctrl+-", Action: func() { s.zoomBy(-0.1) }},
		ui.ContextMenuItem{Label: "Reset Zoom", Shortcut: "Ctrl+0", Action: func() { s.setZoom(1) }},
	)
	return items
}

func (s *notepadScene) toggleWordWrap() {
	s.wordWrap = !s.wordWrap
	if s.editor != nil {
		s.editor.SetWordWrap(s.wordWrap)
	}
	s.syncRibbonToggleState()
	s.persistQuickPrefs()
}

func (s *notepadScene) editorZoom() float32 {
	if s.editor != nil && s.editor.Zoom > 0 {
		return s.editor.Zoom
	}
	return 1
}

func (s *notepadScene) persistQuickPrefs() {
	if s.store == nil {
		return
	}
	p := s.store.LoadPrefs()
	p.WordWrap = s.wordWrap
	p.ShowStatus = s.showStatus
	p.SyntaxHighlight = s.syntaxHighlight
	p.SpellCheck = s.spellCheck
	p.SyntaxLangPick = s.syntaxLangPick
	p.DefaultZoom = s.editorZoom()
	p.ShowSessionPanel = s.showSession
	p.MobilePanelSheets = s.mobilePanelSheets
	p.MobileBottomNav = s.mobileBottomNav
	p.ChromeMode = s.chromeMode
	p.AppearanceMode = s.appearanceMode
	p.ChromaStyle = s.chromaStyleName
	p.FindUseModal = s.findUseModal
	p.EditorBackground = ""
	_ = s.store.SavePrefs(p)
}

func notepadAppearanceIndex(mode string) int {
	switch ui.AppearanceMode(mode) {
	case ui.AppearanceDark:
		return 2
	case ui.AppearanceLight:
		return 1
	default:
		return 0
	}
}

func notepadAppearanceFromIndex(i int) string {
	switch i {
	case 2:
		return ui.AppearanceDark
	case 1:
		return ui.AppearanceLight
	default:
		return ui.AppearanceSystem
	}
}

func notepadDefaultChromaStyle(appearanceMode string) string {
	if ui.EffectiveAppearance(appearanceMode) == ui.AppearanceDark {
		return "catppuccin-mocha"
	}
	return "github"
}

func notepadChromaStyleOptions() []string {
	return ui.CuratedChromaStyles
}

func notepadChromaStyleIndex(name string) int {
	opts := notepadChromaStyleOptions()
	name = strings.TrimSpace(name)
	for i, o := range opts {
		if o == name {
			return i
		}
	}
	effective := ui.ChromaStyleName()
	for i, o := range opts {
		if o == effective {
			return i
		}
	}
	return 0
}

func (s *notepadScene) syncRibbonChromaTheme() {
	if s.ribbonChromaDD == nil {
		return
	}
	name := s.chromaStyleName
	if name == "" {
		name = ui.ChromaStyleName()
	}
	idx := notepadChromaStyleIndex(name)
	if s.ribbonChromaDD.SelectedIndex.Get() != idx {
		s.ribbonChromaDD.SelectedIndex.Set(idx)
	}
}

func (s *notepadScene) applyEditorChromeStyles() {
	if s.editor == nil {
		return
	}
	edStyle := ui.GetThemeStyle("text-editor")
	edStyle.Mono = true
	edStyle.FontSize = 17
	edStyle.MinFontSize = 17
	paper := edStyle.BackgroundColor
	s.editor.SetStyleOverrides(edStyle)
	s.editor.MarkDrawDirty()
	if s.editorVP != nil {
		scrollStyle := ui.GetThemeStyle("editor-scroll")
		scrollStyle.BackgroundColor = paper
		s.editorVP.SetStyleOverrides(scrollStyle)
		s.editorVP.InvalidateResolvedStyle()
		s.editorVP.MarkDrawDirty()
	}
	if s.previewVP != nil {
		s.previewVP.MarkDrawDirty()
	}
	if s.split != nil {
		s.split.MarkDrawDirty()
	}
}

func (s *notepadScene) applyNotepadThemePrefs() {
	ui.SetAppearance(s.appearanceMode)
	appicon.SetPreferDarkIcon(ui.ThemeIsDark())
	if rl.IsWindowReady() {
		appicon.ReapplyWindowIcon()
	}
	ui.SetChromaStyle(s.chromaStyleName)
	s.applyEditorChromeStyles()
	if s.uiDoc != nil {
		ui.MarkTreeThemeDirty(s.uiDoc.Root)
		s.uiDoc.UnloadCache()
		if s.uiDoc.Root != nil {
			s.uiDoc.Root.MarkDirty()
		}
	}
	if s.ribbon != nil {
		s.ribbon.MarkDirty()
	}
	if s.menuBar != nil {
		s.menuBar.MarkDirty()
	}
	s.applyEditorSyntax()
	if s.showPreview {
		s.rebuildPreview()
	}
	s.syncRibbonChromaTheme()
}

func (s *notepadScene) toggleSyntaxHighlight() {
	s.syntaxHighlight = !s.syntaxHighlight
	s.applyEditorSyntax()
	s.syncRibbonToggleState()
	s.refreshStatus()
}

func (s *notepadScene) editorSyntaxLang() string {
	if s.editor == nil || !s.syntaxHighlight {
		return ""
	}
	return s.editor.SyntaxLanguage()
}

func (s *notepadScene) toggleStatusBar() {
	s.showStatus = !s.showStatus
	s.toggleStatusBarApply(s.showStatus)
	s.syncRibbonToggleState()
	s.persistQuickPrefs()
}

func (s *notepadScene) zoomBy(delta float32) {
	if s.editor == nil {
		return
	}
	z := s.editor.Zoom + delta
	s.setZoom(z)
}

func (s *notepadScene) setZoom(z float32) {
	if s.editor == nil {
		return
	}
	// Snap near-100% so prefs don't keep a sticky 1.05 / 0.95 after Ctrl+0.
	if z > 0.95 && z < 1.05 {
		z = 1
	}
	s.editor.SetZoom(z)
	s.refreshStatus()
	s.persistQuickPrefs()
}

func notepadDocType(path, syntaxLang string) string {
	if syntaxLang != "" {
		switch syntaxLang {
		case "go":
			return "Go"
		case "markdown":
			return "Markdown"
		case "python":
			return "Python"
		case "rust":
			return "Rust"
		case "javascript":
			return "JavaScript"
		case "typescript":
			return "TypeScript"
		case "json":
			return "JSON"
		case "yaml":
			return "YAML"
		case "sql":
			return "SQL"
		case "html":
			return "HTML"
		case "css":
			return "CSS"
		case "bash":
			return "Bash"
		case "c":
			return "C"
		case "cpp":
			return "C++"
		case "java":
			return "Java"
		case "lua":
			return "Lua"
		default:
			return strings.ToUpper(syntaxLang[:1]) + syntaxLang[1:]
		}
	}
	switch filepath.Ext(path) {
	case ".md", ".MD":
		return "Markdown"
	default:
		return "Plain Text"
	}
}

func (s *notepadScene) OnUpdate(doc *ui.Document, dt float32) {
	if s.editor == nil {
		return
	}
	if s.deferSpellInit {
		s.deferSpellInit = false
		s.applyEditorSpellCheck()
	}
	s.handleSettingsKeys()
	if s.inSettings {
		return
	}
	s.tickPreview(dt)
	s.softResetPreviewOnHostResize(doc)
	if s.persistTimer > 0 {
		s.persistTimer -= dt
		if s.persistTimer <= 0 {
			s.persistTimer = 0
			s.persistOpenDocs()
		}
	}
	s.updateMobileBottomNav(doc)
	s.handleNotepadPointer(doc)
	s.handleNotepadSearchKeys()
	if ui.IsModalOpen() {
		s.refreshStatus()
		return
	}
	if keyChordCtrl(rl.KeyO) {
		s.pickOpen()
	} else if keyChordCtrl(rl.KeyS) {
		if rl.IsKeyDown(rl.KeyLeftShift) || rl.IsKeyDown(rl.KeyRightShift) {
			s.doSave(true)
		} else {
			s.doSave(false)
		}
	} else if keyChordCtrl(rl.KeyN) {
		s.doNew()
	} else if keyChordCtrl(rl.KeyEqual) || keyChordCtrl(rl.KeyKpAdd) {
		s.zoomBy(0.1)
	} else if keyChordCtrl(rl.KeyMinus) || keyChordCtrl(rl.KeyKpSubtract) {
		s.zoomBy(-0.1)
	} else if keyChordCtrl(rl.KeyZero) || keyChordCtrl(rl.KeyKp0) {
		s.setZoom(1)
	}
	s.refreshStatus()
}

func keyChordCtrl(key int32) bool {
	ctrl := rl.IsKeyDown(rl.KeyLeftControl) || rl.IsKeyDown(rl.KeyRightControl)
	return ctrl && rl.IsKeyPressed(key)
}

// handleNotepadPointer focuses the editor when clicking the edit surface.
// Caret placement is handled by TextEditor.updateMouseSelection (do not intercept clicks).
func (s *notepadScene) handleNotepadPointer(doc *ui.Document) {
	if s.inSettings {
		return
	}
	if !rl.IsWindowFocused() || ui.IsContextMenuOpen() || ui.OverlayBlocksSceneInput() ||
		ui.ScenePointerBlocked() {
		return
	}
	mouse := rl.GetMousePosition()
	if s.bottomNav != nil && !s.bottomNav.IsHidden() &&
		rl.CheckCollisionPointRec(mouse, s.bottomNav.Bounds()) {
		return
	}
	if s.useMobileBottomNav() && mouse.Y >= float32(doc.Height)-shellFooterLift {
		return
	}
	if s.editorVP != nil && !s.editorVP.IsHidden() &&
		rl.CheckCollisionPointRec(mouse, s.editorVP.Bounds()) {
		if s.editor != nil && !s.editor.Disabled {
			if rl.IsMouseButtonPressed(rl.MouseRightButton) {
				doc.SetFocus(s.editor)
				s.editor.ShowContextMenuAt(mouse)
			} else if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
				if doc.Focused != s.editor {
					doc.SetFocus(s.editor)
				}
				s.editor.RestartCaretCycle()
			}
		}
		return
	}
	if rl.IsMouseButtonPressed(rl.MouseLeftButton) && s.editor != nil && s.editor.IsFocused() {
		s.editor.ClearSelection()
	}
	if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		focusEditableAt(doc)
	}
}
