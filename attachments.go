// Package main implements Pop, a tool for sending emails from your terminal.
package main

import (
	"io"
	"path/filepath"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
)

type attachment string

func (a attachment) FilterValue() string {
	return string(a)
}

type attachmentDelegate struct {
	focused bool
	styles  composerStyles
}

func (d attachmentDelegate) Height() int {
	return 1
}

func (d attachmentDelegate) Spacing() int {
	return 0
}

func (d attachmentDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	path := filepath.Base(item.(attachment).FilterValue())
	style := d.styles.text
	if m.Index() == index && d.focused {
		style = d.styles.activeText
	}

	if m.Index() == index {
		_, _ = w.Write([]byte(style.Render("• " + path)))
	} else {
		_, _ = w.Write([]byte(style.Render("  " + path)))
	}
}

func (d attachmentDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd {
	return nil
}
