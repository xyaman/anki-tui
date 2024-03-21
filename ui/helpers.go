package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/xyaman/anki-tui/core"
	"github.com/xyaman/anki-tui/models"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("56"))

type SessionState string

const (
	MainPanel  SessionState = "MainPanel"
	QueryPanel SessionState = "QueryPanel"
)

func GoToPanel(panel SessionState) tea.Cmd {
	return func() tea.Msg {
		return SessionState(panel)
	}
}

type FetchNotesMsg struct {
	notes []models.Note
	start int
	end   int
}

// Return also end
func FetchNotes(query string, start, end int) tea.Cmd {
	return func() tea.Msg {
		result, err := core.App.AnkiConnect.FetchNotesFromQuery(query, start, end)
		if err != nil {
			panic(err)
		}
		return FetchNotesMsg{notes: result.Result, start: start, end: len(result.Result)}
	}
}
