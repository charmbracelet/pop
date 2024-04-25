package main

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/resendlabs/resend-go"
	mail "github.com/xhit/go-simple-mail/v2"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	renderer "github.com/yuin/goldmark/renderer/html"
)

// ToSeparator is the separator used to split the To, Cc, and Bcc fields.
const ToSeparator = ","

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
		var err error
		to := strings.Split(m.To.Value(), ToSeparator)
		cc := strings.Split(m.Cc.Value(), ToSeparator)
		bcc := strings.Split(m.Bcc.Value(), ToSeparator)
		switch m.DeliveryMethod {
		case SMTP:
			err = sendSMTPEmail(to, cc, bcc, m.From.Value(), m.Subject.Value(), m.Body.Value(), attachments)
		case Resend:
			err = sendResendEmail(to, cc, bcc, m.From.Value(), m.Subject.Value(), m.Body.Value(), attachments)
		default:
			err = errors.New("[ERROR]: unknown delivery method")
		}
		if err != nil {
			// Store the body if the email fails to send
			path, storeErr := storeEmailBody(m.Body.Value())
			if storeErr == nil {
				err = fmt.Errorf("%w\nStored a copy of the email body at %s", err, path)
			}
			return sendEmailFailureMsg(err)
		}
		return sendEmailSuccessMsg{}
	}
}

const gmailSuffix = "@gmail.com"
const gmailSMTPHost = "smtp.gmail.com"
const gmailSMTPPort = 587

func sendSMTPEmail(to, cc, bcc []string, from, subject, body string, attachments []string) error {
	server := mail.NewSMTPClient()

	var err error
	server.Username = smtpUsername
	server.Password = smtpPassword
	server.Host = smtpHost
	server.Port = smtpPort

	// Set defaults for gmail.
	if strings.HasSuffix(server.Username, gmailSuffix) {
		if server.Port == 0 {
			server.Port = gmailSMTPPort
		}
		if server.Host == "" {
			server.Host = gmailSMTPHost
		}
	}

	switch strings.ToLower(smtpEncryption) {
	case "ssl":
		server.Encryption = mail.EncryptionSSLTLS
	case "none":
		server.Encryption = mail.EncryptionNone
	default:
		server.Encryption = mail.EncryptionSTARTTLS
	}

	server.KeepAlive = false
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second
	server.TLSConfig = &tls.Config{
		InsecureSkipVerify: smtpInsecureSkipVerify, //nolint:gosec
		ServerName:         server.Host,
	}

	smtpClient, err := server.Connect()

	if err != nil {
		return err
	}

	email := mail.NewMSG()
	email.SetFrom(from).
		AddTo(to...).
		AddCc(cc...).
		AddBcc(bcc...).
		SetSubject(subject)

	html := bytes.NewBufferString("")
	convertErr := goldmark.Convert([]byte(body), html)

	if convertErr != nil {
		email.SetBody(mail.TextPlain, body)
	} else {
		email.SetBody(mail.TextHTML, html.String())
	}

	for _, a := range attachments {
		email.Attach(&mail.File{
			FilePath: a,
			Name:     filepath.Base(a),
		})
	}

	return email.Send(smtpClient)
}

func sendResendEmail(to, _, _ []string, from, subject, body string, attachments []string) error {
	client := resend.NewClient(resendAPIKey)

	html := bytes.NewBufferString("")
	// If the conversion fails, we'll simply send the plain-text body.
	if unsafe {
		markdown := goldmark.New(
			goldmark.WithRendererOptions(
				renderer.WithUnsafe(),
			),
			goldmark.WithExtensions(
				extension.Strikethrough,
				extension.Table,
				extension.Linkify,
			),
		)
		_ = markdown.Convert([]byte(body), html)
	} else {
		_ = goldmark.Convert([]byte(body), html)
	}

	request := &resend.SendEmailRequest{
		From:        from,
		To:          to,
		Subject:     subject,
		Html:        html.String(),
		Text:        body,
		Attachments: makeAttachments(attachments),
	}

	_, err := client.Emails.Send(request)
	if err != nil {
		return err
	}

	return nil
}

func makeAttachments(paths []string) []resend.Attachment {
	if len(paths) == 0 {
		return nil
	}

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

	return attachments
}

// Store the email body in /tmp. Returns a path to the file created.
func storeEmailBody(body string) (string, error) {
	dir := "/tmp/pop"

	err := os.MkdirAll(dir, 0700)
	if err != nil {
		return "", fmt.Errorf("creating %s: %w", dir, err)
	}

	curr := time.Now()
	fileName := fmt.Sprintf("%d-%d-%d-%d", curr.Year(), curr.Month(), curr.Day(), secondsSinceDayStart(curr))
	fullPath := filepath.Join(dir, fileName)

	file, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("creating %s: %w", fullPath, err)
	}
	defer file.Close()

	_, err = file.Write([]byte(body))
	if err != nil {
		return "", fmt.Errorf("writing to %s: %w", fullPath, err)
	}

	return fullPath, nil
}

// Returns the seconds passed since the start of t.
// The start of t is determined as y:m:d:0:0:0:0
func secondsSinceDayStart(t time.Time) int {
	y, m, d := t.Date()
	start := time.Date(y, m, d, 0, 0, 0, 0, t.Location())
	elapsed := time.Since(start)
	return int(elapsed.Seconds())
}
