//go:build android

// Gru Notepad — JSON prefs/session persistence on Android (no modernc/sqlite).
package notepad

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/ledocorp/gru/ui"
)

// NotepadStore is a JSON-backed store under notepadDataDir().
type NotepadStore struct {
	dir string
}

type jsonStoreSnapshot struct {
	SchemaVersion int               `json:"schema_version"`
	WelcomePack   string            `json:"welcome_pack"`
	Prefs         map[string]string `json:"prefs"`
	Recent        []jsonRecent      `json:"recent"`
	OpenDocs      []OpenDocRow      `json:"open_docs"`
}

type jsonRecent struct {
	Path       string `json:"path"`
	LastOpened int64  `json:"last_opened"`
}

// OpenNotepadStore opens (or creates) the JSON store for this install.
func OpenNotepadStore() (*NotepadStore, error) {
	appDir, err := notepadDataDir()
	if err != nil {
		return nil, err
	}
	s := &NotepadStore{dir: appDir}
	if err := s.ensure(); err != nil {
		return nil, err
	}
	if err := s.EnsureWelcomePack(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *NotepadStore) storePath() string {
	return filepath.Join(s.dir, "store.json")
}

func (s *NotepadStore) ensure() error {
	if s == nil {
		return nil
	}
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return err
	}
	if _, err := os.Stat(s.storePath()); err == nil {
		return nil
	}
	snap := jsonStoreSnapshot{
		SchemaVersion: notepadSchemaVersion,
		Prefs:         map[string]string{},
		Recent:        nil,
		OpenDocs:      nil,
	}
	return s.writeSnap(snap)
}

func (s *NotepadStore) readSnap() (jsonStoreSnapshot, error) {
	var snap jsonStoreSnapshot
	data, err := os.ReadFile(s.storePath())
	if err != nil {
		if os.IsNotExist(err) {
			return jsonStoreSnapshot{SchemaVersion: notepadSchemaVersion, Prefs: map[string]string{}}, nil
		}
		return snap, err
	}
	if err := json.Unmarshal(data, &snap); err != nil {
		return snap, err
	}
	if snap.Prefs == nil {
		snap.Prefs = map[string]string{}
	}
	return snap, nil
}

func (s *NotepadStore) writeSnap(snap jsonStoreSnapshot) error {
	if snap.Prefs == nil {
		snap.Prefs = map[string]string{}
	}
	snap.SchemaVersion = notepadSchemaVersion
	data, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.storePath() + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.storePath())
}

func (s *NotepadStore) Close() error { return nil }

func (s *NotepadStore) setPref(key, value string) error {
	snap, err := s.readSnap()
	if err != nil {
		return err
	}
	snap.Prefs[key] = value
	return s.writeSnap(snap)
}

func (s *NotepadStore) getPref(key, defaultVal string) string {
	snap, err := s.readSnap()
	if err != nil {
		return defaultVal
	}
	if v, ok := snap.Prefs[key]; ok {
		return v
	}
	return defaultVal
}

// LoadPrefs reads prefs from JSON.
func (s *NotepadStore) LoadPrefs() NotepadPrefs {
	if s == nil {
		return defaultNotepadPrefs()
	}
	p := defaultNotepadPrefs()
	p.WordWrap = s.getPref("word_wrap", "0") == "1"
	p.ShowStatus = s.getPref("show_status", "1") == "1"
	p.SyntaxHighlight = s.getPref("syntax_highlight", "1") == "1"
	p.SyntaxLangPick = s.getPref("syntax_lang_pick", "auto")
	if z, err := strconv.ParseFloat(s.getPref("default_zoom", "1"), 32); err == nil && z > 0 {
		p.DefaultZoom = float32(z)
	}
	p.ShowSessionPanel = s.getPref("show_session_panel", "1") == "1"
	p.MobilePanelSheets = s.getPref("mobile_panel_sheets", "1") == "1"
	p.MobileBottomNav = s.getPref("mobile_bottom_nav", "1") == "1"
	p.SpellCheck = s.getPref("spell_check", "0") == "1"
	p.ChromeMode = s.getPref("chrome_mode", notepadChromeMenubar)
	if p.ChromeMode != notepadChromeRibbon {
		p.ChromeMode = notepadChromeMenubar
	}
	p.AppearanceMode = ui.AppearanceMode(s.getPref("appearance_mode", ui.AppearanceSystem))
	p.ChromaStyle = s.getPref("chroma_style", "")
	if p.ChromaStyle == "" {
		p.ChromaStyle = notepadDefaultChromaStyle(p.AppearanceMode)
	}
	p.FindUseModal = s.getPref("find_use_modal", "0") == "1"
	p.EditorBackground = s.getPref("editor_background", "")
	return p
}

