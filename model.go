package main

import (
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/ordered"
	"github.com/resendlabs/resend-go"
)

// State is the current state of the application.
type State int

const (
	editingFrom State = iota
	editingTo
	editingSubject
	editingBody
	editingAttachments
	hoveringSendButton
	pickingFile
	sendingEmail
)

// DeliveryMethod is the method of delivery for the email.
type DeliveryMethod int

const (
	// None is the default delivery method.
	None DeliveryMethod = iota
	// Resend uses https://resend.com to send an email.
	Resend
	// SMTP uses an SMTP server to send an email.
	SMTP
)

// Model is Pop's application model.
type Model struct {
	// state represents the current state of the application.
	state State

	// DeliveryMethod is whether we are using DeliveryMethod or Resend.
	DeliveryMethod DeliveryMethod

	// From represents the sender's email address.
	From textinput.Model
	// To represents the recipient's email address.
	// This can be a comma-separated list of addresses.
	To textinput.Model
	// Subject represents the email's subject.
	Subject textinput.Model
	// Body represents the email's body.
	// This can be written in markdown and will be converted to HTML.
	Body textarea.Model
	// Attachments represents the email's attachments.
	// This is a list of file paths which are picked with a filepicker.
	Attachments list.Model

	Cc  textinput.Model
	Bcc textinput.Model

	// filepicker is used to pick file attachments.
	filepicker     filepicker.Model
	loadingSpinner spinner.Model
	help           help.Model
	keymap         KeyMap
	quitting       bool
	abort          bool
	err            error
}

// NewModel returns a new model for the application.
func NewModel(defaults resend.SendEmailRequest, deliveryMethod DeliveryMethod) Model {
	from := textinput.New()
	from.Prompt = "From "
	from.Placeholder = "me@example.com"
	from.PromptStyle = labelStyle.Copy()
	from.PromptStyle = labelStyle
	from.TextStyle = textStyle
	from.Cursor.Style = cursorStyle
	from.PlaceholderStyle = placeholderStyle
	from.SetValue(defaults.From)

	to := textinput.New()
	to.Prompt = "To "
	to.PromptStyle = labelStyle.Copy()
	to.Cursor.Style = cursorStyle
	to.PlaceholderStyle = placeholderStyle
	to.TextStyle = textStyle
	to.Placeholder = "you@example.com"
	to.SetValue(strings.Join(defaults.To, ToSeparator))

	subject := textinput.New()
	subject.Prompt = "Subject "
	subject.PromptStyle = labelStyle.Copy()
	subject.Cursor.Style = cursorStyle
	subject.PlaceholderStyle = placeholderStyle
	subject.TextStyle = textStyle
	subject.Placeholder = "Hello!"
	subject.SetValue(defaults.Subject)

	body := textarea.New()
	body.Placeholder = "# Email"
	body.ShowLineNumbers = false
	body.FocusedStyle.CursorLine = activeTextStyle
	body.FocusedStyle.Prompt = activeLabelStyle
	body.FocusedStyle.Text = activeTextStyle
	body.FocusedStyle.Placeholder = placeholderStyle
	body.BlurredStyle.CursorLine = textStyle
	body.BlurredStyle.Prompt = labelStyle
	body.BlurredStyle.Text = textStyle
	body.BlurredStyle.Placeholder = placeholderStyle
	body.Cursor.Style = cursorStyle
	body.CharLimit = 4000
	body.SetValue(defaults.Text)

	// Adjust for signature (if none, this is a no-op)
	body.CursorUp()
	body.CursorUp()

	body.Blur()

	// Decide which input to focus.
	var state State
	switch {
	case defaults.From == "":
		state = editingFrom
	case len(defaults.To) == 0:
		state = editingTo
	case defaults.Subject == "":
		state = editingSubject
	case defaults.Text == "":
		state = editingBody
	}

	attachments := list.New([]list.Item{}, attachmentDelegate{}, 0, 3)
	attachments.DisableQuitKeybindings()
	attachments.SetShowTitle(true)
	attachments.Title = "Attachments"
	attachments.Styles.Title = labelStyle
	attachments.Styles.TitleBar = labelStyle
	attachments.Styles.NoItems = placeholderStyle
	attachments.SetShowHelp(false)
	attachments.SetShowStatusBar(false)
	attachments.SetStatusBarItemName("attachment", "attachments")
	attachments.SetShowPagination(false)

	for _, a := range defaults.Attachments {
		attachments.InsertItem(0, attachment(a.Filename))
	}

	picker := filepicker.New()
	picker.CurrentDirectory, _ = os.UserHomeDir()

	loadingSpinner := spinner.New()
	loadingSpinner.Style = activeLabelStyle
	loadingSpinner.Spinner = spinner.Dot

	m := Model{
		state:          state,
		From:           from,
		To:             to,
		Subject:        subject,
		Body:           body,
		Attachments:    attachments,
		filepicker:     picker,
		help:           help.New(),
		keymap:         DefaultKeybinds(),
		loadingSpinner: loadingSpinner,
		DeliveryMethod: deliveryMethod,
	}

	m.focusActiveInput()

	return m
}

// Init initializes the model.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.From.Cursor.BlinkCmd(),
	)
}

type clearErrMsg struct{}

func clearErrAfter(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return clearErrMsg{}
	})
}

