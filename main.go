package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	mcobra "github.com/muesli/mango-cobra"
	"github.com/muesli/roff"
	"github.com/resendlabs/resend-go"
	"github.com/spf13/cobra"
)

// PopUnsafeHTML is the environment variable that enables unsafe HTML in the
// email body.
const PopUnsafeHTML = "POP_UNSAFE_HTML"

// ResendAPIKey is the environment variable that enables Resend as a delivery
// method and uses it to send the email.
const ResendAPIKey = "RESEND_API_KEY" //nolint:gosec

// PopFrom is the environment variable that sets the default "from" address.
const PopFrom = "POP_FROM"

// PopSignature is the environment variable that sets the default signature.
const PopSignature = "POP_SIGNATURE"

// PopSMTPHost is the host for the SMTP server if the user is using the SMTP delivery method.
const PopSMTPHost = "POP_SMTP_HOST"

// PopSMTPPort is the port for the SMTP server if the user is using the SMTP delivery method.
const PopSMTPPort = "POP_SMTP_PORT"

// PopSMTPUsername is the username for the SMTP server if the user is using the SMTP delivery method.
const PopSMTPUsername = "POP_SMTP_USERNAME"

// PopSMTPPassword is the password for the SMTP server if the user is using the SMTP delivery method.
const PopSMTPPassword = "POP_SMTP_PASSWORD" //nolint:gosec

// PopSMTPEncryption is the encryption type for the SMTP server if the user is using the SMTP delivery method.
const PopSMTPEncryption = "POP_SMTP_ENCRYPTION" //nolint:gosec

// PopSMTPInsecureSkipVerify is whether or not to skip TLS verification for the
// SMTP server if the user is using the SMTP delivery method.
const PopSMTPInsecureSkipVerify = "POP_SMTP_INSECURE_SKIP_VERIFY"

var (
	from                   string
	to                     []string
	cc                     []string
	bcc                    []string
	subject                string
	body                   string
	attachments            []string
	preview                bool
	unsafe                 bool
	signature              string
	smtpHost               string
	smtpPort               int
	smtpUsername           string
	smtpPassword           string
	smtpEncryption         string
	smtpInsecureSkipVerify bool
	resendAPIKey           string
)

