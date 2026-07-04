package main

import (
	"os"
	"strings"
	"time"

	"charm.land/bubbles/v2/filepicker"
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/exp/ordered"
	"github.com/resendlabs/resend-go"
)

// State is the current state of the application.
type State int

const (
	editingFrom State = iota
	editingTo
	editingCc
	editingBcc
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
	// Unknown is set when the user has not chosen a single delivery method.
	// i.e. multiple delivery methods are set.
	Unknown
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

	showCc bool
	Cc     textinput.Model
	Bcc    textinput.Model

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
	fromStyles := textinput.DefaultDarkStyles()
	fromStyles.Focused.Prompt = activeLabelStyle
	fromStyles.Focused.Text = activeTextStyle
	fromStyles.Focused.Placeholder = placeholderStyle
	fromStyles.Blurred.Prompt = labelStyle
	fromStyles.Blurred.Text = textStyle
	fromStyles.Blurred.Placeholder = placeholderStyle
	fromStyles.Cursor.Color = whiteColor
	from.SetStyles(fromStyles)
	from.SetVirtualCursor(false)
	from.SetValue(defaults.From)

	to := textinput.New()
	to.Prompt = "To "
	toStyles := textinput.DefaultDarkStyles()
	toStyles.Focused.Prompt = activeLabelStyle
	toStyles.Focused.Text = activeTextStyle
	toStyles.Focused.Placeholder = placeholderStyle
	toStyles.Blurred.Prompt = labelStyle
	toStyles.Blurred.Text = textStyle
	toStyles.Blurred.Placeholder = placeholderStyle
	toStyles.Cursor.Color = whiteColor
	to.SetStyles(toStyles)
	to.SetVirtualCursor(false)
	to.Placeholder = "you@example.com"
	to.SetValue(strings.Join(defaults.To, ToSeparator))

	cc := textinput.New()
	cc.Prompt = "Cc "
	ccStyles := textinput.DefaultDarkStyles()
	ccStyles.Focused.Prompt = activeLabelStyle
	ccStyles.Focused.Text = activeTextStyle
	ccStyles.Focused.Placeholder = placeholderStyle
	ccStyles.Blurred.Prompt = labelStyle
	ccStyles.Blurred.Text = textStyle
	ccStyles.Blurred.Placeholder = placeholderStyle
	ccStyles.Cursor.Color = whiteColor
	cc.SetStyles(ccStyles)
	cc.SetVirtualCursor(false)
	cc.Placeholder = "cc@example.com"
	cc.SetValue(strings.Join(defaults.Cc, ToSeparator))

	bcc := textinput.New()
	bcc.Prompt = "Bcc "
	bccStyles := textinput.DefaultDarkStyles()
	bccStyles.Focused.Prompt = activeLabelStyle
	bccStyles.Focused.Text = activeTextStyle
	bccStyles.Focused.Placeholder = placeholderStyle
	bccStyles.Blurred.Prompt = labelStyle
	bccStyles.Blurred.Text = textStyle
	bccStyles.Blurred.Placeholder = placeholderStyle
	bccStyles.Cursor.Color = whiteColor
	bcc.SetStyles(bccStyles)
	bcc.SetVirtualCursor(false)
	bcc.Placeholder = "bcc@example.com"
	bcc.SetValue(strings.Join(defaults.Bcc, ToSeparator))

	subject := textinput.New()
	subject.Prompt = "Subject "
	subjectStyles := textinput.DefaultDarkStyles()
	subjectStyles.Focused.Prompt = activeLabelStyle
	subjectStyles.Focused.Text = activeTextStyle
	subjectStyles.Focused.Placeholder = placeholderStyle
	subjectStyles.Blurred.Prompt = labelStyle
	subjectStyles.Blurred.Text = textStyle
	subjectStyles.Blurred.Placeholder = placeholderStyle
	subjectStyles.Cursor.Color = whiteColor
	subject.SetStyles(subjectStyles)
	subject.SetVirtualCursor(false)
	subject.Placeholder = "Hello!"
	subject.SetValue(defaults.Subject)

	body := textarea.New()
	body.Placeholder = "# Email"
	body.ShowLineNumbers = false
	bodyStyles := textarea.DefaultDarkStyles()
	bodyStyles.Focused.CursorLine = activeTextStyle
	bodyStyles.Focused.Prompt = activeLabelStyle
	bodyStyles.Focused.Text = activeTextStyle
	bodyStyles.Focused.Placeholder = placeholderStyle
	bodyStyles.Blurred.CursorLine = textStyle
	bodyStyles.Blurred.Prompt = labelStyle
	bodyStyles.Blurred.Text = textStyle
	bodyStyles.Blurred.Placeholder = placeholderStyle
	bodyStyles.Cursor.Color = whiteColor
	body.SetStyles(bodyStyles)
	body.SetVirtualCursor(false)
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
		showCc:         len(cc.Value()) > 0 || len(bcc.Value()) > 0,
		Cc:             cc,
		Bcc:            bcc,
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
	return nil
}

type clearErrMsg struct{}

