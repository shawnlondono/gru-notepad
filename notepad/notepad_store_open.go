//go:build !android

package notepad

import (
	"database/sql"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// OpenNotepadStore opens (or creates) the app database for this install.
func OpenNotepadStore() (*NotepadStore, error) {
	appDir, err := notepadDataDir()
	if err != nil {
		return nil, err
	}
	dbPath := filepath.Join(appDir, "notepad.db")
	db, err := sql.Open("sqlite", dbPath+"?_pragma=foreign_keys(1)")
	if err != nil {
		return nil, err
	}
	s := &NotepadStore{db: db}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, err
	}
	// Re-seed when welcome pack version bumps (even if schema is already current).
	if err := s.EnsureWelcomePack(); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

