// Gru Notepad — Open Notes row overflow menu (rename / remove from session).
package notepad

import (
	rl "github.com/gen2brain/raylib-go/raylib"

	"github.com/ledocorp/gru/ui"
)

func notepadOpenNoteMenuButton(id string) *ui.IconButton {
	btn := ui.NewIconButton(id, "⋮", "", 0, 0, 36, 36)
	btn.SetStyle("appbar-icon")
	btn.SetPhosphorIcon(ui.PhosphorDotsThreeVertical, ui.PhosphorRegular)
	ui.Phosphor.EnsureLoaded(ui.PhosphorDotsThreeVertical, ui.PhosphorRegular)
	btn.PreferredWidth = 36
	return btn
}

func (s *notepadScene) newOpenNoteMenuButton(docID string) *ui.IconButton {
	btn := notepadOpenNoteMenuButton("notepad-sess-menu-" + docID)
	btn.OnClick = func() {
		s.showOpenNoteMenu(docID, btn.Bounds())
	}
	return btn
}

func (s *notepadScene) showOpenNoteMenu(docID string, anchor rl.Rectangle) {
	ui.CloseContextMenu()
	ui.ShowContextMenu([]ui.ContextMenuItem{
		{
			Label: "Rename",
			Action: func() {
				ui.CloseContextMenu()
				s.showRenameNoteDialog(docID)
			},
		},
		{
			Label: "Remove from list",
			Action: func() {
				ui.CloseContextMenu()
				s.closeNoteByID(docID)
			},
		},
	}, anchor.X, anchor.Y+anchor.Height)
}
