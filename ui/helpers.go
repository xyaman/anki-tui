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
	notes  []models.Note
	start  int
	end    int
	morphs bool
}

// Return also end
func FetchNotes(query string, start, end int, morphs bool) tea.Cmd {
	return func() tea.Msg {
		res, err := core.App.AnkiConnect.FetchNotesFromQuery(query, start, end)
		if err != nil {
			return core.Log(core.InfoLog{Text: err.Error(), Seconds: 3, Type: "error"})
		}

		for i := range res.Result {
			res.Result[i].GetFieldsValues(
				core.App.Config.SentenceFieldName,
				core.App.Config.MorphFieldName,
				core.App.Config.AudioFieldName,
				core.App.Config.ImageFieldName,
			)
		}

		return FetchNotesMsg{notes: res.Result, start: start, end: len(res.Result), morphs: morphs}
	}
}
