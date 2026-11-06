package main

import (
	"image/color"
	"reflect"
	"testing"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/exp/charmtone"
	"github.com/resendlabs/resend-go"
)

func TestComposerBackground(t *testing.T) {
	other := NewModel(resend.SendEmailRequest{}, SMTP)
	m := NewModel(resend.SendEmailRequest{
		From: "sender@example.com", To: []string{"recipient@example.com"},
		Cc: []string{"cc@example.com"}, Bcc: []string{"bcc@example.com"},
		Subject: "A draft", Text: "Keep this draft\nand this line.",
	}, SMTP)
	m.setCommonWidths(80)
	m.Attachments.InsertItem(0, attachment("/tmp/invoice.pdf"))
	m.blurInputs()
	m.state = editingSubject
	m.focusActiveInput()
	m.Subject.SetCursor(2)
	beforeSubjectCursor := m.Subject.Position()
	beforeBodyCursor := m.Body.Cursor()
	beforeHelp := m.help.Styles

	if m.Init() == nil {
		t.Fatal("composer must request the terminal background")
	}
	updated, cmd := m.Update(tea.BackgroundColorMsg{Color: color.White})
	if cmd != nil {
		t.Fatal("background response must not start another command")
	}
	m = updated.(Model)
	for name, input := range map[string]textinput.Model{
		"from": m.From, "to": m.To, "cc": m.Cc, "bcc": m.Bcc, "subject": m.Subject,
	} {
		if got := input.Styles().Focused.Text.GetForeground(); !sameColor(got, charmtone.Pepper) {
			t.Errorf("%s focused text should use the light palette", name)
		}
		if got := input.Styles().Cursor.Color; !sameColor(got, charmtone.Pepper) {
			t.Errorf("%s cursor should use the light palette", name)
		}
	}
	if !sameColor(m.Body.Styles().Focused.Text.GetForeground(), charmtone.Pepper) {
		t.Error("body should use the light palette")
	}
	if reflect.DeepEqual(beforeHelp, m.help.Styles) {
		t.Error("help styles should adapt to the light background")
	}
	if m.From.Value() != "sender@example.com" || m.To.Value() != "recipient@example.com" ||
		m.Cc.Value() != "cc@example.com" || m.Bcc.Value() != "bcc@example.com" ||
		m.Subject.Value() != "A draft" || m.Body.Value() != "Keep this draft\nand this line." {
		t.Fatal("background response changed the draft")
	}
	if m.state != editingSubject || !m.Subject.Focused() || m.From.Focused() ||
		m.Subject.Position() != beforeSubjectCursor || !reflect.DeepEqual(m.Body.Cursor(), beforeBodyCursor) {
		t.Fatal("background response changed editing state or cursor position")
	}
	if len(m.Attachments.Items()) != 1 || m.Attachments.Items()[0].FilterValue() != "/tmp/invoice.pdf" {
		t.Fatal("background response changed attachments")
	}

	if !sameColor(other.From.Styles().Focused.Text.GetForeground(), charmtone.Soda) {
		t.Error("a light background response must not change another model")
	}

	// Subsequent responses must allow switching back.
	updated, _ = m.Update(tea.BackgroundColorMsg{Color: color.Black})
	m = updated.(Model)
	if !sameColor(m.From.Styles().Focused.Text.GetForeground(), charmtone.Soda) {
		t.Error("dark background should restore the existing text color")
	}
	if !sameColor(other.From.Styles().Focused.Text.GetForeground(), charmtone.Soda) {
		t.Error("a background response must not change another model")
	}
}

func TestComposerDarkFallback(t *testing.T) {
	m := NewModel(resend.SendEmailRequest{}, SMTP)
	if !sameColor(m.From.Styles().Focused.Text.GetForeground(), charmtone.Soda) {
		t.Fatal("without a background response, preserve the existing dark palette")
	}
}

func TestBackgroundPreservesAttachmentFocus(t *testing.T) {
	m := NewModel(resend.SendEmailRequest{}, SMTP)
	m.blurInputs()
	m.state = editingAttachments
	m.focusActiveInput()
	updated, _ := m.Update(tea.BackgroundColorMsg{Color: color.White})
	m = updated.(Model)
	if m.state != editingAttachments || !reflect.DeepEqual(m.Attachments.Styles.Title, attachmentsTitleActiveStyle) {
		t.Fatal("theme change lost attachment focus styling")
	}
	m.blurInputs()
	if !sameColor(m.Attachments.Styles.Title.GetForeground(), charmtone.Oyster) {
		t.Fatal("blurred attachment title should keep the light palette")
	}
}

func sameColor(a, b color.Color) bool {
	return color.RGBAModel.Convert(a) == color.RGBAModel.Convert(b)
}

func TestBackgroundPreservesBodyEditing(t *testing.T) {
	m := NewModel(resend.SendEmailRequest{Text: "first line\nsecond line"}, SMTP)
	m.setCommonWidths(80)
	m.blurInputs()
	m.state = editingBody
	m.focusActiveInput()
	m.Body.CursorDown()
	m.Body.CursorStart()
	before := m.Body.Cursor()
	beforeLine := m.Body.Line()
	beforeInfo := m.Body.LineInfo()
	updated, _ := m.Update(tea.BackgroundColorMsg{Color: color.White})
	m = updated.(Model)
	if !m.Body.Focused() || m.Body.Line() != beforeLine || !reflect.DeepEqual(m.Body.LineInfo(), beforeInfo) {
		t.Fatal("background response changed body editing state")
	}
	after := m.Body.Cursor()
	if before == nil || after == nil || before.X != after.X || before.Y != after.Y {
		t.Fatal("background response moved the body cursor")
	}
}
