// Gru Notepad — A4 session restore from SQLite open_docs.
package notepad

import (
	"fmt"
	"github.com/ledocorp/gru/examples"
)

func (s *notepadScene) persistOpenDocs() {
	if s.store == nil || s.sess == nil {
		return
	}
	s.sess.flushActive(s)
	s.sess.dedupeByPath()
	_ = s.store.SaveOpenDocs(s.sess)
}

func (s *notepadScene) tryRestoreOpenDocs() bool {
	if s.store == nil || examples.StartupOpenFilePath() != "" {
		return false
	}
	rows, err := s.store.LoadOpenDocs()
	if err != nil || len(rows) == 0 {
		return false
	}
	sess := restoreNotepadSession(rows)
	if sess == nil || len(sess.docs) == 0 {
		return false
	}
	s.sess = sess
	active := sess.active()
	if active != nil {
		s.path = active.Path
		s.savedText = active.SavedText
		s.noteDisplayName = active.DisplayName
		s.showPreview = active.ShowPreview
	}
	return true
}

func restoreNotepadSession(rows []OpenDocRow) *notepadSession {
	sess := &notepadSession{nextNum: 1}
	activeID := ""
	for _, row := range rows {
		text, saved, path, name := resolveOpenDocBuffer(row)
		if path == "" && text == "" && name == "" {
			continue
		}
		d := &notepadDoc{
			ID:          row.DocID,
			Path:        path,
			DisplayName: name,
			Text:        text,
			SavedText:   saved,
			ShowPreview: row.ShowPreview,
		}
		sess.docs = append(sess.docs, d)
		if row.IsActive {
			activeID = d.ID
		}
		bumpDocNumFromID(sess, d.ID)
	}
	if len(sess.docs) == 0 {
		return nil
	}
	if activeID == "" {
		activeID = sess.docs[0].ID
	}
	sess.activeID = activeID
	sess.dedupeByPath()
	return sess
}

func bumpDocNumFromID(sess *notepadSession, id string) {
	var n int
	if _, err := fmt.Sscanf(id, "doc-%d", &n); err == nil && n >= sess.nextNum {
		sess.nextNum = n + 1
	}
}

// resolveOpenDocBuffer restores session buffers from SQLite (not disk reload).
func resolveOpenDocBuffer(row OpenDocRow) (text, savedText, path, displayName string) {
	path = row.Path
	text = row.BufferText
	savedText = row.SavedText
	displayName = row.DisplayName
	if savedText == "" && text != "" {
		savedText = text
	}
	return text, savedText, path, displayName
}
