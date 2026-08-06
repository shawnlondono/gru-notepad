package notepad

import (
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func openTestNotepadStore(t *testing.T) *NotepadStore {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	db, err := sql.Open("sqlite", dbPath+"?_pragma=foreign_keys(1)")
	if err != nil {
		t.Fatal(err)
	}
	s := &NotepadStore{db: db}
	if err := s.migrate(); err != nil {
		t.Fatal(err)
	}
	return s
}

func TestNotepadStoreOpenDocsRoundTrip(t *testing.T) {
	store := openTestNotepadStore(t)
	defer store.Close()

	sess := &notepadSession{
		activeID: "doc-1",
		nextNum:  2,
		docs: []*notepadDoc{
			{
				ID: "doc-1", Path: "/tmp/a.md", DisplayName: "Draft A",
				Text: "alpha edit", SavedText: "alpha", ShowPreview: true,
			},
			{
				ID: "doc-2", DisplayName: "Scratch",
				Text: "untitled body", SavedText: "untitled body",
			},
		},
	}
	if err := store.SaveOpenDocs(sess); err != nil {
		t.Fatal(err)
	}
	rows, err := store.LoadOpenDocs()
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("rows: got %d want 2", len(rows))
	}
	if rows[0].DocID != "doc-1" || !rows[0].IsActive || !rows[0].ShowPreview {
		t.Fatalf("first row: %+v", rows[0])
	}
	if rows[0].BufferText != "alpha edit" || rows[0].SavedText != "alpha" || rows[0].DisplayName != "Draft A" {
		t.Fatalf("first row buffers: %+v", rows[0])
	}
	if rows[1].BufferText != "untitled body" || rows[1].DisplayName != "Scratch" {
		t.Fatalf("second row: %+v", rows[1])
	}
}

func TestRestoreNotepadSessionNamedUntitled(t *testing.T) {
	sess := restoreNotepadSession([]OpenDocRow{
		{
			DocID: "doc-1", DisplayName: "test", IsActive: true,
		},
	})
	if sess == nil || len(sess.docs) != 1 {
		t.Fatalf("sess: %+v", sess)
	}
	if sess.docs[0].DisplayName != "test" || sess.docs[0].displayTitle() != "test" {
		t.Fatalf("doc: %+v", sess.docs[0])
	}
}

func TestRestoreNotepadSessionDirtyUntitled(t *testing.T) {
	sess := restoreNotepadSession([]OpenDocRow{
		{
			DocID: "doc-2", BufferText: "hello world", SavedText: "hello",
			IsActive: true, DisplayName: "My Note",
		},
	})
	if sess == nil || len(sess.docs) != 1 {
		t.Fatalf("sess: %+v", sess)
	}
	d := sess.docs[0]
	if !d.isDirty() || d.DisplayName != "My Note" || d.Text != "hello world" {
		t.Fatalf("doc: %+v dirty=%v", d, d.isDirty())
	}
}

func TestNotepadWelcomePackSeedsFourNotes(t *testing.T) {
	store := openTestNotepadStore(t)
	defer store.Close()

	rows, err := store.LoadOpenDocs()
	if err != nil {
		t.Fatal(err)
	}
	want := notepadWelcomeNotes()
	if len(rows) != len(want) {
		t.Fatalf("open_docs=%d want %d welcome notes", len(rows), len(want))
	}
	for i, n := range want {
		if rows[i].DocID != n.id || rows[i].DisplayName != n.displayName {
			t.Fatalf("row %d: got %s/%s want %s/%s", i, rows[i].DocID, rows[i].DisplayName, n.id, n.displayName)
		}
		if rows[i].BufferText == "" || rows[i].BufferText != rows[i].SavedText {
			t.Fatalf("row %d: expected clean saved welcome body", i)
		}
	}
	// Second EnsureWelcomePack must not wipe user edits once pack version matches.
	rows[0].BufferText = "edited"
	sess := restoreNotepadSession(rows)
	if err := store.SaveOpenDocs(sess); err != nil {
		t.Fatal(err)
	}
	if err := store.EnsureWelcomePack(); err != nil {
		t.Fatal(err)
	}
	again, err := store.LoadOpenDocs()
	if err != nil {
		t.Fatal(err)
	}
	if again[0].BufferText != "edited" {
		t.Fatalf("welcome pack re-ran and wiped edits")
	}
}
