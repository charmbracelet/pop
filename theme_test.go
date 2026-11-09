package main

import (
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/resendlabs/resend-go"
)

func TestComposerUsesTerminalColors(t *testing.T) {
	m := NewModel(resend.SendEmailRequest{}, SMTP)
	s := m.From.Styles()
	for _, style := range []lipgloss.Style{s.Focused.Text, s.Blurred.Text, s.Blurred.Prompt, s.Focused.Placeholder, m.Body.Styles().Focused.Text} {
		if _, inherited := style.GetForeground().(lipgloss.NoColor); !inherited {
			t.Error("composer must inherit the terminal foreground")
		}
	}
	if !s.Blurred.Placeholder.GetFaint() || !s.Blurred.Prompt.GetFaint() || s.Blurred.Text.GetFaint() {
		t.Error("placeholders and labels must be dimmer than real values")
	}
}
