package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/xyaman/anki-tui/core"
	"github.com/xyaman/anki-tui/ui"
)

type tickMsg time.Time

func tick() tea.Msg {
	time.Sleep(time.Second)
	return tickMsg{}
}

type model struct {
	state     ui.SessionState
	MainPage  tea.Model
	QueryPage tea.Model

	logs []core.InfoLog
}

func initialModel() model {
	return model{
		state:     ui.MainPanel,
		MainPage:  ui.NewMainPage(),
		QueryPage: ui.NewQueryPage(),
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.MainPage.Init(), m.QueryPage.Init(), tick)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.KeyMsg:
		k := msg.String()
		if k == "ctrl+c" {
			return m, tea.Quit
		}
	case ui.SessionState:
		m.state = msg
		return m, nil
	case ui.FetchNotesMsg:
		var cmd tea.Cmd
		m.QueryPage, cmd = m.QueryPage.Update(msg)
		return m, cmd

	case tea.WindowSizeMsg:
		core.App.Height = msg.Height
		core.App.Width = msg.Width

		core.App.AvailableHeight = msg.Height - 5
		core.App.AvailableWidth = msg.Width - 2

		var cmds = make([]tea.Cmd, 2)
		m.MainPage, cmds[0] = m.MainPage.Update(msg)
		m.QueryPage, cmds[1] = m.QueryPage.Update(msg)
		return m, tea.Batch(cmds...)

	case tickMsg:
		newlogs := make([]core.InfoLog, 0)
		for i := range m.logs {
			m.logs[i].Seconds -= 1
			if m.logs[i].Seconds > 0 {
				newlogs = append(newlogs, m.logs[i])
			}
		}
		m.logs = newlogs

		return m, tick

	case core.InfoLog:
		m.logs = append(m.logs, msg)
		return m, nil

	case ui.ErrorMsg:
		m.logs = append(m.logs, core.InfoLog{Text: string(msg), Type: "error", Seconds: 3})
		return m, nil
	}

	var cmd tea.Cmd

	if m.state == ui.MainPanel {
		m.MainPage, cmd = m.MainPage.Update(msg)
	} else if m.state == ui.QueryPanel {
		m.QueryPage, cmd = m.QueryPage.Update(msg)
	}

	return m, cmd
}

func (m model) View() string {

	var b strings.Builder
	if m.state == ui.MainPanel {
		b.WriteString(m.MainPage.View())
	} else if m.state == ui.QueryPanel {
		b.WriteString(m.QueryPage.View())
	}

	var logs []string
	// Add all logs
	for _, log := range m.logs {
		if log.Type == "error" {
			txt := ui.ErrorNotificationStyle.Render("Error: " + log.Text)
			logs = append(logs, txt)
		} else {
			txt := ui.InfoNotificationStyle.Render("Info: " + log.Text)
			logs = append(logs, txt)
		}
	}

	return lipgloss.JoinVertical(
		lipgloss.Top,
		lipgloss.NewStyle().Height(core.App.Height-3).Render(b.String()),
		lipgloss.JoinVertical(lipgloss.Top, logs...),
	)
}

func main() {

	// load config
	// Save config as a global variable somewhere and access from everywhere
	core.App = core.NewAnkiTui()

	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
