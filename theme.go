package main

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/exp/charmtone"
)

// composerStyles contains the colors that depend on the terminal background.
// Keep these on the model so a background response does not change global styles.
type composerStyles struct {
	activeLabel     lipgloss.Style
	link            lipgloss.Style
	attachmentTitle lipgloss.Style
	activeText      lipgloss.Style
	text            lipgloss.Style
	label           lipgloss.Style
	placeholder     lipgloss.Style
}

func newComposerStyles(isDark bool) composerStyles {
	s := composerStyles{
		activeLabel:     activeLabelStyle.Foreground(charmtone.Hazy),
		link:            linkStyle,
		attachmentTitle: attachmentsTitleActiveStyle,
		activeText:      activeTextStyle,
		text:            textStyle,
		label:           labelStyle,
		placeholder:     placeholderStyle.Foreground(charmtone.Squid),
	}
	if !isDark {
		s.activeLabel = s.activeLabel.Foreground(charmtone.Jelly)
		s.link = s.link.Foreground(charmtone.Pepper)
		s.attachmentTitle = s.attachmentTitle.Foreground(charmtone.Jelly)
		s.activeText = s.activeText.Foreground(charmtone.Pepper)
		s.text = s.text.Foreground(charmtone.Iron)
		s.label = s.label.Foreground(charmtone.Oyster)
		s.placeholder = s.placeholder.Foreground(charmtone.Oyster)
	}
	return s
}

// applyTheme updates styles in place, preserving the draft and editing state.
func (m *Model) applyTheme(isDark bool) {
	m.applyComposerStyles(newComposerStyles(isDark), isDark)
}

// applyDefaultTheme uses the terminal foreground until its background is known.
// Assuming a dark background makes text unreadable on light terminals that do
// not answer OSC 11 queries.
func (m *Model) applyDefaultTheme() {
	m.applyComposerStyles(composerStyles{
		activeLabel:     lipgloss.NewStyle().Bold(true),
		link:            lipgloss.NewStyle().Underline(true),
		attachmentTitle: lipgloss.NewStyle().Bold(true),
		placeholder:     lipgloss.NewStyle().Italic(true),
	}, true)
}

func (m *Model) applyComposerStyles(styles composerStyles, isDark bool) {
	m.styles = styles
	cursorColor := m.styles.activeText.GetForeground()
	if _, inherited := cursorColor.(lipgloss.NoColor); inherited {
		cursorColor = nil
	}
	inputStyles := textinput.DefaultStyles(isDark)
	inputStyles.Focused.Prompt = m.styles.activeLabel
	inputStyles.Focused.Text = m.styles.activeText
	inputStyles.Focused.Placeholder = m.styles.placeholder
	inputStyles.Blurred.Prompt = m.styles.label
	inputStyles.Blurred.Text = m.styles.text
	inputStyles.Blurred.Placeholder = m.styles.placeholder
	inputStyles.Cursor.Color = cursorColor
	for _, input := range []*textinput.Model{&m.From, &m.To, &m.Cc, &m.Bcc, &m.Subject} {
		input.SetStyles(inputStyles)
	}

	bodyStyles := textarea.DefaultStyles(isDark)
	bodyStyles.Focused.CursorLine = m.styles.activeText
	bodyStyles.Focused.Prompt = m.styles.activeLabel
	bodyStyles.Focused.Text = m.styles.activeText
	bodyStyles.Focused.Placeholder = m.styles.placeholder
	bodyStyles.Blurred.CursorLine = m.styles.text
	bodyStyles.Blurred.Prompt = m.styles.label
	bodyStyles.Blurred.Text = m.styles.text
	bodyStyles.Blurred.Placeholder = m.styles.placeholder
	bodyStyles.Cursor.Color = cursorColor
	bodyStyles.Focused.EndOfBuffer = m.styles.placeholder
	bodyStyles.Blurred.EndOfBuffer = m.styles.placeholder
	m.Body.SetStyles(bodyStyles)
	m.help.Styles = help.Styles{
		ShortKey:       m.styles.label.Bold(true),
		ShortDesc:      m.styles.text,
		ShortSeparator: m.styles.label,
		FullKey:        m.styles.label.Bold(true),
		FullDesc:       m.styles.text,
		FullSeparator:  m.styles.label,
		Ellipsis:       m.styles.label,
	}

	focused := m.state == editingAttachments
	m.Attachments.Styles.Title = m.styles.label
	if focused {
		m.Attachments.Styles.Title = m.styles.attachmentTitle
	}
	m.Attachments.Styles.TitleBar = m.styles.label
	m.Attachments.Styles.NoItems = m.styles.placeholder
	m.Attachments.SetDelegate(attachmentDelegate{focused: focused, styles: m.styles})
}
