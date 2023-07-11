package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/resendlabs/resend-go"
	"github.com/spf13/cobra"
)

const RESEND_API_KEY = "RESEND_API_KEY"
const POP_FROM = "POP_FROM"
const POP_SIGNATURE = "POP_SIGNATURE"

var (
	from        string
	to          []string
	subject     string
	body        string
	attachments []string
	preview     bool
	signature   string
)

var rootCmd = &cobra.Command{
	Use:   "email",
	Short: "email is a command line interface for sending emails.",
	Long:  `email is a command line interface for sending emails.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if os.Getenv(RESEND_API_KEY) == "" {
			fmt.Printf("\n  %s %s %s\n\n", errorHeaderStyle.String(), inlineCodeStyle.Render(RESEND_API_KEY), "environment variable is required.")
			fmt.Printf("  %s %s\n\n", commentStyle.Render("You can grab one at"), linkStyle.Render("https://resend.com/api-keys"))
			os.Exit(1)
		}

		if hasStdin() {
			b, err := io.ReadAll(os.Stdin)
			if err != nil {
				return err
			}
			body = string(b)
		}

		if signature != "" {
			body += "\n\n" + signature
		}

		if len(to) > 0 && from != "" && subject != "" && body != "" && !preview {
			err := sendEmail(to, from, subject, body, attachments)
			if err != nil {
				cmd.SilenceUsage = true
				cmd.SilenceErrors = true
				fmt.Println(errorStyle.Render(err.Error()))
				return err
			}
			fmt.Print(emailSummary(to, subject))
			return nil
		}

		p := tea.NewProgram(NewModel(resend.SendEmailRequest{
			From:        from,
			To:          to,
			Subject:     subject,
			Text:        body,
			Attachments: makeAttachments(attachments),
		}))
		m, err := p.Run()
		if err != nil {
			return err
		}
		mm := m.(Model)
		if !mm.abort {
			fmt.Print(emailSummary(strings.Split(mm.To.Value(), TO_SEPARATOR), mm.Subject.Value()))
		}
		return nil
	},
}

// hasStdin returns whether there is data in stdin.
func hasStdin() bool {
	stat, err := os.Stdin.Stat()
	return err == nil && (stat.Mode()&os.ModeCharDevice) == 0
}

func init() {
	rootCmd.Flags().StringSliceVar(&to, "bcc", []string{}, "BCC recipients")
	rootCmd.Flags().StringSliceVar(&to, "cc", []string{}, "CC recipients")
	rootCmd.Flags().StringSliceVarP(&attachments, "attach", "a", []string{}, "Email's attachments")
	rootCmd.Flags().StringSliceVarP(&to, "to", "t", []string{}, "Recipients")
	rootCmd.Flags().StringVarP(&body, "body", "b", "", "Email's contents")
	envFrom := os.Getenv(POP_FROM)
	rootCmd.Flags().StringVarP(&from, "from", "f", envFrom, "Email's sender "+commentStyle.Render("($"+POP_FROM+")"))
	rootCmd.Flags().StringVarP(&subject, "subject", "s", "", "Email's subject")
	rootCmd.Flags().BoolVarP(&preview, "preview", "p", false, "Whether to preview the email before sending")
	envSignature := os.Getenv("POP_SIGNATURE")
	rootCmd.Flags().StringVarP(&signature, "signature", "x", envSignature, "Signature to display at the end of the email. "+commentStyle.Render("($"+POP_SIGNATURE+")"))
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
