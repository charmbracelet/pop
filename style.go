package main

import "github.com/charmbracelet/lipgloss"

const accentColor = lipgloss.Color("99")
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
	commentStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#757575"))

	inlineCodeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5F87")).Background(lipgloss.Color("#3A3A3A")).Padding(0, 1)
	linkStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("#00AF87")).Underline(true)
)
