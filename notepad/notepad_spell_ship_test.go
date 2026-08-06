package notepad

import (
	"testing"

	"github.com/ledocorp/gru/ui"
)

func TestNotepadHunspellShipProtocol(t *testing.T) {
	c, err := ui.TryHunspellChecker()
	if err != nil {
		t.Skip("dict not present — run scripts/build/fetch-en-us-dict.ps1 or package with dict/")
	}
	if !c.Check("hello") {
		t.Fatal("hello should pass")
	}
	if c.Check("teh") {
		t.Fatal("teh should fail")
	}
	if c.Check("xyzzy") {
		t.Fatal("xyzzy should fail")
	}
}
