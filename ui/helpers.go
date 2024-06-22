package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/xyaman/anki-tui/core"
	"github.com/xyaman/anki-tui/models"
	"github.com/xyaman/anki-tui/ui/components/modal"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("56"))

type ShowModalMsg modal.Model
type HideModalMsg struct{}
type SessionStateMsg string

const (
	MainPanel  SessionStateMsg = "MainPanel"
	QueryPanel SessionStateMsg = "QueryPanel"
)

func ShowModal(modal modal.Model) tea.Cmd {
  return func() tea.Msg {
    return ShowModalMsg(modal)
  }
}

func HideModal() tea.Cmd {
  return func() tea.Msg {
    return HideModalMsg{}
  }
}

func GoToPanel(panel SessionStateMsg) tea.Cmd {
	return func() tea.Msg {
		return SessionStateMsg(panel)
	}
}

type FetchNotesMsg struct {
	notes  []models.Note
	start  int
	end    int
	morphs bool
}

// Return also end
func FetchNotes(query string, start, end int, morphs bool, external bool) tea.Cmd {
	return func() tea.Msg {
		if !external {
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

		// External sources
		results := []models.Note{}
		for _, source := range core.App.ExternalSources {
			res, err := source.FetchNotesFromQuery(query, start, end)
			if err != nil {
				return core.Log(core.InfoLog{Text: err.Error(), Seconds: 3, Type: "error"})
			}

			results = append(results, res...)
		}

		return FetchNotesMsg{notes: results, start: start, end: len(results), morphs: morphs}
	}
}
