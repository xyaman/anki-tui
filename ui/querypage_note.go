package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mistakenelf/teacup/image"
	"github.com/xyaman/anki-tui/core"
	"github.com/xyaman/anki-tui/models"
)

type NotePage struct {
	note      *models.Note
	image     image.Model
	imagepath string
}

func NewNotePage() NotePage {

	image := image.New(true, true, lipgloss.AdaptiveColor{Light: "#000000", Dark: "#ffffff"})

	return NotePage{
		note:  nil,
		image: image,
	}
}

func (m NotePage) Init() tea.Cmd {
	return nil
}

func (m NotePage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

  // este tamaño se muestra
  m.image.SetSize(core.App.Width, core.App.Height)

	var cmd tea.Cmd
	m.image, cmd = m.image.Update(msg)
	return m, cmd
}

func (m NotePage) View() string {
	// Show image
	// Show sentence
	// Show morphs
	// Show tags

	return m.image.View()
}
