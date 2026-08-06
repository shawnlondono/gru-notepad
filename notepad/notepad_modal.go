// Gru Notepad — modal helpers (widgets_demo confirm contract).
package notepad

import "github.com/ledocorp/gru/ui"

// notepadModalConfirmBody is a single wrapped message (widgets_demo Info modal).
func notepadModalConfirmBody(id, message string) ui.Node {
	body := ui.NewContainer(id, 0, 0, 0, 0)
	body.LayoutType = ui.LayoutFlex
	body.FlexDirection = ui.FlexColumn
	body.Gap = 8
	body.AutoHeight = true
	body.SetStyle("transparent")
	lbl := ui.NewLabel(id+"-text", message, 0, 0, 0, 0)
	lbl.SetStyle("form-value")
	lbl.Align = ui.LabelAlignLeft
	lbl.Wrap = true
	body.AddChild(lbl)
	return body
}

// notepadModalCaption is a muted field caption inside a modal form.
func notepadModalCaption(id, text string) *ui.Label {
	lbl := ui.NewLabel(id, text, 0, 0, 0, 0)
	lbl.SetStyle("modal-body")
	lbl.Align = ui.LabelAlignLeft
	return lbl
}

// notepadModalTextInput is a modal form field with grey placeholder hints.
func notepadModalTextInput(id, placeholder string) *ui.TextInput {
	in := ui.NewTextInput(id, "", 0, 0, 0, 44)
	in.SetStyle("modal-input")
	in.Placeholder = placeholder
	return in
}

// notepadModalFormBody stacks captions and inputs inside the modal content band.
func notepadModalFormBody(id string, children ...ui.Node) *ui.Container {
	body := ui.NewContainer(id, 0, 0, 0, 0)
	body.LayoutType = ui.LayoutFlex
	body.FlexDirection = ui.FlexColumn
	body.Gap = 8
	body.AutoHeight = true
	body.SetStyle("transparent")
	for _, c := range children {
		body.AddChild(c)
	}
	// TextInput focus accent draws ~7px below layout bounds; keep it inside the clip band.
	pad := ui.NewContainer(id+"-focus-pad", 0, 0, 0, 8)
	pad.SetStyle("transparent")
	body.AddChild(pad)
	return body
}

// notepadSingleFieldModalH fits one caption + one 44px input + focus-accent pad (Find / Rename / Go To).
const notepadSingleFieldModalH = float32(280)

// notepadReplaceModalH fits two captions + two 44px inputs in the modal content band.
const notepadReplaceModalH = float32(400)
