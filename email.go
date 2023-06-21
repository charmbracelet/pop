package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/resendlabs/resend-go"
	"github.com/yuin/goldmark"
)

const TO_SEPARATOR = ","

// sendEmailSuccessMsg is the tea.Msg handled by Bubble Tea when the email has
// been sent successfully.
type sendEmailSuccessMsg struct{}

// sendEmailFailureMsg is the tea.Msg handled by Bubble Tea when the email has
// failed to send.
type sendEmailFailureMsg error

// sendEmailCmd returns a tea.Cmd that sends the email.
func (m Model) sendEmailCmd() tea.Cmd {
	return func() tea.Msg {
		attachments := make([]string, len(m.Attachments.Items()))
		for i, a := range m.Attachments.Items() {
			attachments[i] = a.FilterValue()
		}
		err := sendEmail(strings.Split(m.To.Value(), TO_SEPARATOR), m.From.Value(), m.Subject.Value(), m.Body.Value(), attachments)
		if err != nil {
			return sendEmailFailureMsg(err)
		}
		return sendEmailSuccessMsg{}
	}
}

func sendEmail(to []string, from, subject, body string, paths []string) error {
	client := resend.NewClient(os.Getenv(RESEND_API_KEY))

	html := bytes.NewBufferString("")
	// If the conversion fails, we'll simply send the plain-text body.
	_ = goldmark.Convert([]byte(body), html)

	attachments := make([]resend.Attachment, len(paths))
	for i, a := range paths {
		f, err := os.ReadFile(a)
		if err != nil {
			continue
		}
		attachments[i] = resend.Attachment{
			Content:  string(f),
			Filename: filepath.Base(a),
		}
	}

	if len(attachments) == 0 {
		attachments = nil
	}

	request := &resend.SendEmailRequest{
		From:        from,
		To:          to,
		Subject:     subject,
		Html:        html.String(),
		Text:        body,
		Attachments: attachments,
	}

	_, err := client.Emails.Send(request)
	if err != nil {
		return err
	}

	return nil
}