var rootCmd = &cobra.Command{
	Use:   "pop",
	Short: "Send emails from your terminal",
	Long:  `Pop is a tool for sending emails from your terminal.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var deliveryMethod DeliveryMethod
		switch {
		case resendAPIKey != "" && smtpUsername != "" && smtpPassword != "":
			deliveryMethod = Many
		case resendAPIKey != "":
			deliveryMethod = Resend
		case smtpUsername != "" && smtpPassword != "":
			deliveryMethod = SMTP
			from = smtpUsername
		}

		if deliveryMethod == Many {
			fmt.Printf("\n  Multiple sending methods are configured.\n")
			fmt.Printf(
				"\n  Please have either %s or %s environment variables set.\n\n",
				inlineCodeStyle.Render(ResendAPIKey),
				inlineCodeStyle.Render("POP_SMPT_*"),
			)
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true
			return errors.New("multiple environment variables set")
		}
		if deliveryMethod == None {
			fmt.Printf("\n  %s %s %s\n\n", errorHeaderStyle.String(), inlineCodeStyle.Render(ResendAPIKey), "environment variable is required.")
			fmt.Printf("  %s %s\n\n", commentStyle.Render("You can grab one at"), linkStyle.Render("https://resend.com/api-keys"))
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true
			return errors.New("missing required environment variable")
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
			var err error
			switch deliveryMethod {
			case SMTP:
				err = sendSMTPEmail(to, cc, bcc, from, subject, body, attachments)
			case Resend:
				err = sendResendEmail(to, cc, bcc, from, subject, body, attachments)
			default:
				err = fmt.Errorf("unknown delivery method")
			}
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
		}, deliveryMethod))

		m, err := p.Run()
		if err != nil {
			return err
		}
		mm := m.(Model)
		if !mm.abort {
			fmt.Print(emailSummary(strings.Split(mm.To.Value(), ToSeparator), mm.Subject.Value()))
		}
		return nil
	},
}

// hasStdin returns whether there is data in stdin.
func hasStdin() bool {
	stat, err := os.Stdin.Stat()
	return err == nil && (stat.Mode()&os.ModeCharDevice) == 0
}

var (
	// Version stores the build version of VHS at the time of package through
	// -ldflags.
	//
	// go build -ldflags "-s -w -X=main.Version=$(VERSION)"
	Version string

	// CommitSHA stores the git commit SHA at the time of package through -ldflags.
	CommitSHA string
)

// ManCmd is the cobra command for the manual.
var ManCmd = &cobra.Command{
	Use:    "man",
	Short:  "Generate man page",
	Long:   `To generate the man page`,
	Args:   cobra.NoArgs,
	Hidden: true,
	RunE: func(_ *cobra.Command, _ []string) error {
		page, err := mcobra.NewManPage(1, rootCmd) // .
		if err != nil {
			return err
		}

		page = page.WithSection("Copyright", "© 2023 Charmbracelet, Inc.\n"+"Released under MIT License.")
		fmt.Println(page.Build(roff.NewDocument()))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(ManCmd)

	rootCmd.Flags().StringSliceVar(&bcc, "bcc", []string{}, "BCC recipients")
	rootCmd.Flags().StringSliceVar(&cc, "cc", []string{}, "CC recipients")
	rootCmd.Flags().StringSliceVarP(&attachments, "attach", "a", []string{}, "Email's attachments")
	rootCmd.Flags().StringSliceVarP(&to, "to", "t", []string{}, "Recipients")
	rootCmd.Flags().StringVarP(&body, "body", "b", "", "Email's contents")
	envFrom := os.Getenv(PopFrom)
	rootCmd.Flags().StringVarP(&from, "from", "f", envFrom, "Email's sender"+commentStyle.Render("($"+PopFrom+")"))
	rootCmd.Flags().StringVarP(&subject, "subject", "s", "", "Email's subject")
	rootCmd.Flags().BoolVar(&preview, "preview", false, "Whether to preview the email before sending")
	envUnsafe := os.Getenv(PopUnsafeHTML) == "true"
	rootCmd.Flags().BoolVarP(&unsafe, "unsafe", "u", envUnsafe, "Whether to allow unsafe HTML in the email body, also enable some extra markdown features (Experimental)")
	envSignature := os.Getenv(PopSignature)
	rootCmd.Flags().StringVarP(&signature, "signature", "x", envSignature, "Signature to display at the end of the email."+commentStyle.Render("($"+PopSignature+")"))
	envSMTPHost := os.Getenv(PopSMTPHost)
	rootCmd.Flags().StringVarP(&smtpHost, "smtp.host", "H", envSMTPHost, "Host of the SMTP server"+commentStyle.Render("($"+PopSMTPHost+")"))
	envSMTPPort, _ := strconv.Atoi(os.Getenv(PopSMTPPort))
	if envSMTPPort == 0 {
		envSMTPPort = 587
	}
	rootCmd.Flags().IntVarP(&smtpPort, "smtp.port", "P", envSMTPPort, "Port of the SMTP server"+commentStyle.Render("($"+PopSMTPPort+")"))
	envSMTPUsername := os.Getenv(PopSMTPUsername)
	rootCmd.Flags().StringVarP(&smtpUsername, "smtp.username", "U", envSMTPUsername, "Username of the SMTP server"+commentStyle.Render("($"+PopSMTPUsername+")"))
	envSMTPPassword := os.Getenv(PopSMTPPassword)
	rootCmd.Flags().StringVarP(&smtpPassword, "smtp.password", "p", envSMTPPassword, "Password of the SMTP server"+commentStyle.Render("($"+PopSMTPPassword+")"))
	envSMTPEncryption := os.Getenv(PopSMTPEncryption)
	rootCmd.Flags().StringVarP(&smtpEncryption, "smtp.encryption", "e", envSMTPEncryption, "Encryption type of the SMTP server (starttls, ssl, or none)"+commentStyle.Render("($"+PopSMTPEncryption+")"))
	envInsecureSkipVerify := os.Getenv(PopSMTPInsecureSkipVerify) == "true"
	rootCmd.Flags().BoolVarP(&smtpInsecureSkipVerify, "smtp.insecure", "i", envInsecureSkipVerify, "Skip TLS verification with SMTP server"+commentStyle.Render("($"+PopSMTPInsecureSkipVerify+")"))
	envResendAPIKey := os.Getenv(ResendAPIKey)
	rootCmd.Flags().StringVarP(&resendAPIKey, "resend.key", "r", envResendAPIKey, "API key for the Resend.com"+commentStyle.Render("($"+ResendAPIKey+")"))

	rootCmd.CompletionOptions.HiddenDefaultCmd = true

	if len(CommitSHA) >= 7 { //nolint:gomnd
		vt := rootCmd.VersionTemplate()
		rootCmd.SetVersionTemplate(vt[:len(vt)-1] + " (" + CommitSHA[0:7] + ")\n")
	}
	if Version == "" {
		if info, ok := debug.ReadBuildInfo(); ok && info.Main.Sum != "" {
			Version = info.Main.Version
		} else {
			Version = "unknown (built from source)"
		}
	}
	rootCmd.Version = Version
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
