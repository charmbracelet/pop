package main

import (
	"image/color"
	"reflect"
	"strings"
	"testing"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/resendlabs/resend-go"
)

func TestComposerIgnoresMisreportedBackground(t *testing.T) {
	m := NewModel(resend.SendEmailRequest{}, SMTP)
	before := m.styles
	// A light response can arrive while an embedded terminal is visibly dark,
	// and vice versa. Neither may replace the user's terminal foreground.
	for _, background := range []color.Color{color.White, color.Black} {
		updated, _ := m.Update(tea.BackgroundColorMsg{Color: background})
		m = updated.(Model)
		if !reflect.DeepEqual(before, m.styles) {
			t.Fatal("background response overrode terminal-native styles")
		}
		for name, input := range map[string]textinput.Model{"from": m.From, "to": m.To, "cc": m.Cc, "bcc": m.Bcc, "subject": m.Subject} {
			s := input.Styles()
			for _, style := range []lipgloss.Style{s.Focused.Prompt, s.Blurred.Prompt, s.Focused.Text, s.Blurred.Text, s.Focused.Placeholder, s.Blurred.Placeholder} {
				if _, inherited := style.GetForeground().(lipgloss.NoColor); !inherited {
					t.Errorf("%s must inherit foreground", name)
				}
				if _, inherited := style.GetBackground().(lipgloss.NoColor); !inherited {
					t.Errorf("%s must inherit background", name)
				}
			}
			if !s.Focused.Prompt.GetBold() || !s.Focused.Placeholder.GetItalic() {
				t.Errorf("%s should distinguish focus and placeholders without dimming", name)
			}
		}
	}
}

func TestSummaryPreservesTerminalColors(t *testing.T) {
	recipients := []string{"first@example.test", "second@example.test"}
	const subject = "Theme test"
	const want = "\n  Email \"Theme test\" sent to first@example.test, second@example.test\n\n"
	m := NewModel(resend.SendEmailRequest{}, SMTP)
	summaries := []string{emailSummary(recipients, subject), m.styles.emailSummary(recipients, subject)}
	for _, background := range []color.Color{color.White, color.Black} {
		updated, _ := m.Update(tea.BackgroundColorMsg{Color: background})
		m = updated.(Model)
		summaries = append(summaries, m.styles.emailSummary(recipients, subject))
	}
	for _, summary := range summaries {
		if ansi.Strip(summary) != want {
			t.Fatalf("unexpected summary: %q", summary)
		}
		if strings.Contains(summary, "38;") || strings.Contains(summary, "48;") {
			t.Fatalf("summary overrides terminal colors: %q", summary)
		}
	}
}
