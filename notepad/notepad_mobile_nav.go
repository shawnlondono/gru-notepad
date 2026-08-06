// Gru Notepad — compact mobile/tablet bottom navigation (Phase 4).
package notepad

import (
	"github.com/ledocorp/gru/examples"
	"github.com/ledocorp/gru/ui"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	notepadNavHome   = 0
	notepadNavEdit   = 1
	notepadNavView   = 2
	notepadNavSearch = 3
	notepadNavSyntax = 4
)

func notepadNavToRibbonTab(navIdx int) int {
	switch navIdx {
	case notepadNavEdit:
		return 1
	case notepadNavView:
		return 3
	case notepadNavSearch:
		return 2
	case notepadNavSyntax:
		return 4
	default:
		return 0
	}
}

func (s *notepadScene) useMobileBottomNav() bool {
	if !s.mobileBottomNav {
		return false
	}
	return ui.CurrentBreakpoint(float32(rl.GetScreenWidth())) < ui.BreakpointLG
}

func (s *notepadScene) buildMobileBottomNav() *ui.BottomNavigationBar {
	icons := []string{
		ui.PhosphorHome2,
		ui.PhosphorEditBox,
		ui.PhosphorBookRead,
		ui.PhosphorSearch,
		ui.PhosphorCodeBox,
	}
	for _, name := range icons {
		ui.Phosphor.EnsureLoaded(name, ui.PhosphorRegular)
	}
	nav := ui.NewBottomNavigationBar("notepad-mobile-nav", []ui.BottomNavItem{
		{Phosphor: ui.PhosphorHome2},
		{Phosphor: ui.PhosphorEditBox},
		{Phosphor: ui.PhosphorBookRead},
		{Phosphor: ui.PhosphorSearch},
		{Phosphor: ui.PhosphorCodeBox},
	}, ui.NewSignal(notepadNavHome), 0, 0, 0, 0)
	nav.Compact = true
	nav.SetZIndex(120)
	nav.Hide()
	nav.OnChange = func(i int) {
		if i == notepadNavHome {
			s.closeMobileSheetsTransient()
		}
		s.syncMobileRibbonTab(i)
	}
	return nav
}

func (s *notepadScene) syncMobileRibbonTab(navIdx int) {
	if s.ribbon == nil {
		return
	}
	tab := notepadNavToRibbonTab(navIdx)
	if s.ribbon.ActiveGroup.Get() != tab {
		s.ribbon.ActiveGroup.Set(tab)
	}
	s.ribbon.MarkDirty()
}

func (s *notepadScene) syncMobileRibbonHeight() {
	if s.ribbon == nil {
		return
	}
	want := notepadRibbonH
	if s.useMobileBottomNav() {
		want = notepadMobileRibbonH
	}
	b := s.ribbon.Bounds()
	if b.Height < want-0.5 || b.Height > want+0.5 {
		b.Height = want
		s.ribbon.SetBounds(b)
		s.ribbon.MarkDirty()
	}
}

func (s *notepadScene) applyMobileRibbonChrome() {
	if s.ribbon == nil {
		return
	}
	mobile := s.useMobileBottomNav()
	s.ribbon.SetHideRibbonTabs(mobile)
	s.ribbon.SetHideRibbonSectionLabels(mobile)
	s.syncMobileRibbonHeight()
	if mobile && s.bottomNav != nil {
		s.syncMobileRibbonTab(s.bottomNav.Selected.Get())
	}
	s.ribbon.Layout()
	s.ribbon.MarkDrawDirty()
}

func (s *notepadScene) updateMobileBottomNav(doc *ui.Document) {
	if s.bottomNav == nil || doc == nil {
		return
	}
	if !s.useMobileBottomNav() || s.inSettings {
		if !s.bottomNav.IsHidden() {
			s.bottomNav.Hide()
			s.syncMobileBottomChrome()
		}
		return
	}
	examples.UpdateShellFooterAutoHide(s.bottomNav, doc)
	s.syncMobileBottomChrome()
}

func (s *notepadScene) syncMobileBottomChrome() {
	if !s.useMobileBottomNav() || s.inSettings {
		s.toggleStatusBarApply(s.showStatus)
		return
	}
	navOpen := s.bottomNav != nil && !s.bottomNav.IsHidden()
	if navOpen {
		if s.statusBar != nil {
			s.statusBar.Hide()
		}
		return
	}
	s.toggleStatusBarApply(s.showStatus)
}