// SavePrefs writes prefs to JSON.
func (s *NotepadStore) SavePrefs(p NotepadPrefs) error {
	if s == nil {
		return nil
	}
	b := func(v bool) string {
		if v {
			return "1"
		}
		return "0"
	}
	snap, err := s.readSnap()
	if err != nil {
		return err
	}
	mode := p.ChromeMode
	if mode != notepadChromeRibbon {
		mode = notepadChromeMenubar
	}
	snap.Prefs["word_wrap"] = b(p.WordWrap)
	snap.Prefs["show_status"] = b(p.ShowStatus)
	snap.Prefs["syntax_highlight"] = b(p.SyntaxHighlight)
	snap.Prefs["syntax_lang_pick"] = p.SyntaxLangPick
	snap.Prefs["default_zoom"] = fmt.Sprintf("%g", p.DefaultZoom)
	snap.Prefs["show_session_panel"] = b(p.ShowSessionPanel)
	snap.Prefs["mobile_panel_sheets"] = b(p.MobilePanelSheets)
	snap.Prefs["mobile_bottom_nav"] = b(p.MobileBottomNav)
	snap.Prefs["spell_check"] = b(p.SpellCheck)
	snap.Prefs["chrome_mode"] = mode
	snap.Prefs["appearance_mode"] = ui.AppearanceMode(p.AppearanceMode)
	snap.Prefs["chroma_style"] = p.ChromaStyle
	snap.Prefs["find_use_modal"] = b(p.FindUseModal)
	snap.Prefs["editor_background"] = p.EditorBackground
	return s.writeSnap(snap)
}

// AddRecent records a file open (most recent first, capped).
func (s *NotepadStore) AddRecent(path string) error {
	if s == nil || path == "" {
		return nil
	}
	snap, err := s.readSnap()
	if err != nil {
		return err
	}
	now := time.Now().Unix()
	out := make([]jsonRecent, 0, len(snap.Recent)+1)
	out = append(out, jsonRecent{Path: path, LastOpened: now})
	for _, r := range snap.Recent {
		if r.Path == path {
			continue
		}
		out = append(out, r)
	}
	if len(out) > notepadRecentMax {
		out = out[:notepadRecentMax]
	}
	snap.Recent = out
	return s.writeSnap(snap)
}

// RecentFiles returns paths ordered by last opened (newest first).
func (s *NotepadStore) RecentFiles() ([]string, error) {
	if s == nil {
		return nil, nil
	}
	snap, err := s.readSnap()
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(snap.Recent))
	for _, r := range snap.Recent {
		out = append(out, r.Path)
	}
	return out, nil
}

// ClearRecent removes all recent file entries.
func (s *NotepadStore) ClearRecent() error {
	if s == nil {
		return nil
	}
	snap, err := s.readSnap()
	if err != nil {
		return err
	}
	snap.Recent = nil
	return s.writeSnap(snap)
}

// SaveOpenDocs replaces the open-doc snapshot.
func (s *NotepadStore) SaveOpenDocs(sess *notepadSession) error {
	if s == nil || sess == nil || len(sess.docs) == 0 {
		return nil
	}
	snap, err := s.readSnap()
	if err != nil {
		return err
	}
	rows := make([]OpenDocRow, 0, len(sess.docs))
	for i, d := range sess.docs {
		if i >= notepadOpenDocsMax {
			break
		}
		rows = append(rows, OpenDocRow{
			DocID:       d.ID,
			Path:        d.Path,
			DisplayName: d.DisplayName,
			SortOrder:   i,
			ShowPreview: d.ShowPreview,
			IsActive:    d.ID == sess.activeID,
			BufferText:  d.Text,
			SavedText:   d.SavedText,
		})
	}
	snap.OpenDocs = rows
	return s.writeSnap(snap)
}

// LoadOpenDocs returns saved open documents ordered by sort_order.
func (s *NotepadStore) LoadOpenDocs() ([]OpenDocRow, error) {
	if s == nil {
		return nil, nil
	}
	snap, err := s.readSnap()
	if err != nil {
		return nil, err
	}
	return snap.OpenDocs, nil
}

// ClearOpenDocs removes persisted session restore data.
func (s *NotepadStore) ClearOpenDocs() error {
	if s == nil {
		return nil
	}
	snap, err := s.readSnap()
	if err != nil {
		return err
	}
	snap.OpenDocs = nil
	return s.writeSnap(snap)
}

// EnsureWelcomePack seeds curated welcome notes once per pack version.
func (s *NotepadStore) EnsureWelcomePack() error {
	if s == nil {
		return nil
	}
	snap, err := s.readSnap()
	if err != nil {
		return err
	}
	if snap.WelcomePack == notepadWelcomePackVersion {
		return nil
	}
	notes := notepadWelcomeNotes()
	sess := &notepadSession{nextNum: 1}
	for i, n := range notes {
		d := &notepadDoc{
			ID:          n.id,
			DisplayName: n.displayName,
			Text:        n.text,
			SavedText:   n.text,
			ShowPreview: n.showPreview,
		}
		sess.docs = append(sess.docs, d)
		if i == 0 {
			sess.activeID = d.ID
		}
	}
	if err := s.SaveOpenDocs(sess); err != nil {
		return err
	}
	snap, err = s.readSnap()
	if err != nil {
		return err
	}
	snap.WelcomePack = notepadWelcomePackVersion
	snap.Recent = nil
	return s.writeSnap(snap)
}
