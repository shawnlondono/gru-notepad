// Gru Notepad — save dialog filename hints from open-notes titles.
package notepad

import (
	"path/filepath"
	"strings"
)

// notepadSanitizeFilename strips characters illegal in file names on common desktops.
func notepadSanitizeFilename(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	replacer := strings.NewReplacer(
		`\`, "_",
		`/`, "_",
		`:`, "_",
		`*`, "_",
		`?`, "_",
		`"`, "_",
		`<`, "_",
		`>`, "_",
		`|`, "_",
	)
	name = strings.TrimSpace(replacer.Replace(name))
	return strings.TrimRight(name, "_")
}

// activeDocTitle is the human title for window/status (open-notes name before path exists).
func (s *notepadScene) activeDocTitle() string {
	if s.path != "" {
		return filepath.Base(s.path)
	}
	if name := strings.TrimSpace(s.noteDisplayName); name != "" {
		return name
	}
	if s.sess != nil {
		if d := s.sess.active(); d != nil {
			if name := strings.TrimSpace(d.DisplayName); name != "" {
				return name
			}
		}
	}
	return "Untitled"
}

// saveDefaultFilename is the native Save As dialog hint for the active note.
func (s *notepadScene) saveDefaultFilename(saveAs bool) string {
	if s.path != "" {
		return filepath.Base(s.path)
	}
	name := strings.TrimSpace(s.noteDisplayName)
	if name == "" && s.sess != nil {
		if d := s.sess.active(); d != nil {
			name = strings.TrimSpace(d.DisplayName)
		}
	}
	name = notepadSanitizeFilename(name)
	if name == "" {
		return defaultUntitledSaveName(s)
	}
	if filepath.Ext(name) == "" {
		name += defaultSaveExtension(s)
	}
	return name
}

func defaultUntitledSaveName(s *notepadScene) string {
	return "Untitled" + defaultSaveExtension(s)
}

func defaultSaveExtension(s *notepadScene) string {
	if s != nil && s.editorSyntaxLang() == "markdown" {
		return ".md"
	}
	return ".txt"
}

func (s *notepadScene) syncDisplayNameAfterSave(savedPath string) {
	if s.sess == nil {
		return
	}
	d := s.sess.active()
	if d == nil {
		return
	}
	d.Path = savedPath
	name := strings.TrimSpace(d.DisplayName)
	if name == "" {
		return
	}
	base := filepath.Base(savedPath)
	stem := strings.TrimSuffix(base, filepath.Ext(base))
	if strings.EqualFold(notepadSanitizeFilename(name), notepadSanitizeFilename(stem)) {
		return
	}
	d.DisplayName = ""
	s.noteDisplayName = ""
}
