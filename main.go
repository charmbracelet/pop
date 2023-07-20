package main

import (
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	mcobra "github.com/muesli/mango-cobra"
	"github.com/muesli/roff"
	"github.com/resendlabs/resend-go"
	"github.com/spf13/cobra"
)

const RESEND_API_KEY = "RESEND_API_KEY"
const UNSAFE_HTML = "UNSAFE_HTML"
const POP_FROM = "POP_FROM"
const POP_SIGNATURE = "POP_SIGNATURE"

var (
	from        string
	to          []string
	subject     string
	body        string
	attachments []string
	preview     bool
	unsafe      bool
	signature   string
)

var rootCmd = &cobra.Command{
	Use:   "pop",
	Short: "Send emails from your terminal",
	Long:  `Pop is a tool for sending emails from your terminal.`,
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

var (
	// Version stores the build version of VHS at the time of package through
	// -ldflags.
	//
	// go build -ldflags "-s -w -X=main.Version=$(VERSION)"
	Version string

	// CommitSHA stores the git commit SHA at the time of package through -ldflags.
	CommitSHA string
)

var CompletionCmd = &cobra.Command{
	Use:                   "completion [bash|zsh|fish|powershell]",
	Short:                 "Generate completion script",
	Long:                  `To load completions`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return rootCmd.GenBashCompletion(os.Stdout)
		case "zsh":
			return rootCmd.GenZshCompletion(os.Stdout)
		case "fish":
			return rootCmd.GenFishCompletion(os.Stdout, true)
		case "powershell":
			return rootCmd.GenPowerShellCompletion(os.Stdout)
		}
		return nil
	},
}

var ManCmd = &cobra.Command{
	Use:    "man",
	Short:  "Generate man page",
	Long:   `To generate the man page`,
	Args:   cobra.NoArgs,
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
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
	rootCmd.AddCommand(CompletionCmd, ManCmd)

	rootCmd.Flags().StringSliceVar(&to, "bcc", []string{}, "BCC recipients")
	rootCmd.Flags().StringSliceVar(&to, "cc", []string{}, "CC recipients")
	rootCmd.Flags().StringSliceVarP(&attachments, "attach", "a", []string{}, "Email's attachments")
	rootCmd.Flags().StringSliceVarP(&to, "to", "t", []string{}, "Recipients")
	rootCmd.Flags().StringVarP(&body, "body", "b", "", "Email's contents")
	envFrom := os.Getenv(POP_FROM)
	rootCmd.Flags().StringVarP(&from, "from", "f", envFrom, "Email's sender "+commentStyle.Render("($"+POP_FROM+")"))
	rootCmd.Flags().StringVarP(&subject, "subject", "s", "", "Email's subject")
	rootCmd.Flags().BoolVarP(&preview, "preview", "p", false, "Whether to preview the email before sending")
	unsafe := os.Getenv(UNSAFE_HTML) == "true"
	rootCmd.Flags().BoolVarP(&unsafe, "unsafe", "u", false, "Whether to allow unsafe HTML in the email body, also enable some extra markdown features (Experimental)")
	envSignature := os.Getenv("POP_SIGNATURE")
	rootCmd.Flags().StringVarP(&signature, "signature", "x", envSignature, "Signature to display at the end of the email. "+commentStyle.Render("($"+POP_SIGNATURE+")"))

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
