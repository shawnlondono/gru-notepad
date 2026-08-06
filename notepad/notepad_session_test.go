package notepad

import (
	"path/filepath"
	"testing"
)

func TestFindByPathReturnsExistingDoc(t *testing.T) {
	sess := &notepadSession{
		docs: []*notepadDoc{
			{ID: "doc-1", Path: filepath.Join("C:", "notes", "readme.md")},
			{ID: "doc-2", Path: ""},
		},
	}
	want := filepath.Join("c:", "notes", "readme.md")
	found := sess.findByPath(want)
	if found == nil || found.ID != "doc-1" {
		t.Fatalf("findByPath(%q) = %v, want doc-1", want, found)
	}
	if sess.findByPath(filepath.Join("C:", "other.txt")) != nil {
		t.Fatal("expected nil for unopened path")
	}
}

func TestDedupeByPathKeepsActive(t *testing.T) {
	sess := &notepadSession{
		activeID: "doc-2",
		docs: []*notepadDoc{
			{ID: "doc-1", Path: filepath.Join("C:", "same.txt"), Text: "a", SavedText: "a"},
			{ID: "doc-2", Path: filepath.Join("c:", "same.txt"), Text: "b", SavedText: "b"},
		},
	}
	sess.dedupeByPath()
	if len(sess.docs) != 1 {
		t.Fatalf("docs: %d", len(sess.docs))
	}
	if sess.docs[0].ID != "doc-2" || sess.docs[0].Text != "b" {
		t.Fatalf("kept: %+v", sess.docs[0])
	}
}

func TestNotepadPathsEqual(t *testing.T) {
	a := filepath.Join("tmp", "a.txt")
	b := filepath.Join("tmp", "a.txt")
	if !notepadPathsEqual(a, b) {
		t.Fatalf("expected equal: %q %q", a, b)
	}
}

func TestNeedsDiscardConfirm(t *testing.T) {
	clean := &notepadDoc{Path: "/x.md", Text: "hi", SavedText: "hi"}
	if clean.needsDiscardConfirm() {
		t.Fatal("clean saved doc should not prompt")
	}
	dirty := &notepadDoc{Path: "/x.md", Text: "edit", SavedText: "hi"}
	if !dirty.needsDiscardConfirm() {
		t.Fatal("dirty doc should prompt")
	}
	untitled := &notepadDoc{Text: "draft"}
	if !untitled.needsDiscardConfirm() {
		t.Fatal("untitled with content should prompt")
	}
	empty := &notepadDoc{}
	if empty.needsDiscardConfirm() {
		t.Fatal("empty untitled should not prompt")
	}
}
