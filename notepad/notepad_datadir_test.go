package notepad

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNotepadPortableDataDirDetectsShipLayout(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "dict", "en_US"), 0o755); err != nil {
		t.Fatal(err)
	}
	aff := filepath.Join(root, "dict", "en_US", "en_US.aff")
	if err := os.WriteFile(aff, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Simulate exe in root
	got := portableDataDirForRoot(root)
	want := filepath.Join(root, "data")
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestNotepadPortableDataDirMissingDict(t *testing.T) {
	root := t.TempDir()
	if portableDataDirForRoot(root) != "" {
		t.Fatal("expected empty without dict layout")
	}
}
