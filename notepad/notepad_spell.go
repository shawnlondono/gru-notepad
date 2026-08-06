// Gru Notepad — spell-check wiring (Phase 3b + Hunspell).
package notepad

import (
	"fmt"
	"os"
	"sync"

	"github.com/ledocorp/gru/ui"
)

var (
	notepadSpellOnce    sync.Once
	notepadSpellChecker ui.SpellChecker
	notepadSpellBackend string
)

// NotepadSpellChecker returns the shared spell checker, resolving Hunspell on first use
// so dictionary paths beside the running executable are found reliably.
func NotepadSpellChecker() ui.SpellChecker {
	notepadSpellOnce.Do(initNotepadSpell)
	return notepadSpellChecker
}

// NotepadSpellBackend reports "hunspell" or "stub" after the first init.
func NotepadSpellBackend() string {
	notepadSpellOnce.Do(initNotepadSpell)
	return notepadSpellBackend
}

func initNotepadSpell() {
	if c, err := ui.TryHunspellChecker(notepadSpellExtraWords...); err == nil {
		notepadSpellChecker = c
		notepadSpellBackend = "hunspell"
	} else {
		notepadSpellChecker = ui.NewSimpleSpellChecker(notepadSpellExtraWords...)
		notepadSpellBackend = "stub"
		fmt.Fprintf(os.Stderr, "Gru Notepad spell: stub (%v)\n", err)
		return
	}
	fmt.Fprintf(os.Stderr, "Gru Notepad spell: hunspell\n")
}
