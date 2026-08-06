// Gru Notepad — remove notes from Open Notes (session only; file on disk unchanged).
package notepad

import (
	"strings"
	"time"

	"github.com/ledocorp/gru/ui"
)

func (d *notepadDoc) needsDiscardConfirm() bool {
	if d.isDirty() {
		return true
	}
	if d.Path == "" && strings.TrimSpace(d.Text) != "" {
		return true
	}
	return false
}

// closeNoteByID removes a note from the open session. Saved files stay on disk.
func (s *notepadScene) closeNoteByID(id string) {
	if s.sess == nil {
		return
	}
	if len(s.sess.docs) <= 1 {
		ui.ShowToast("Cannot close the only open note", ui.ToastInfo, 2*time.Second)
		return
	}
	d := s.sess.find(id)
	if d == nil {
		return
	}
	if d.needsDiscardConfirm() {
		s.confirmDiscardNote(id)
		return
	}
	s.removeNoteFromSession(id)
}

func (s *notepadScene) confirmDiscardNote(id string) {
	if s.sess == nil {
		return
	}
	d := s.sess.find(id)
	if d == nil {
		return
	}
	msg := "Discard unsaved changes and remove this note from Open Notes?"
	if d.Path == "" {
		msg = "Discard this untitled note?"
	}
	ui.ShowModalSized("Discard note?", notepadModalConfirmBody("notepad-discard", msg), []ui.ModalButton{
		{Label: "Discard", Style: "danger", Action: func() {
			ui.CloseModal()
			s.removeNoteFromSession(id)
		}},
		{Label: "Cancel", Action: ui.CloseModal},
	}, 420, 240)
}

func (s *notepadScene) removeNoteFromSession(id string) {
	if s.sess == nil || len(s.sess.docs) <= 1 {
		return
	}
	idx := -1
	for i, doc := range s.sess.docs {
		if doc.ID == id {
			idx = i
			break
		}
	}
	if idx < 0 {
		return
	}
	wasActive := s.sess.activeID == id
	s.sess.docs = append(s.sess.docs[:idx], s.sess.docs[idx+1:]...)
	if wasActive {
		if idx > 0 {
			s.sess.activeID = s.sess.docs[idx-1].ID
		} else {
			s.sess.activeID = s.sess.docs[0].ID
		}
		s.sess.loadActive(s)
		s.applyEditorSyntax()
		s.refreshPreviewForDocument()
	}
	s.rebuildSessionList()
	s.refreshChrome()
	s.persistOpenDocs()
}
