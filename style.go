package main

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const accentColor = lipgloss.Color("99")
const yellowColor = lipgloss.Color("#ECFD66")
const whiteColor = lipgloss.Color("255")
const grayColor = lipgloss.Color("241")
const darkGrayColor = lipgloss.Color("236")
const lightGrayColor = lipgloss.Color("247")

var (
	activeTextStyle = lipgloss.NewStyle().Foreground(whiteColor)
	textStyle       = lipgloss.NewStyle().Foreground(lightGrayColor)

	activeLabelStyle = lipgloss.NewStyle().Foreground(accentColor)
	labelStyle       = lipgloss.NewStyle().Foreground(grayColor)

	placeholderStyle = lipgloss.NewStyle().Foreground(darkGrayColor)
	cursorStyle      = lipgloss.NewStyle().Foreground(whiteColor)

	paddedStyle = lipgloss.NewStyle().Padding(1)

	errorHeaderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#F1F1F1")).Background(lipgloss.Color("#FF5F87")).Bold(true).Padding(0, 1).SetString("ERROR")
	errorStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5F87"))
	commentStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#757575")).PaddingLeft(1)

	sendButtonActiveStyle   = lipgloss.NewStyle().Background(accentColor).Foreground(yellowColor).Padding(0, 2)
	sendButtonInactiveStyle = lipgloss.NewStyle().Background(darkGrayColor).Foreground(lightGrayColor).Padding(0, 2)
	sendButtonStyle         = lipgloss.NewStyle().Background(darkGrayColor).Foreground(grayColor).Padding(0, 2)

	inlineCodeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5F87")).Background(lipgloss.Color("#3A3A3A")).Padding(0, 1)
	linkStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("#00AF87")).Underline(true)
)

// emailSummary returns a summary of the email that was sent. It is used when
// the user has sent an email successfully.
func emailSummary(to []string, subject string) string {
	var s strings.Builder
	s.WriteString("\n  Email ")
	s.WriteString(activeTextStyle.Render("\"" + subject + "\""))
	s.WriteString(" sent to ")
	for i, t := range to {
		if i > 0 {
			s.WriteString(", ")
		}
		s.WriteString(linkStyle.Render(t))
	}
	s.WriteString("\n\n")

	return s.String()
}
