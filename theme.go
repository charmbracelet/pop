package main

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	"charm.land/lipgloss/v2"
)

// composerStyles groups terminal-native styles for one composer.
type composerStyles struct {
	activeLabel     lipgloss.Style
	link            lipgloss.Style
	attachmentTitle lipgloss.Style
	activeText      lipgloss.Style
	text            lipgloss.Style
	label           lipgloss.Style
	placeholder     lipgloss.Style
}

func newComposerStyles() composerStyles {
	// Use the terminal's foreground on every background. Some embedded terminals
	// report the application's theme rather than the visible terminal background.
	// Faint labels and faint italic placeholders keep hierarchy on any background
	// without dimming real values.
	return composerStyles{
		activeLabel:     activeLabelStyle,
		link:            linkStyle.UnsetForeground(),
		attachmentTitle: attachmentsTitleActiveStyle,
		activeText:      activeTextStyle.UnsetForeground(),
		text:            textStyle.UnsetForeground(),
		label:           labelStyle.UnsetForeground().Faint(true),
		placeholder:     placeholderStyle.UnsetForeground().Faint(true).Italic(true),
	}
}

// applyTheme uses terminal-native colors, preserving contrast even when a
// background query is unanswered or returns a stale or incorrect color.
func (m *Model) applyTheme() {
	m.styles = newComposerStyles()
	inputStyles := textinput.DefaultStyles(true)
	inputStyles.Focused.Prompt = m.styles.activeLabel
	inputStyles.Focused.Text = m.styles.activeText
	inputStyles.Focused.Placeholder = m.styles.placeholder
	inputStyles.Blurred.Prompt = m.styles.label
	inputStyles.Blurred.Text = m.styles.text
	inputStyles.Blurred.Placeholder = m.styles.placeholder
	inputStyles.Cursor.Color = nil
	for _, input := range []*textinput.Model{&m.From, &m.To, &m.Cc, &m.Bcc, &m.Subject} {
		input.SetStyles(inputStyles)
	}

	bodyStyles := textarea.DefaultStyles(true)
	bodyStyles.Focused.CursorLine = m.styles.activeText
	bodyStyles.Focused.Prompt = m.styles.activeLabel
	bodyStyles.Focused.Text = m.styles.activeText
	bodyStyles.Focused.Placeholder = m.styles.placeholder
	bodyStyles.Blurred.CursorLine = m.styles.text
	bodyStyles.Blurred.Prompt = m.styles.label
	bodyStyles.Blurred.Text = m.styles.text
	bodyStyles.Blurred.Placeholder = m.styles.placeholder
	bodyStyles.Cursor.Color = nil
	bodyStyles.Focused.EndOfBuffer = m.styles.placeholder
	bodyStyles.Blurred.EndOfBuffer = m.styles.placeholder
	m.Body.SetStyles(bodyStyles)
	m.help.Styles = help.Styles{
		ShortKey:       m.styles.label,
		ShortDesc:      m.styles.text,
		ShortSeparator: m.styles.label,
		FullKey:        m.styles.label,
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
