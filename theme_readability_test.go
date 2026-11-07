package main

import (
	"image/color"
	"strings"
	"testing"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/exp/charmtone"
	"github.com/resendlabs/resend-go"
)

func TestThemeLabelsAndPlaceholders(t *testing.T) {
	for _, tc := range []struct {
		name                                   string
		background, label, active, placeholder color.Color
	}{
		{"light", color.White, charmtone.Oyster, charmtone.Jelly, charmtone.Oyster},
		{"dark", color.Black, charmtone.Steam, charmtone.Hazy, charmtone.Squid},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := NewModel(resend.SendEmailRequest{}, SMTP)
			updated, _ := m.Update(tea.BackgroundColorMsg{Color: tc.background})
			m = updated.(Model)
			for name, input := range map[string]textinput.Model{"from": m.From, "to": m.To, "cc": m.Cc, "bcc": m.Bcc, "subject": m.Subject} {
				s := input.Styles()
				if !sameColor(s.Focused.Prompt.GetForeground(), tc.active) || !sameColor(s.Blurred.Prompt.GetForeground(), tc.label) ||
					!sameColor(s.Focused.Placeholder.GetForeground(), tc.placeholder) || !sameColor(s.Blurred.Placeholder.GetForeground(), tc.placeholder) {
					t.Errorf("%s labels or placeholders use the wrong palette", name)
				}
			}
			s := m.Body.Styles()
			if !sameColor(s.Focused.Prompt.GetForeground(), tc.active) || !sameColor(s.Blurred.Prompt.GetForeground(), tc.label) ||
				!sameColor(s.Focused.Placeholder.GetForeground(), tc.placeholder) || !sameColor(s.Blurred.Placeholder.GetForeground(), tc.placeholder) {
				t.Error("body labels or placeholders use the wrong palette")
			}
			m.blurInputs()
			m.state = editingAttachments
			m.focusActiveInput()
			if !sameColor(m.Attachments.Styles.Title.GetForeground(), m.styles.attachmentTitle.GetForeground()) {
				t.Error("focused attachment title lost its palette")
			}
		})
	}
}

func TestSummaryUsesDetectedTheme(t *testing.T) {
	recipients := []string{"first@example.test", "second@example.test"}
	const subject = "Theme test"
	const want = "\n  Email \"Theme test\" sent to first@example.test, second@example.test\n\n"
	for _, tc := range []struct {
		name                           string
		background, subject, recipient color.Color
	}{
		{"light", color.White, charmtone.Pepper, charmtone.Pepper},
		{"dark", color.Black, charmtone.Soda, charmtone.Guac},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := NewModel(resend.SendEmailRequest{}, SMTP)
			updated, _ := m.Update(tea.BackgroundColorMsg{Color: tc.background})
			m = updated.(Model)
			got := m.styles.emailSummary(recipients, subject)
			if ansi.Strip(got) != want {
				t.Fatalf("unexpected summary: %q", got)
			}
			if !sameColor(m.styles.activeText.GetForeground(), tc.subject) || !sameColor(m.styles.link.GetForeground(), tc.recipient) {
				t.Error("summary palette is inappropriate for the detected background")
			}
			if !strings.Contains(got, m.styles.activeText.Render("\""+subject+"\"")) || !strings.Contains(got, m.styles.link.Render(recipients[0])) {
				t.Error("summary does not apply the theme to its subject and recipients")
			}
		})
	}
	cli := emailSummary(recipients, subject)
	if ansi.Strip(cli) != want || strings.Contains(cli, "38;") {
		t.Errorf("CLI summary should inherit foreground: %q", cli)
	}
}

func TestSummaryWithoutBackgroundResponse(t *testing.T) {
	m := NewModel(resend.SendEmailRequest{}, SMTP)
	summary := m.styles.emailSummary([]string{"recipient@example.test"}, "Subject")
	if strings.Contains(summary, "38;") || strings.Contains(summary, "48;") {
		t.Fatalf("unknown-background summary forces a color: %q", summary)
	}
	if ansi.Strip(summary) != "\n  Email \"Subject\" sent to recipient@example.test\n\n" {
		t.Fatalf("unexpected summary: %q", summary)
	}
}
