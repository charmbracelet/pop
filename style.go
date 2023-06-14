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
)
