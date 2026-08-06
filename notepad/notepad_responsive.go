// Gru Notepad — Phase 4 narrow-width panel layout (split pinned full, no drag).
package notepad

import (
	"github.com/ledocorp/gru/ui"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func (s *notepadScene) useMobilePanelSheets() bool {
	if !s.mobilePanelSheets {
		return false
	}
	return ui.CurrentBreakpoint(float32(rl.GetScreenWidth())) < ui.BreakpointLG
}

func (s *notepadScene) applyResponsivePanelLayout() {
	s.applySessionPanelLayout()
	s.applyPreviewLayout()
}

// focusEditor returns keyboard focus to the editor and restarts the caret blink.
func (s *notepadScene) focusEditor() {
	if s.uiDoc == nil || s.editor == nil || s.inSettings {
		return
	}
	s.uiDoc.SetFocus(s.editor)
	s.editor.RestartCaretCycle()
}

// closeMobileSheetsTransient hides full-width notes/preview without changing prefs.
func (s *notepadScene) closeMobileSheetsTransient() {
	if !s.useMobilePanelSheets() {
		return
	}
	changed := false
	if s.showSession {
		s.showSession = false
		changed = true
	}
	if s.showPreview {
		s.showPreview = false
		if s.sess != nil && s.sess.active() != nil {
			s.sess.active().ShowPreview = false
		}
		changed = true
	}
	if !changed {
		return
	}
	s.applySessionPanelLayout()
	s.applyPreviewLayout()
	if s.setSessionTog != nil {
		s.setSessionTog.Value.Set(s.showSession)
	}
	s.syncRibbonToggleState()
	s.focusEditor()
}
