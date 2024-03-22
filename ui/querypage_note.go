package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/xyaman/anki-tui/core"
	"github.com/xyaman/anki-tui/image"
	"github.com/xyaman/anki-tui/models"
)

type NotePage struct {
	note      *models.Note
	image     image.Model
	imagepath string
}

func NewNotePage() NotePage {

	image := image.New()

	return NotePage{
		note:  nil,
		image: image,
	}
}

func (m NotePage) Init() tea.Cmd {
	return nil
}

func (m NotePage) Update(msg tea.Msg) (NotePage, tea.Cmd) {

	var cmd tea.Cmd
  image, cmd := m.image.Update(msg)
  m.image = image

	return m, cmd
}

func (m NotePage) View() string {
	// Show image
	// Show sentence
	// Show morphs
	// Show tags

	// Render a box using lipgloss
	noteStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("56"))

  sentence, _, _ := getNoteFields(m.note)
  b := lipgloss.JoinVertical(lipgloss.Center, m.image.View(), sentence)

	renderImage := noteStyle.Render(b)
	return lipgloss.Place(core.App.Width, core.App.Height, lipgloss.Center, lipgloss.Center, renderImage)
}
