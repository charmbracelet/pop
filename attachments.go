package main

import (
	"io"
	"path/filepath"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type attachment string

func (a attachment) FilterValue() string {
	return string(a)
}

type attachmentDelegate struct {
	focused bool
}

func (d attachmentDelegate) Height() int {
	return 1
}

func (d attachmentDelegate) Spacing() int {
	return 0
}

func (d attachmentDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	path := filepath.Base(item.(attachment).FilterValue())
	style := textStyle
	if m.Index() == index && d.focused {
		style = activeTextStyle
	}

	if m.Index() == index {
		_, _ = w.Write([]byte(style.Render("• " + path)))
	} else {
		_, _ = w.Write([]byte(style.Render("  " + path)))
	}
}

func (d attachmentDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}
