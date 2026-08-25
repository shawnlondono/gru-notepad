// Gru Notepad — multi-document session and open-notes sidebar.
package notepad

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/ledocorp/gru/ui"
)

type notepadDoc struct {
	ID          string
	Path        string
	DisplayName string // open-notes title before/after save (optional)
	Text        string
	SavedText   string
	ShowPreview bool
}

type notepadSession struct {
	docs     []*notepadDoc
	activeID string
	nextNum  int
}

func newNotepadSession() *notepadSession {
	s := &notepadSession{nextNum: 1}
	d := s.newDoc()
	s.docs = []*notepadDoc{d}
	s.activeID = d.ID
	return s
}

func (sess *notepadSession) newDoc() *notepadDoc {
	id := fmt.Sprintf("doc-%d", sess.nextNum)
	sess.nextNum++
	return &notepadDoc{ID: id}
}

func (sess *notepadSession) active() *notepadDoc {
	for _, d := range sess.docs {
		if d.ID == sess.activeID {
			return d
		}
	}
	if len(sess.docs) > 0 {
		return sess.docs[0]
	}
	return nil
}

func (sess *notepadSession) find(id string) *notepadDoc {
	for _, d := range sess.docs {
		if d.ID == id {
			return d
		}
	}
	return nil
}

// findByPath returns an open note for path, if already in the session.
func (sess *notepadSession) findByPath(path string) *notepadDoc {
	if sess == nil || path == "" {
		return nil
	}
	for _, d := range sess.docs {
		if d.Path != "" && notepadPathsEqual(d.Path, path) {
			return d
		}
	}
	return nil
}

func notepadPathsEqual(a, b string) bool {
	a = filepath.Clean(a)
	b = filepath.Clean(b)
	if runtime.GOOS == "windows" {
		return strings.EqualFold(a, b)
	}
	return a == b
}

// activateOpenDoc focuses an already-open note (Recent, Open dedupe).
func (s *notepadScene) activateOpenDoc(id string) {
	if s.sess == nil || s.sess.find(id) == nil {
		return
	}
	if s.sess.activeID != id {
		s.sess.switchTo(s, id)
	}
	if s.useMobilePanelSheets() && s.showSession {
		s.dismissOpenNotesAfterPick()
	} else {
		s.focusEditor()
	}
}

// flushActive copies the live editor state into the active document.
func (sess *notepadSession) flushActive(s *notepadScene) {
	if sess == nil || s == nil || s.editor == nil {
		return
	}
	d := sess.active()
	if d == nil {
		return
	}
	d.Text = s.editor.Text.Get()
	d.SavedText = s.savedText
	d.Path = s.path
	d.DisplayName = s.noteDisplayName
	d.ShowPreview = s.showPreview
}

func (sess *notepadSession) loadActive(s *notepadScene) {
	d := sess.active()
	if d == nil || s == nil {
		return
	}
	s.path = d.Path
	s.savedText = d.SavedText
	s.noteDisplayName = d.DisplayName
	if s.editor != nil {
		s.editor.SetTextContent(d.Text)
	}
	// Preview pane is a session UI toggle: loadActive does not force d.ShowPreview
	// so clicking another note never auto-opens markdown preview.
	s.dirty = d.Text != d.SavedText
	s.syntaxLangPick = "auto"
	d.ShowPreview = s.showPreview
}

func (sess *notepadSession) switchTo(s *notepadScene, id string) {
	if sess.find(id) == nil {
		return
	}
	keepPreview := s.showPreview
	sess.flushActive(s)
	sess.activeID = id
	sess.loadActive(s)
	s.showPreview = keepPreview
	if d := sess.active(); d != nil {
		d.ShowPreview = keepPreview
	}
	s.applyEditorSyntax()
	s.refreshPreviewForDocument()
	s.paneSettlePending = true
	s.refreshChrome()
	s.rebuildSessionList()
	s.persistOpenDocs()
	if s.uiDoc != nil {
		s.uiDoc.QueueMain(func() { s.warmInactiveMarkdownCaches() })
	}
}

// openNoteFromList switches the active note and, on mobile full-width Open Notes,
// dismisses the list so the editor is immediately visible (master-detail pattern).
func (s *notepadScene) openNoteFromList(docID string) {
	if s.sess == nil {
		return
	}
	s.sess.switchTo(s, docID)
	s.dismissOpenNotesAfterPick()
}

