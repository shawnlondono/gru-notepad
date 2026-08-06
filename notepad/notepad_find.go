package notepad

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ledocorp/gru/ui"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func (s *notepadScene) showFindDialog() {
	if s.editor == nil || s.uiDoc == nil {
		return
	}
	ui.CloseContextMenu()
	query := notepadModalTextInput("notepad-find-query", "Find")
	if s.findQuery != "" {
		query.Text.Set(s.findQuery)
	}
	body := notepadModalFormBody("notepad-find-body",
		notepadModalCaption("notepad-find-label", "Find what:"),
		query,
	)
	runFind := func() {
		s.findQuery = strings.TrimSpace(query.Text.Get())
		if s.findQuery == "" {
			ui.ShowToast("Enter text to find", ui.ToastWarning, 2*time.Second)
			return
		}
		ui.CloseModal()
		if !s.editor.FindNext(s.findQuery, s.findMatchCase) {
			ui.ShowToast("Not found: "+s.findQuery, ui.ToastInfo, 2*time.Second)
		}
	}
	query.OnEnter = runFind
	ui.ShowModalSized("Find", body, []ui.ModalButton{
		{Label: "Find Next", Style: "primary", Action: runFind},
		{Label: "Cancel", Action: ui.CloseModal},
	}, 440, notepadSingleFieldModalH)
	s.uiDoc.SetFocus(query)
}

func (s *notepadScene) showReplaceDialog() {
	if s.editor == nil || s.uiDoc == nil {
		return
	}
	ui.CloseContextMenu()
	findIn := notepadModalTextInput("notepad-replace-find", "Find")
	replIn := notepadModalTextInput("notepad-replace-with", "Replace")
	if s.findQuery != "" {
		findIn.Text.Set(s.findQuery)
	}
	if s.replaceWith != "" {
		replIn.Text.Set(s.replaceWith)
	}
	body := notepadModalFormBody("notepad-replace-body",
		notepadModalCaption("notepad-replace-find-label", "Find what:"),
		findIn,
		notepadModalCaption("notepad-replace-with-label", "Replace with:"),
		replIn,
	)
	// 400px box → ~248px content band for two captions + two inputs + focus pad.
	apply := func(all bool) {
		s.findQuery = strings.TrimSpace(findIn.Text.Get())
		s.replaceWith = replIn.Text.Get()
		if s.findQuery == "" {
			ui.ShowToast("Enter text to find", ui.ToastWarning, 2*time.Second)
			return
		}
		ui.CloseModal()
		if all {
			n := s.editor.ReplaceAll(s.findQuery, s.replaceWith, s.findMatchCase)
			if n == 0 {
				ui.ShowToast("Not found: "+s.findQuery, ui.ToastInfo, 2*time.Second)
			} else {
				ui.ShowToast(fmt.Sprintf("Replaced %d occurrence(s)", n), ui.ToastSuccess, 2*time.Second)
			}
			return
		}
		if !s.editor.ReplaceOne(s.findQuery, s.replaceWith, s.findMatchCase) {
			ui.ShowToast("Not found: "+s.findQuery, ui.ToastInfo, 2*time.Second)
		}
	}
	findIn.OnEnter = func() { apply(false) }
	replIn.OnEnter = func() { apply(false) }
	ui.ShowModalSized("Replace", body, []ui.ModalButton{
		{Label: "Replace", Style: "primary", Action: func() { apply(false) }},
		{Label: "Replace All", Action: func() { apply(true) }},
		{Label: "Cancel", Action: ui.CloseModal},
	}, 440, notepadReplaceModalH)
	s.uiDoc.SetFocus(findIn)
}

func (s *notepadScene) showGoToDialog() {
	if s.editor == nil || s.uiDoc == nil {
		return
	}
	ui.CloseContextMenu()
	lineIn := notepadModalTextInput("notepad-goto-line", "Line number")
	if line, _ := s.editor.CursorLineCol(); line > 0 {
		lineIn.Text.Set(strconv.Itoa(line))
	}
	body := notepadModalFormBody("notepad-goto-body",
		notepadModalCaption("notepad-goto-label", "Line number:"),
		lineIn,
	)
	goLine := func() {
		n, err := strconv.Atoi(strings.TrimSpace(lineIn.Text.Get()))
		if err != nil || n < 1 {
			ui.ShowToast("Enter a valid line number", ui.ToastWarning, 2*time.Second)
			return
		}
		ui.CloseModal()
		if !s.editor.GoToLine(n) {
			ui.ShowToast(fmt.Sprintf("Line %d is out of range", n), ui.ToastWarning, 2*time.Second)
		}
	}
	lineIn.OnEnter = goLine
	ui.ShowModalSized("Go To Line", body, []ui.ModalButton{
		{Label: "Go", Style: "primary", Action: goLine},
		{Label: "Cancel", Action: ui.CloseModal},
	}, 400, notepadSingleFieldModalH)
	s.uiDoc.SetFocus(lineIn)
}

func (s *notepadScene) findNextOrPrompt() {
	if s.editor == nil {
		return
	}
	if s.findQuery == "" {
		if s.findUseModal {
			s.showFindDialog()
		} else {
			s.focusRibbonFindBar()
		}
		return
	}
	if !s.editor.FindNext(s.findQuery, s.findMatchCase) {
		ui.ShowToast("Not found: "+s.findQuery, ui.ToastInfo, 2*time.Second)
	}
}

func (s *notepadScene) findPreviousOrPrompt() {
	if s.editor == nil {
		return
	}
	if s.findQuery == "" {
		if s.findUseModal {
			s.showFindDialog()
		} else {
			s.focusRibbonFindBar()
		}
		return
	}
	if !s.editor.FindPrevious(s.findQuery, s.findMatchCase) {
		ui.ShowToast("Not found: "+s.findQuery, ui.ToastInfo, 2*time.Second)
	}
}

func (s *notepadScene) handleNotepadSearchKeys() {
	if s.editor == nil || ui.IsModalOpen() {
		return
	}
	if keyChordCtrl(rl.KeyF) {
		s.focusRibbonFindBar()
		if s.findUseModal {
			s.showFindDialog()
		}
	} else if keyChordCtrl(rl.KeyH) {
		if s.ribbon != nil && s.ribbonTabSearch >= 0 {
			s.ribbon.ActiveGroup.Set(s.ribbonTabSearch)
			s.ribbon.MarkDirty()
		}
		s.showReplaceDialog()
	} else if keyChordCtrl(rl.KeyG) {
		s.showGoToDialog()
	} else if rl.IsKeyPressed(rl.KeyF3) {
		shift := rl.IsKeyDown(rl.KeyLeftShift) || rl.IsKeyDown(rl.KeyRightShift)
		if shift {
			s.findPreviousOrPrompt()
		} else {
			s.findNextOrPrompt()
		}
	}
}
