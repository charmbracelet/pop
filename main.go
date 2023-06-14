package main

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

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
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
