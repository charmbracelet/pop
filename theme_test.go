package main

import (
	"image/color"
	"reflect"
	"testing"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
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

	if m.Init() != nil {
		t.Fatal("composer must not depend on background detection")
	}
	updated, cmd := m.Update(tea.BackgroundColorMsg{Color: color.White})
	if cmd != nil {
		t.Fatal("background response must not start another command")
	}
	m = updated.(Model)
	for name, input := range map[string]textinput.Model{"from": m.From, "to": m.To, "cc": m.Cc, "bcc": m.Bcc, "subject": m.Subject} {
		if _, inherited := input.Styles().Focused.Text.GetForeground().(lipgloss.NoColor); !inherited {
			t.Errorf("%s must inherit foreground despite background response", name)
		}
		if input.Styles().Cursor.Color != nil {
			t.Errorf("%s must retain terminal cursor color", name)
		}
	}
	if _, inherited := m.Body.Styles().Focused.Text.GetForeground().(lipgloss.NoColor); !inherited {
		t.Error("body must inherit terminal foreground")
	}
	if !reflect.DeepEqual(beforeHelp, m.help.Styles) {
		t.Error("background response must not override terminal-native help")
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

	if _, inherited := other.From.Styles().Focused.Text.GetForeground().(lipgloss.NoColor); !inherited {
		t.Error("a light background response must not change another model")
	}

	// Subsequent responses must allow switching back.
	updated, _ = m.Update(tea.BackgroundColorMsg{Color: color.Black})
	m = updated.(Model)
	if _, inherited := m.From.Styles().Focused.Text.GetForeground().(lipgloss.NoColor); !inherited {
		t.Error("dark background response must not override terminal foreground")
	}
	if _, inherited := other.From.Styles().Focused.Text.GetForeground().(lipgloss.NoColor); !inherited {
		t.Error("a background response must not change another model")
	}
}

func TestComposerInheritedFallback(t *testing.T) {
	m := NewModel(resend.SendEmailRequest{}, SMTP)
	for name, input := range map[string]textinput.Model{"from": m.From, "to": m.To, "cc": m.Cc, "bcc": m.Bcc, "subject": m.Subject} {
		styles := input.Styles()
		for _, style := range []lipgloss.Style{styles.Focused.Text, styles.Blurred.Text, styles.Focused.Prompt, styles.Blurred.Prompt, styles.Focused.Placeholder, styles.Blurred.Placeholder} {
			if _, inherited := style.GetForeground().(lipgloss.NoColor); !inherited {
				t.Errorf("%s assumes a background before detection", name)
			}
		}
		if styles.Cursor.Color != nil {
			t.Errorf("%s overrides the default cursor color", name)
		}
	}
	for _, style := range []lipgloss.Style{m.Body.Styles().Focused.Text, m.Body.Styles().Blurred.Text, m.Body.Styles().Focused.Placeholder, m.Body.Styles().Blurred.Placeholder, m.help.Styles.ShortKey, m.help.Styles.ShortDesc, m.Attachments.Styles.NoItems, m.styles.link} {
		if _, inherited := style.GetForeground().(lipgloss.NoColor); !inherited {
			t.Error("unknown background must inherit the terminal foreground")
		}
	}
	if m.Body.Styles().Cursor.Color != nil {
		t.Error("body overrides the default cursor color")
	}
}

func TestBackgroundPreservesAttachmentFocus(t *testing.T) {
	m := NewModel(resend.SendEmailRequest{}, SMTP)
	m.blurInputs()
	m.state = editingAttachments
	m.focusActiveInput()
	updated, _ := m.Update(tea.BackgroundColorMsg{Color: color.White})
	m = updated.(Model)
	if m.state != editingAttachments || !reflect.DeepEqual(m.Attachments.Styles.Title, m.styles.attachmentTitle) {
		t.Fatal("theme change lost attachment focus styling")
	}
	m.blurInputs()
	if _, inherited := m.Attachments.Styles.Title.GetForeground().(lipgloss.NoColor); !inherited {
		t.Fatal("blurred attachment title should inherit terminal foreground")
	}
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
