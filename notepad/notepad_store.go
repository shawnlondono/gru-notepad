//go:build !android

// Gru Notepad — SQLite prefs and recent files (UI persistence baselines).
// See docs/SQLITE_UI_PERSISTENCE.md.
package notepad

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/ledocorp/gru/ui"
)

// NotepadStore provides prefs and recent-file persistence for Gru Notepad.
type NotepadStore struct {
	db *sql.DB
}

func (s *NotepadStore) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *NotepadStore) migrate() error {
	_, err := s.db.Exec(`
CREATE TABLE IF NOT EXISTS meta (
  key TEXT PRIMARY KEY,
  value TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS prefs (
  key TEXT PRIMARY KEY,
  value TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS recent_files (
  path TEXT PRIMARY KEY,
  last_opened INTEGER NOT NULL
);
CREATE TABLE IF NOT EXISTS open_docs (
  doc_id TEXT PRIMARY KEY,
  path TEXT NOT NULL DEFAULT '',
  sort_order INTEGER NOT NULL,
  show_preview INTEGER NOT NULL DEFAULT 0,
  is_active INTEGER NOT NULL DEFAULT 0,
  buffer_text TEXT NOT NULL DEFAULT '',
  saved_text TEXT NOT NULL DEFAULT '',
  display_name TEXT NOT NULL DEFAULT ''
);
`)
	if err != nil {
		return err
	}
	ver, err := s.schemaVersion()
	if err != nil {
		return err
	}
	if ver == 0 {
		_, err = s.db.Exec(`INSERT INTO meta (key, value) VALUES ('schema_version', ?)`, strconv.Itoa(notepadSchemaVersion))
		if err != nil {
			return err
		}
		return s.EnsureWelcomePack()
	}
	if ver < 2 {
		if _, err := s.db.Exec(`
CREATE TABLE IF NOT EXISTS open_docs (
  doc_id TEXT PRIMARY KEY,
  path TEXT NOT NULL DEFAULT '',
  sort_order INTEGER NOT NULL DEFAULT 0,
  show_preview INTEGER NOT NULL DEFAULT 0,
  is_active INTEGER NOT NULL DEFAULT 0,
  buffer_text TEXT NOT NULL DEFAULT ''
);`); err != nil {
			return err
		}
		ver = 2
	}
	if ver < 3 {
		_, _ = s.db.Exec(`ALTER TABLE open_docs ADD COLUMN saved_text TEXT NOT NULL DEFAULT ''`)
		_, _ = s.db.Exec(`ALTER TABLE open_docs ADD COLUMN display_name TEXT NOT NULL DEFAULT ''`)
		if _, err := s.db.Exec(`UPDATE meta SET value = ? WHERE key = 'schema_version'`, "3"); err != nil {
			return err
		}
		ver = 3
	}
	if ver < 4 {
		if err := s.EnsureWelcomePack(); err != nil {
			return err
		}
		if _, err := s.db.Exec(`UPDATE meta SET value = ? WHERE key = 'schema_version'`, "4"); err != nil {
			return err
		}
	}
	return nil
}

func (s *NotepadStore) schemaVersion() (int, error) {
	if s == nil || s.db == nil {
		return 0, nil
	}
	var ver string
	err := s.db.QueryRow(`SELECT value FROM meta WHERE key = 'schema_version'`).Scan(&ver)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	n, err := strconv.Atoi(ver)
	if err != nil {
		return 0, err
	}
	return n, nil
}

func (s *NotepadStore) setPref(key, value string) error {
	_, err := s.db.Exec(`
INSERT INTO prefs (key, value) VALUES (?, ?)
ON CONFLICT(key) DO UPDATE SET value = excluded.value`, key, value)
	return err
}

func (s *NotepadStore) getPref(key, defaultVal string) string {
	var v string
	err := s.db.QueryRow(`SELECT value FROM prefs WHERE key = ?`, key).Scan(&v)
	if err != nil {
		return defaultVal
	}
	return v
}

// LoadPrefs reads prefs from the database.
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

// SavePrefs writes prefs to the database.
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
	if err := s.setPref("word_wrap", b(p.WordWrap)); err != nil {
		return err
	}
	if err := s.setPref("show_status", b(p.ShowStatus)); err != nil {
		return err
	}
	if err := s.setPref("syntax_highlight", b(p.SyntaxHighlight)); err != nil {
		return err
	}
	if err := s.setPref("syntax_lang_pick", p.SyntaxLangPick); err != nil {
		return err
	}
	if err := s.setPref("default_zoom", fmt.Sprintf("%g", p.DefaultZoom)); err != nil {
		return err
	}
	if err := s.setPref("show_session_panel", b(p.ShowSessionPanel)); err != nil {
		return err
	}
	if err := s.setPref("mobile_panel_sheets", b(p.MobilePanelSheets)); err != nil {
		return err
	}
	if err := s.setPref("mobile_bottom_nav", b(p.MobileBottomNav)); err != nil {
		return err
	}
	if err := s.setPref("spell_check", b(p.SpellCheck)); err != nil {
		return err
	}
	mode := p.ChromeMode
	if mode != notepadChromeRibbon {
		mode = notepadChromeMenubar
	}
	if err := s.setPref("chrome_mode", mode); err != nil {
		return err
	}
	if err := s.setPref("appearance_mode", ui.AppearanceMode(p.AppearanceMode)); err != nil {
		return err
	}
	if err := s.setPref("chroma_style", p.ChromaStyle); err != nil {
		return err
	}
	if err := s.setPref("find_use_modal", b(p.FindUseModal)); err != nil {
		return err
	}
	return s.setPref("editor_background", p.EditorBackground)
}


