package main

import (
	"strings"
	"time"

	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
)

type tickMsg time.Time

type model struct {
	percent          float64
	progress         progress.Model
	downloadFullPath string
}

func (m model) Init() tea.Cmd {
	return tickCmd()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		return m, nil

	case tea.WindowSizeMsg:
		m.progress.SetWidth(msg.Width - padding*2 - 4)
		return m, nil

	case tickMsg:
		m.percent += 0.25
		if m.percent > 1.0 {
			m.percent = 1.0
			time.Sleep(1 * time.Second)
			return m, tea.Quit
		}
		return m, tickCmd()

	default:
		return m, nil
	}
}

const (
	padding = 2
)

func (m model) View() tea.View {
	pad := strings.Repeat(" ", padding)

	output := "\n" +
		pad + "📥 " + m.downloadFullPath + "\n\n" +
		pad + m.progress.ViewAs(m.percent) + "\n\n" +
		pad + "Press Ctrl+C to quit"

	if m.percent == 1.0 {
		output = "\n" +
			pad + "📥 " + m.downloadFullPath + "\n\n" +
			pad + "☑️ Transfer Completed"
	}

	v := tea.NewView(output)

	v.AltScreen = true

	return v
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
