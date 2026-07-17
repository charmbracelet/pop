// Package main implements Pop, a tool for sending emails from your terminal.
package main

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/exp/charmtone"
)

var (
	accentColor    = charmtone.Charple
	yellowColor    = charmtone.Zest
	whiteColor     = charmtone.Soda
	grayColor      = charmtone.Steam
	darkGrayColor  = charmtone.Iron
	lightGrayColor = charmtone.Squid
)

var (
	activeTextStyle = lipgloss.NewStyle().Foreground(whiteColor)
	textStyle       = lipgloss.NewStyle().Foreground(lightGrayColor)

	activeLabelStyle              = lipgloss.NewStyle().Foreground(accentColor)
	labelStyle                    = lipgloss.NewStyle().Foreground(grayColor)
	attachmentsTitleActiveStyle   = lipgloss.NewStyle().Foreground(charmtone.Blush)
	attachmentsTitleInactiveStyle = lipgloss.NewStyle().Foreground(grayColor)

	placeholderStyle = lipgloss.NewStyle().Foreground(darkGrayColor)

	paddedStyle = lipgloss.NewStyle().Padding(1)

	errorHeaderStyle = lipgloss.NewStyle().
				Foreground(charmtone.Soda).
				Background(charmtone.Coral).
				Bold(true).
				Padding(0, 1).
				SetString("ERROR")
	errorStyle = lipgloss.NewStyle().
			Foreground(charmtone.Coral)

	// Headers in CLI output.
	noticeHeaderStyle = errorHeaderStyle.
				Background(charmtone.Charple)

	// Paragraphs in CLI output.
	paragraphStyle = lipgloss.NewStyle().
			Padding(0, 2).
			Width(80)

	commentStyle = lipgloss.NewStyle().
			Foreground(charmtone.Oyster).PaddingLeft(1)

	sendButtonActiveStyle = lipgloss.NewStyle().
				Background(accentColor).
				Foreground(yellowColor).
				Padding(0, 2)
	sendButtonInactiveStyle = lipgloss.NewStyle().
				Background(darkGrayColor).
				Foreground(lightGrayColor).
				Padding(0, 2)
	sendButtonStyle = lipgloss.NewStyle().
			Background(darkGrayColor).
			Foreground(grayColor).
			Padding(0, 2)

	inlineCodeStyle = lipgloss.NewStyle().
			Foreground(charmtone.Coral).
			Background(charmtone.Char).
			Padding(0, 1)
	linkStyle = lipgloss.NewStyle().
			Foreground(charmtone.Guac).
			Underline(true)
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