func clearErrAfter(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(_ time.Time) tea.Msg {
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
		return m, clearErrAfter(10 * time.Second)
	case clearErrMsg:
		m.err = nil
	case tea.WindowSizeMsg:
		m.setCommonWidths(msg.Width)
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.keymap.NextInput):
			m.blurInputs()
			switch m.state {
			case editingFrom:
				m.state = editingTo
				m.To.Focus()
			case editingTo:
				if m.showCc {
					m.state = editingCc
				} else {
					m.state = editingSubject
				}
			case editingCc:
				m.state = editingBcc
			case editingBcc:
				m.state = editingSubject
			case editingSubject:
				m.state = editingBody
			case editingBody:
				m.state = editingAttachments
			case editingAttachments:
				m.state = hoveringSendButton
			case hoveringSendButton:
				m.state = editingFrom
			case pickingFile, sendingEmail:
			}
			m.focusActiveInput()

		case key.Matches(msg, m.keymap.PrevInput):
			m.blurInputs()
			switch m.state {
			case editingFrom:
				m.state = hoveringSendButton
			case editingTo:
				m.state = editingFrom
			case editingCc:
				m.state = editingTo
			case editingBcc:
				m.state = editingCc
			case editingSubject:
				if m.showCc {
					m.state = editingBcc
				} else {
					m.state = editingTo
				}
			case editingBody:
				m.state = editingSubject
			case editingAttachments:
				m.state = editingBody
			case hoveringSendButton:
				m.state = editingAttachments
			case pickingFile, sendingEmail:
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
	if m.showCc {
		m.Cc, cmd = m.Cc.Update(msg)
		cmds = append(cmds, cmd)
		m.Bcc, cmd = m.Bcc.Update(msg)
		cmds = append(cmds, cmd)
	}
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
	case editingFrom, editingTo, editingCc, editingBcc, editingSubject, editingBody, hoveringSendButton:
	}

	m.help, cmd = m.help.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *Model) setCommonWidths(width int) {
	inputWidth := width - paddedStyle.GetHorizontalFrameSize() - lipgloss.Width("Subject ")
	m.From.SetWidth(inputWidth)
	m.To.SetWidth(inputWidth)
	m.Cc.SetWidth(inputWidth)
	m.Bcc.SetWidth(inputWidth)
	m.Subject.SetWidth(inputWidth)
	m.Body.SetWidth(width - paddedStyle.GetHorizontalFrameSize())
	m.Attachments.SetWidth(width - paddedStyle.GetHorizontalFrameSize())
	m.help.SetWidth(width - paddedStyle.GetHorizontalFrameSize())
}

func (m *Model) blurInputs() {
	m.From.Blur()
	m.To.Blur()
	m.Subject.Blur()
	m.Body.Blur()
	if m.showCc {
		m.Cc.Blur()
		m.Bcc.Blur()
	}
	m.Attachments.Styles.Title = labelStyle
	m.Attachments.SetDelegate(attachmentDelegate{false})
}

func (m *Model) focusActiveInput() {
	switch m.state {
	case editingFrom:
		m.From.Focus()
		m.From.CursorEnd()
	case editingTo:
		m.To.Focus()
		m.To.CursorEnd()
	case editingCc:
		m.Cc.Focus()
		m.Cc.CursorEnd()
	case editingBcc:
		m.Bcc.Focus()
		m.Bcc.CursorEnd()
	case editingSubject:
		m.Subject.Focus()
		m.Subject.CursorEnd()
	case editingBody:
		m.Body.Focus()
		m.Body.CursorEnd()
	case editingAttachments:
		m.Attachments.Styles.Title = activeLabelStyle
		m.Attachments.SetDelegate(attachmentDelegate{true})
	case hoveringSendButton, pickingFile, sendingEmail:
	}
}

// View displays the application.
func (m Model) View() tea.View {
	if m.quitting {
		return tea.NewView("")
	}

	switch m.state {
	case pickingFile:
		return tea.NewView("\n" + activeLabelStyle.Render("Attachments") + " " + commentStyle.Render(m.filepicker.CurrentDirectory) +
			"\n\n" + m.filepicker.View())
	case sendingEmail:
		return tea.NewView("\n " + m.loadingSpinner.View() + "Sending email")
	case editingFrom, editingTo, editingCc, editingBcc, editingSubject, editingBody, editingAttachments, hoveringSendButton:
	}

	var s strings.Builder

	s.WriteString(m.From.View())
	s.WriteString("\n")
	s.WriteString(m.To.View())
	s.WriteString("\n")
	ccLines := 0
	if m.showCc {
		s.WriteString(m.Cc.View())
		s.WriteString("\n")
		s.WriteString(m.Bcc.View())
		s.WriteString("\n")
		ccLines = 2
	}
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

	v := tea.NewView(paddedStyle.Render(s.String()))

	// Position the real cursor based on which field is focused.
	padY := paddedStyle.GetPaddingTop()
	padX := paddedStyle.GetPaddingLeft()
	switch m.state {
	case editingFrom:
		if c := m.From.Cursor(); c != nil {
			c.Position.Y += padY
			c.Position.X += padX
			v.Cursor = c
		}
	case editingTo:
		if c := m.To.Cursor(); c != nil {
			c.Position.Y += padY + 1
			c.Position.X += padX
			v.Cursor = c
		}
	case editingCc:
		if c := m.Cc.Cursor(); c != nil {
			c.Position.Y += padY + 2
			c.Position.X += padX
			v.Cursor = c
		}
	case editingBcc:
		if c := m.Bcc.Cursor(); c != nil {
			c.Position.Y += padY + 3
			c.Position.X += padX
			v.Cursor = c
		}
	case editingSubject:
		if c := m.Subject.Cursor(); c != nil {
			c.Position.Y += padY + ccLines + 2
			c.Position.X += padX
			v.Cursor = c
		}
	case editingBody:
		if c := m.Body.Cursor(); c != nil {
			c.Position.Y += padY + ccLines + 4
			c.Position.X += padX
			v.Cursor = c
		}
	}

	return v
}
