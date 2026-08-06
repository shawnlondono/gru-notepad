// Gru Notepad — ribbon inline find bar vs modal find.
package notepad

import (
	"strings"

	"github.com/ledocorp/gru/ui"
)

const (
	ribbonFindBarW    = float32(136) // ~2 icon cells; flex-shrinks before scroll
	ribbonFindBarMinW = float32(88)
	ribbonFindBarMaxW = float32(200)
)

func (s *notepadScene) wireRibbonFindBar(ti *ui.TextInput) {
	if ti == nil {
		return
	}
	if s.findQuery != "" {
		ti.Text.Set(s.findQuery)
	}
	ti.Text.Subscribe(func() {
		q := strings.TrimSpace(ti.Text.Get())
		s.findQuery = q
		if q == "" || s.editor == nil {
			return
		}
		s.editor.FindNext(q, s.findMatchCase)
	})
}

func (s *notepadScene) syncRibbonFindMode() {
	if s.ribbon == nil {
		return
	}
	s.ribbon.SetItemSuppressed("find", !s.findUseModal)
	s.ribbon.SetItemSuppressed("find-bar", s.findUseModal)
	if s.ribbonFindBar != nil {
		s.ribbonFindBar.PreferredWidth = ribbonFindBarW
		s.ribbonFindBar.MinWidth = ribbonFindBarMinW
		s.ribbonFindBar.MaxWidth = ribbonFindBarMaxW
	}
	s.ribbon.MarkDirty()
}

func (s *notepadScene) focusRibbonFindBar() {
	if s.ribbon != nil && s.ribbonTabSearch >= 0 {
		s.ribbon.ActiveGroup.Set(s.ribbonTabSearch)
		s.ribbon.MarkDirty()
	}
	if s.ribbonFindBar == nil || s.uiDoc == nil {
		return
	}
	if s.findQuery != "" {
		s.ribbonFindBar.Text.Set(s.findQuery)
	}
	s.uiDoc.SetFocus(s.ribbonFindBar)
}
