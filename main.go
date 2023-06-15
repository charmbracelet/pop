package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

const RESEND_API_KEY = "RESEND_API_KEY"

var rootCmd = &cobra.Command{
	Use:   "email",
	Short: "email is a command line interface for sending emails.",
	Long:  `email is a command line interface for sending emails.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		p := tea.NewProgram(NewModel())
		_, err := p.Run()
		if err != nil {
			return err
		}
		return nil
	},
}

func main() {
	key := os.Getenv(RESEND_API_KEY)
	if key == "" {
		fmt.Printf("\n  %s %s %s\n\n", errorHeaderStyle.String(), inlineCodeStyle.Render(RESEND_API_KEY), "environment variable is required.")
		fmt.Printf("  %s %s\n\n", commentStyle.Render("You can grab one at"), linkStyle.Render("https://resend.com"))
		os.Exit(1)
	}

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
