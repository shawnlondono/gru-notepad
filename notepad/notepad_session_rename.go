// Gru Notepad — rename open notes (display title before/after save).
package notepad

import (
	"strings"
	"time"

	"github.com/ledocorp/gru/ui"
)

func (s *notepadScene) showRenameNoteDialog(docID string) {
	if s.sess == nil || s.uiDoc == nil {
		return
	}
	d := s.sess.find(docID)
	if d == nil {
		return
	}
	ui.CloseContextMenu()
	name := notepadModalTextInput("notepad-rename-note", "Note name")
	name.Text.Set(d.displayTitle())
	body := notepadModalFormBody("notepad-rename-body",
		notepadModalCaption("notepad-rename-label", "Name in Open Notes:"),
		name,
	)
	apply := func() {
		title := strings.TrimSpace(name.Text.Get())
		if title == "" {
			ui.ShowToast("Enter a name", ui.ToastWarning, 2*time.Second)
			return
		}
		d.DisplayName = title
		if s.sess.activeID == docID {
			s.noteDisplayName = title
		}
		s.rebuildSessionList()
		s.refreshChrome()
		s.persistOpenDocs()
		ui.CloseModal()
	}
	name.OnEnter = apply
	ui.ShowModalSized("Rename note", body, []ui.ModalButton{
		{Label: "OK", Style: "primary", Action: apply},
		{Label: "Cancel", Action: ui.CloseModal},
	}, 440, notepadSingleFieldModalH)
	s.uiDoc.SetFocus(name)
}
