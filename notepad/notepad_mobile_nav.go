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
	if ui.AndroidGLES() {
		// Phone framebuffer (~1080) maps to BreakpointLG; always use mobile chrome.
		return true
	}
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
	// Draw path prefers Fill when PreferIconSharp1x is on — warm that weight so the
	// first visible frame is not a Regular→Fill swap (looks like a one-frame distort).
	ui.Phosphor.EnsureLoadedMany(icons, ui.PhosphorFill)
	ui.Phosphor.EnsureLoadedMany(icons, ui.PhosphorRegular)
	nav := ui.NewBottomNavigationBar("notepad-mobile-nav", []ui.BottomNavItem{
		{Phosphor: ui.PhosphorHome2, Label: "Home"},
		{Phosphor: ui.PhosphorEditBox, Label: "Edit"},
		{Phosphor: ui.PhosphorBookRead, Label: "View"},
		{Phosphor: ui.PhosphorSearch, Label: "Search"},
		{Phosphor: ui.PhosphorCodeBox, Label: "Syntax"},
	}, ui.NewSignal(notepadNavHome), 0, 0, 0, 0)
	// Compact icons-only is too small on phone framebuffer pixels; labeled bar
	// matches App Shell and keeps 56px Android icon slots.
	nav.Compact = false
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
		if ui.AndroidGLES() {
			// Density-independent tool strip (docs/ANDROID_DENSITY.md).
			want = ui.Dp(ui.PhoneToolStripHDp)
		}
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
	s.ribbon.Show()
	// Bottom nav owns Home/Edit/Search/View/Syntax — hide ribbon tab strip so it
	// cannot paint over Open Notes (the earlier "mixing views" bug).
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
	// Desktop waits a few frames so flex can settle before revealing the bar.
	// Android must show immediately — deferred Hide left the first taps unhandled.
	if !ui.AndroidGLES() && s.navMountFrames < 4 {
		s.navMountFrames++
		if !s.bottomNav.IsHidden() {
			s.bottomNav.Hide()
			s.syncMobileBottomChrome()
		}
		return
	}
	s.navMountFrames = 4
	if !s.useMobileBottomNav() || s.inSettings {
		if !s.bottomNav.IsHidden() {
			s.bottomNav.Hide()
			s.syncMobileBottomChrome()
		}
		return
	}
	wasHidden := s.bottomNav.IsHidden()
	examples.UpdateShellFooterAutoHide(s.bottomNav, doc)
	// Auto-hide Show/Hide runs after the frame's Layout. Without an immediate
	// reflow the bar draws one frame at stale/zero flex bounds (distorted strip).
	if wasHidden != s.bottomNav.IsHidden() {
		s.reflowAfterBottomNavToggle(doc)
	} else if !s.bottomNav.IsHidden() {
		s.pinMobileBottomNavBounds(doc)
	}
	s.syncMobileBottomChrome()
}

// reflowAfterBottomNavToggle re-runs root Layout so flex assigns real width/height
// the same frame the bar becomes visible (or frees space when it hides).
func (s *notepadScene) reflowAfterBottomNavToggle(doc *ui.Document) {
	if doc == nil || doc.Root == nil {
		return
	}
	doc.Root.MarkDirty()
	doc.Root.Layout()
	s.pinMobileBottomNavBounds(doc)
}

// pinMobileBottomNavBounds snaps the compact bar to the client bottom edge so a
// partial flex pass cannot leave it at 0×0 or a wrong Y for even one draw/hit.
func (s *notepadScene) pinMobileBottomNavBounds(doc *ui.Document) {
	if s.bottomNav == nil || doc == nil || s.bottomNav.IsHidden() {
		return
	}
	h := s.bottomNav.Bounds().Height
	if h < 1 {
		h = ui.BottomNavCompactHeight()
		if !s.bottomNav.Compact {
			if ui.AndroidGLES() {
				h = ui.Dp(ui.PhoneBottomNavHDp)
			} else {
				h = 64 * ui.PhoneChromeScale()
			}
		}
	}
	// Bounds are screen-absolute; content band starts at ChromeTop.
	y := doc.ChromeTop() + float32(doc.Height) - h
	want := rl.NewRectangle(0, y, float32(doc.Width), h)
	cur := s.bottomNav.Bounds()
	if abs32(cur.X-want.X) > 0.5 || abs32(cur.Y-want.Y) > 0.5 ||
		abs32(cur.Width-want.Width) > 0.5 || abs32(cur.Height-want.Height) > 0.5 {
		// PinScreenBounds: avoid MarkDirty→Layout immediately overwriting the pin.
		s.bottomNav.PinScreenBounds(want)
		s.bottomNav.Layout()
	}
}

// SyncAndroidChromeBand keeps the edge-to-edge shell and BottomNav on the
// ChromeTop content band. Call after Layout and before Update so hit-testing
// matches what will be drawn.
func (s *notepadScene) SyncAndroidChromeBand(doc *ui.Document) {
	if !ui.AndroidGLES() || doc == nil || doc.Root == nil {
		return
	}
	// Content band must sit at ChromeTop so flex BottomNav Y matches draw/hit
	// (docs/ANDROID_DENSITY.md). Prefer Document helper when sizes are known.
	if doc.Width > 0 && doc.Height > 0 {
		fullH := int32(doc.ChromeTop()) + doc.Height + int32(doc.ChromeBottom())
		if fullH < doc.Height {
			fullH = doc.Height
		}
		doc.EnsureContentBand(doc.Width, fullH)
	} else {
		band := rl.NewRectangle(0, doc.ChromeTop(), float32(doc.Width), float32(doc.Height))
		for _, ch := range doc.Root.Children() {
			b := ch.Bounds()
			if abs32(b.Y-band.Y) > 0.5 || abs32(b.Height-band.Height) > 0.5 ||
				abs32(b.Width-band.Width) > 0.5 {
				ch.SetBounds(band)
			}
		}
	}
	s.pinMobileBottomNavBounds(doc)
}

func abs32(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
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