// Update is the update loop for the model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case sendEmailSuccessMsg:
		m.quitting = true
		return m, tea.Quit
	case sendEmailFailureMsg:
		m.blurInputs()
		m.state = editingFrom
		m.focusActiveInput()
		m.err = msg
		return m, clearErrAfter(5 * time.Second)
	case clearErrMsg:
		m.err = nil
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keymap.NextInput):
			m.blurInputs()
			switch m.state {
			case editingFrom:
				m.state = editingTo
				m.To.Focus()
			case editingTo:
				m.state = editingSubject
			case editingSubject:
				m.state = editingBody
			case editingBody:
				m.state = editingAttachments
			case editingAttachments:
				m.state = hoveringSendButton
			case hoveringSendButton:
				m.state = editingFrom
			}
			m.focusActiveInput()

		case key.Matches(msg, m.keymap.PrevInput):
			m.blurInputs()
			switch m.state {
			case editingFrom:
				m.state = hoveringSendButton
			case editingTo:
				m.state = editingFrom
			case editingSubject:
				m.state = editingTo
			case editingBody:
				m.state = editingSubject
			case editingAttachments:
				m.state = editingBody
			case hoveringSendButton:
				m.state = editingAttachments
			}
			m.focusActiveInput()

		case key.Matches(msg, m.keymap.Back):
			m.state = editingAttachments
			m.updateKeymap()
			return m, nil
		case key.Matches(msg, m.keymap.Send):
			m.state = sendingEmail
			return m, tea.Batch(
				m.loadingSpinner.Tick,
				m.sendEmailCmd(),
			)
		case key.Matches(msg, m.keymap.Attach):
			m.state = pickingFile
			return m, m.filepicker.Init()
		case key.Matches(msg, m.keymap.Unattach):
			m.Attachments.RemoveItem(m.Attachments.Index())
			m.Attachments.SetHeight(ordered.Max(len(m.Attachments.Items()), 1) + 2)
		case key.Matches(msg, m.keymap.Quit):
			m.quitting = true
			m.abort = true
			return m, tea.Quit
		}
	}

	m.updateKeymap()

	var cmds []tea.Cmd
	var cmd tea.Cmd
	m.From, cmd = m.From.Update(msg)
	cmds = append(cmds, cmd)
	m.To, cmd = m.To.Update(msg)
	cmds = append(cmds, cmd)
	m.Subject, cmd = m.Subject.Update(msg)
	cmds = append(cmds, cmd)
	m.Body, cmd = m.Body.Update(msg)
	cmds = append(cmds, cmd)
	m.filepicker, cmd = m.filepicker.Update(msg)
	cmds = append(cmds, cmd)

	switch m.state {
	case pickingFile:
		if didSelect, path := m.filepicker.DidSelectFile(msg); didSelect {
			m.Attachments.InsertItem(0, attachment(path))
			m.Attachments.SetHeight(len(m.Attachments.Items()) + 2)
			m.state = editingAttachments
			m.updateKeymap()
		}
	case editingAttachments:
		m.Attachments, cmd = m.Attachments.Update(msg)
		cmds = append(cmds, cmd)
	case sendingEmail:
		m.loadingSpinner, cmd = m.loadingSpinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	m.help, cmd = m.help.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *Model) blurInputs() {
	m.From.Blur()
	m.To.Blur()
	m.Subject.Blur()
	m.Body.Blur()
	m.From.PromptStyle = labelStyle
	m.To.PromptStyle = labelStyle
	m.Subject.PromptStyle = labelStyle
	m.From.TextStyle = textStyle
	m.To.TextStyle = textStyle
	m.Subject.TextStyle = textStyle
	m.Attachments.Styles.Title = labelStyle
	m.Attachments.SetDelegate(attachmentDelegate{false})
}

func (m *Model) focusActiveInput() {
	switch m.state {
	case editingFrom:
		m.From.PromptStyle = activeLabelStyle
		m.From.TextStyle = activeTextStyle
		m.From.Focus()
		m.From.CursorEnd()
	case editingTo:
		m.To.PromptStyle = activeLabelStyle
		m.To.TextStyle = activeTextStyle
		m.To.Focus()
		m.To.CursorEnd()
	case editingSubject:
		m.Subject.PromptStyle = activeLabelStyle
		m.Subject.TextStyle = activeTextStyle
		m.Subject.Focus()
		m.Subject.CursorEnd()
	case editingBody:
		m.Body.Focus()
		m.Body.CursorEnd()
	case editingAttachments:
		m.Attachments.Styles.Title = activeLabelStyle
		m.Attachments.SetDelegate(attachmentDelegate{true})
	}
}

// View displays the application.
func (m Model) View() string {
	if m.quitting {
		return ""
	}

	switch m.state {
	case pickingFile:
		return "\n" + activeLabelStyle.Render("Attachments") + " " + commentStyle.Render(m.filepicker.CurrentDirectory) +
			"\n\n" + m.filepicker.View()
	case sendingEmail:
		return "\n " + m.loadingSpinner.View() + "Sending email"
	}

	var s strings.Builder

	s.WriteString(m.From.View())
	s.WriteString("\n")
	s.WriteString(m.To.View())
	s.WriteString("\n")
	s.WriteString(m.Subject.View())
	s.WriteString("\n\n")
	s.WriteString(m.Body.View())
	s.WriteString("\n\n")
	s.WriteString(m.Attachments.View())
	s.WriteString("\n")
	if m.state == hoveringSendButton && m.canSend() {
		s.WriteString(sendButtonActiveStyle.Render("Send"))
	} else if m.state == hoveringSendButton {
		s.WriteString(sendButtonInactiveStyle.Render("Send"))
	} else {
		s.WriteString(sendButtonStyle.Render("Send"))
	}
	s.WriteString("\n\n")
	s.WriteString(m.help.View(m.keymap))

	if m.err != nil {
		s.WriteString("\n\n")
		s.WriteString(errorStyle.Render(m.err.Error()))
	}

	return paddedStyle.Render(s.String())
}
