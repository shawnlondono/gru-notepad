package notepad

import (
	"testing"

	"github.com/ledocorp/gru/ui"
)

func TestNotepadSanitizeFilename(t *testing.T) {
	got := notepadSanitizeFilename(` draft:notes/test* `)
	if got != "draft_notes_test" {
		t.Fatalf("sanitize = %q", got)
	}
	if notepadSanitizeFilename("") != "" {
		t.Fatal("empty should stay empty")
	}
}

func TestSaveDefaultFilenameUsesDisplayName(t *testing.T) {
	s := &notepadScene{
		noteDisplayName: "My Snippet",
		sess:            newNotepadSession(),
	}
	if got := s.saveDefaultFilename(false); got != "My Snippet.txt" {
		t.Fatalf("got %q want My Snippet.txt", got)
	}
}

func TestSaveDefaultFilenameMarkdownExt(t *testing.T) {
	s := &notepadScene{
		noteDisplayName: "README",
		syntaxHighlight: true,
		sess:            newNotepadSession(),
	}
	ed := ui.NewTextEditor("test-ed", "", 0, 0, 0, 0)
	ed.SetSyntaxLanguage("markdown")
	s.editor = ed
	if got := s.saveDefaultFilename(false); got != "README.md" {
		t.Fatalf("got %q want README.md", got)
	}
}

func TestSaveDefaultFilenameExistingPath(t *testing.T) {
	s := &notepadScene{
		path:            `C:\notes\alpha.md`,
		noteDisplayName: "Ignored",
	}
	if got := s.saveDefaultFilename(true); got != "alpha.md" {
		t.Fatalf("got %q want alpha.md", got)
	}
}

func TestSaveDefaultFilenameFromSessionDoc(t *testing.T) {
	s := &notepadScene{sess: newNotepadSession()}
	d := s.sess.active()
	d.DisplayName = "Meeting Notes"
	if got := s.saveDefaultFilename(false); got != "Meeting Notes.txt" {
		t.Fatalf("got %q want Meeting Notes.txt", got)
	}
}

func TestActiveDocTitlePrefersDisplayName(t *testing.T) {
	s := &notepadScene{noteDisplayName: "Draft A"}
	if got := s.activeDocTitle(); got != "Draft A" {
		t.Fatalf("got %q", got)
	}
}
