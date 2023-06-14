package main

import "github.com/charmbracelet/bubbles/key"

// KeyMap represents the key bindings for the application.
type KeyMap struct {
	NextInput key.Binding
	PrevInput key.Binding
	Send      key.Binding
	Attach    key.Binding
	Unattach  key.Binding
	Quit      key.Binding
}

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
			key.WithKeys("a"),
			key.WithHelp("a", "attach file"),
			key.WithDisabled(),
		),
		Unattach: key.NewBinding(
			key.WithKeys("x"),
			key.WithHelp("x", "remove"),
			key.WithDisabled(),
		),
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),
	}
}

func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.NextInput,
		k.Quit,
		k.Attach,
		k.Unattach,
		k.Send,
	}
}

func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.NextInput, k.Send, k.Attach, k.Unattach, k.Quit},
	}
}

func (m *Model) updateKeymap() {
	m.keymap.Attach.SetEnabled(m.state == editingAttachments)
	canSend := m.From.Value() != "" && m.To.Value() != "" && m.Subject.Value() != "" && m.Body.Value() != ""
	m.keymap.Send.SetEnabled(canSend && m.state != editingBody)
	m.keymap.Unattach.SetEnabled(m.state == editingAttachments && len(m.Attachments.Items()) > 0)
}
