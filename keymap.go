package main

import "github.com/charmbracelet/bubbles/key"

// KeyMap represents the key bindings for the application.
type KeyMap struct {
	NextInput key.Binding
	PrevInput key.Binding
	Send      key.Binding
	Attach    key.Binding
	Unattach  key.Binding
	Back      key.Binding
	Quit      key.Binding
}

// DefaultKeybinds returns the default key bindings for the application.
func DefaultKeybinds() KeyMap {
	return KeyMap{
		NextInput: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next"),
		),
		PrevInput: key.NewBinding(
			key.WithKeys("shift+tab"),
		),
		Send: key.NewBinding(
			key.WithKeys("ctrl+d", "enter"),
			key.WithHelp("enter", "send"),
			key.WithDisabled(),
		),
		Attach: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "attach file"),
			key.WithDisabled(),
		),
		Unattach: key.NewBinding(
			key.WithKeys("x"),
			key.WithHelp("x", "remove"),
			key.WithDisabled(),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
			key.WithDisabled(),
		),
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),
	}
}

// ShortHelp returns the key bindings for the short help screen.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.NextInput,
		k.Quit,
		k.Attach,
		k.Unattach,
		k.Send,
	}
}

// FullHelp returns the key bindings for the full help screen.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.NextInput, k.Send, k.Attach, k.Unattach, k.Quit},
	}
}

func (m *Model) updateKeymap() {
	m.keymap.Attach.SetEnabled(m.state == editingAttachments)
	m.keymap.Send.SetEnabled(m.canSend() && m.state == hoveringSendButton)
	m.keymap.Unattach.SetEnabled(m.state == editingAttachments && len(m.Attachments.Items()) > 0)
	m.keymap.Back.SetEnabled(m.state == pickingFile)

	m.filepicker.KeyMap.Up.SetEnabled(m.state == pickingFile)
	m.filepicker.KeyMap.Down.SetEnabled(m.state == pickingFile)
	m.filepicker.KeyMap.Back.SetEnabled(m.state == pickingFile)
	m.filepicker.KeyMap.Select.SetEnabled(m.state == pickingFile)
	m.filepicker.KeyMap.Open.SetEnabled(m.state == pickingFile)
	m.filepicker.KeyMap.PageUp.SetEnabled(m.state == pickingFile)
	m.filepicker.KeyMap.PageDown.SetEnabled(m.state == pickingFile)
	m.filepicker.KeyMap.GoToTop.SetEnabled(m.state == pickingFile)
	m.filepicker.KeyMap.GoToLast.SetEnabled(m.state == pickingFile)
}

func (m Model) canSend() bool {
	return m.From.Value() != "" && m.To.Value() != "" && m.Subject.Value() != "" && m.Body.Value() != ""
}
