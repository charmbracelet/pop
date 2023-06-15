package main

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/resendlabs/resend-go"
)

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
			at, ok := a.(attachment)
			if !ok {
				continue
			}
			attachments[i] = string(at)
		}
		err := sendEmail(m.From.Value(), m.To.Value(), m.Subject.Value(), m.Body.Value(), attachments)
		if err != nil {
			return sendEmailFailureMsg(err)
		}
		return sendEmailSuccessMsg{}
	}
}

func sendEmail(from, to, subject, body string, attachments []string) error {
	client := resend.NewClient(os.Getenv(RESEND_API_KEY))

	request := &resend.SendEmailRequest{
		From:    from,
		To:      []string{to},
		Subject: subject,
		Html:    body,
	}

	_, err := client.Emails.Send(request)
	if err != nil {
		return err
	}

	return nil
}