func (s *notepadScene) dismissOpenNotesAfterPick() {
	if !s.useMobilePanelSheets() || !s.showSession {
		return
	}
	s.showSession = false
	s.applySessionPanelLayout()
	s.applyPreviewLayout()
	if s.setSessionTog != nil {
		s.setSessionTog.Value.Set(false)
	}
	s.syncRibbonToggleState()
	if s.store != nil {
		p := s.store.LoadPrefs()
		p.ShowSessionPanel = false
		_ = s.store.SavePrefs(p)
	}
	s.paneSettlePending = true
	s.focusEditor()
}

func (sess *notepadSession) addUntitled(s *notepadScene) *notepadDoc {
	sess.flushActive(s)
	d := sess.newDoc()
	sess.docs = append(sess.docs, d)
	sess.activeID = d.ID
	sess.loadActive(s)
	s.applyEditorSyntax()
	s.refreshPreviewForDocument()
	s.refreshChrome()
	s.rebuildSessionList()
	s.ensureOpenNotesVisible()
	s.persistOpenDocs()
	return d
}

func (sess *notepadSession) removeActive(s *notepadScene) bool {
	if len(sess.docs) <= 1 {
		return false
	}
	idx := -1
	for i, d := range sess.docs {
		if d.ID == sess.activeID {
			idx = i
			break
		}
	}
	if idx < 0 {
		return false
	}
	sess.docs = append(sess.docs[:idx], sess.docs[idx+1:]...)
	sess.activeID = sess.docs[0].ID
	if idx > 0 && idx-1 < len(sess.docs) {
		sess.activeID = sess.docs[idx-1].ID
	}
	sess.loadActive(s)
	s.applyEditorSyntax()
	s.refreshPreviewForDocument()
	s.refreshChrome()
	s.rebuildSessionList()
	s.persistOpenDocs()
	return true
}

func (sess *notepadSession) dedupeByPath() {
	if sess == nil || len(sess.docs) < 2 {
		return
	}
	var out []*notepadDoc
	pathIndex := map[string]int{} // normalized path -> index in out
	for _, d := range sess.docs {
		if d.Path == "" {
			out = append(out, d)
			continue
		}
		key := filepath.Clean(d.Path)
		if runtime.GOOS == "windows" {
			key = strings.ToLower(key)
		}
		if idx, ok := pathIndex[key]; ok {
			if d.ID == sess.activeID {
				out[idx] = d
			}
			continue
		}
		pathIndex[key] = len(out)
		out = append(out, d)
	}
	if len(out) != len(sess.docs) {
		sess.docs = out
		if sess.find(sess.activeID) == nil && len(out) > 0 {
			sess.activeID = out[0].ID
		}
	}
}

// closeNoteByID removes an open note from the session (see notepad_session_close.go).
func (d *notepadDoc) displayTitle() string {
	if d.DisplayName != "" {
		return d.DisplayName
	}
	if d.Path != "" {
		return filepath.Base(d.Path)
	}
	return "Untitled"
}

func (d *notepadDoc) isDirty() bool {
	return d.Text != d.SavedText
}

func (d *notepadDoc) displaySubtitle() string {
	if d.isDirty() {
		return "Unsaved"
	}
	if d.Path != "" {
		return filepath.Dir(d.Path)
	}
	return ""
}

func (s *notepadScene) rebuildSessionList() {
	if s.openNotes == nil || s.openNotes.List == nil {
		return
	}
	if s.sess != nil {
		s.sess.dedupeByPath()
	}
	list := s.openNotes.List
	for _, ch := range append([]ui.Node(nil), list.Children()...) {
		list.RemoveChild(ch.ID())
	}
	if s.sess == nil {
		return
	}
	activeID := s.sess.activeID
	for _, d := range s.sess.docs {
		doc := d
		tile := ui.NewListTile("notepad-sess-"+doc.ID, doc.displayTitle(), doc.displaySubtitle(), 0, 0, 0, 0)
		tile.SetStyle("list-tile-pane")
		tile.Selected = doc.ID == activeID
		tile.OnClick = func() {
			s.openNoteFromList(doc.ID)
		}
		menuBtn := s.newOpenNoteMenuButton(doc.ID)
		tile.SetTrailing(menuBtn)
		list.AddChild(tile)
	}
	list.MarkDirty()
	if s.openNotes.Scroll != nil {
		s.openNotes.Scroll.MarkDirty()
	}
}