// AddRecent records a file open (most recent first, capped).
func (s *NotepadStore) AddRecent(path string) error {
	if s == nil || path == "" {
		return nil
	}
	now := time.Now().Unix()
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`
INSERT INTO recent_files (path, last_opened) VALUES (?, ?)
ON CONFLICT(path) DO UPDATE SET last_opened = excluded.last_opened`, path, now); err != nil {
		return err
	}
	_, err = tx.Exec(`
DELETE FROM recent_files WHERE path NOT IN (
  SELECT path FROM recent_files ORDER BY last_opened DESC LIMIT ?
)`, notepadRecentMax)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// RecentFiles returns paths ordered by last opened (newest first).
func (s *NotepadStore) RecentFiles() ([]string, error) {
	if s == nil {
		return nil, nil
	}
	rows, err := s.db.Query(`SELECT path FROM recent_files ORDER BY last_opened DESC LIMIT ?`, notepadRecentMax)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// ClearRecent removes all recent file entries.
func (s *NotepadStore) ClearRecent() error {
	if s == nil {
		return nil
	}
	_, err := s.db.Exec(`DELETE FROM recent_files`)
	return err
}

// SaveOpenDocs replaces the open-doc snapshot (call after flushActive).
func (s *NotepadStore) SaveOpenDocs(sess *notepadSession) error {
	if s == nil || sess == nil || len(sess.docs) == 0 {
		return nil
	}
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`DELETE FROM open_docs`); err != nil {
		return err
	}
	stmt, err := tx.Prepare(`
INSERT INTO open_docs (doc_id, path, sort_order, show_preview, is_active, buffer_text, saved_text, display_name)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for i, d := range sess.docs {
		if i >= notepadOpenDocsMax {
			break
		}
		active := 0
		if d.ID == sess.activeID {
			active = 1
		}
		preview := 0
		if d.ShowPreview {
			preview = 1
		}
		if _, err := stmt.Exec(d.ID, d.Path, i, preview, active, d.Text, d.SavedText, d.DisplayName); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// LoadOpenDocs returns saved open documents ordered by sort_order.
func (s *NotepadStore) LoadOpenDocs() ([]OpenDocRow, error) {
	if s == nil {
		return nil, nil
	}
	rows, err := s.db.Query(`
SELECT doc_id, path, sort_order, show_preview, is_active, buffer_text, saved_text, display_name
FROM open_docs ORDER BY sort_order ASC LIMIT ?`, notepadOpenDocsMax)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []OpenDocRow
	for rows.Next() {
		var r OpenDocRow
		var preview, active int
		if err := rows.Scan(&r.DocID, &r.Path, &r.SortOrder, &preview, &active, &r.BufferText, &r.SavedText, &r.DisplayName); err != nil {
			return nil, err
		}
		r.ShowPreview = preview == 1
		r.IsActive = active == 1
		out = append(out, r)
	}
	return out, rows.Err()
}

// ClearOpenDocs removes persisted session restore data.
func (s *NotepadStore) ClearOpenDocs() error {
	if s == nil {
		return nil
	}
	_, err := s.db.Exec(`DELETE FROM open_docs`)
	return err
}

// EnsureWelcomePack wipes open_docs + recent_files and seeds the curated welcome notes.
// Idempotent via meta.welcome_pack — bump notepadWelcomePackVersion to re-seed.
func (s *NotepadStore) EnsureWelcomePack() error {
	if s == nil || s.db == nil {
		return nil
	}
	var cur string
	err := s.db.QueryRow(`SELECT value FROM meta WHERE key = 'welcome_pack'`).Scan(&cur)
	if err == nil && cur == notepadWelcomePackVersion {
		return nil
	}
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	if _, err := s.db.Exec(`DELETE FROM open_docs`); err != nil {
		return err
	}
	if _, err := s.db.Exec(`DELETE FROM recent_files`); err != nil {
		return err
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
	_, err = s.db.Exec(`
INSERT INTO meta (key, value) VALUES ('welcome_pack', ?)
ON CONFLICT(key) DO UPDATE SET value = excluded.value`, notepadWelcomePackVersion)
	return err
}