func (s *notepadScene) applySessionPanelLayout() {
	if s.mainSplit == nil || s.openNotes == nil || s.openNotes.Root == nil {
		return
	}
	if s.useMobilePanelSheets() {
		if s.showSession {
			s.openNotes.Root.Show()
			s.mainSplit.MinFirst = 160
			s.mainSplit.MinSecond = 0
			s.mainSplit.Ratio.Set(1)
		} else {
			s.openNotes.Root.Hide()
			s.mainSplit.MinFirst = 0
			s.mainSplit.MinSecond = 200
			s.mainSplit.Ratio.Set(0)
		}
		s.applyOpenNotesPaneWidth()
		s.mainSplit.MarkDirty()
		s.mainSplit.MarkDrawDirty()
		if s.mainSplit.Parent() != nil {
			s.mainSplit.Parent().MarkDirty()
			s.mainSplit.Parent().MarkDrawDirty()
		}
		s.deferOpenNotesChromeRelayout()
		return
	}
	if s.showSession {
		s.openNotes.Root.Show()
		s.mainSplit.Ratio.Set(0.24)
		s.mainSplit.MinFirst = 180
		s.mainSplit.MinSecond = 240
	} else {
		s.openNotes.Root.Hide()
		s.mainSplit.Ratio.Set(0)
		s.mainSplit.MinFirst = 0
	}
	s.mainSplit.MarkDirty()
	s.mainSplit.MarkDrawDirty()
	if s.mainSplit.Parent() != nil {
		s.mainSplit.Parent().MarkDirty()
		s.mainSplit.Parent().MarkDrawDirty()
	}
	s.settleNotesPreviewIdle()
}

// ensureOpenNotesVisible shows the sidebar when multiple notes are open.
func (s *notepadScene) ensureOpenNotesVisible() {
	if s.sess == nil || len(s.sess.docs) < 2 {
		return
	}
	if s.showSession {
		return
	}
	s.showSession = true
	s.applySessionPanelLayout()
	s.refreshOpenNotesChrome()
	if s.store != nil {
		p := s.store.LoadPrefs()
		p.ShowSessionPanel = true
		_ = s.store.SavePrefs(p)
	}
}

func (s *notepadScene) toggleSessionPanel() {
	s.showSession = !s.showSession
	s.applySessionPanelLayout()
	if s.useMobilePanelSheets() {
		s.applyPreviewLayout()
	}
	if s.setSessionTog != nil {
		s.setSessionTog.Value.Set(s.showSession)
	}
	s.syncRibbonToggleState()
	if s.store != nil {
		p := s.store.LoadPrefs()
		p.ShowSessionPanel = s.showSession
		_ = s.store.SavePrefs(p)
	}
}

func (s *notepadScene) doSaveAll() {
	if s.sess == nil {
		return
	}
	s.sess.flushActive(s)
	saved := 0
	for _, d := range s.sess.docs {
		if !d.isDirty() {
			continue
		}
		if d.Path == "" {
			continue
		}
		if err := os.WriteFile(d.Path, []byte(d.Text), 0o644); err != nil {
			ui.ShowToast("Save failed: "+filepath.Base(d.Path)+": "+err.Error(), ui.ToastError, 3*time.Second)
			continue
		}
		d.SavedText = d.Text
		saved++
	}
	if s.sess.active() != nil {
		s.savedText = s.sess.active().SavedText
		s.dirty = s.sess.active().isDirty()
	}
	s.refreshChrome()
	s.rebuildSessionList()
	if saved == 0 {
		ui.ShowToast("No saved paths to write (Save As untitled notes first)", ui.ToastInfo, 3*time.Second)
	} else {
		ui.ShowToast(fmt.Sprintf("Saved %d document(s)", saved), ui.ToastSuccess, 2*time.Second)
	}
}

func (s *notepadScene) refreshOpenNotesChrome() {
	if s.openNotes == nil || s.openNotes.Root == nil {
		return
	}
	s.openNotes.Root.MarkDirty()
	if s.openNotes.Hdr != nil {
		s.openNotes.Hdr.MarkDirty()
	}
	if s.openNotes.Title != nil {
		s.openNotes.Title.MarkDirty()
		s.openNotes.Title.Layout()
	}
	if s.uiDoc != nil && s.uiDoc.Root != nil {
		s.uiDoc.Root.Layout()
	}
}

func (s *notepadScene) applyOpenNotesPaneWidth() {
	if s.openNotes == nil || s.openNotes.Root == nil {
		return
	}
	if s.useMobilePanelSheets() && s.showSession {
		s.openNotes.Root.PreferredWidth = 0
		s.openNotes.Root.MinWidth = 0
	} else {
		s.openNotes.Root.PreferredWidth = s.openNotes.PreferredWidth
		s.openNotes.Root.MinWidth = s.openNotes.MinWidth
	}
	s.openNotes.Root.MarkDirty()
}

func (s *notepadScene) deferOpenNotesChromeRelayout() {
	if s.uiDoc == nil {
		s.refreshOpenNotesChrome()
		return
	}
	s.uiDoc.QueueMain(func() { s.refreshOpenNotesChrome() })
}
