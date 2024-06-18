package cardviewer

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/xyaman/anki-tui/core"
	"github.com/xyaman/anki-tui/models"
	"github.com/xyaman/anki-tui/ui/components/image"
)

type Model struct {
	Image image.Model
	help  help.Model
	keys  HelpKeyMap

	Note      *models.Note
	imagepath string

	// Pitch
	// TODO: Make it private
	PitchMode     bool
	PitchCursor   int
	PitchDrops    []int
	PitchSentence string
}

// New creates a new image model
func New() Model {
	return Model{
		Image: image.New(),
		help:  help.New(),
		Note:  nil,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {

	var cmd tea.Cmd
	image, cmd := m.Image.Update(msg)
	m.Image = image

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "i":
			m.PitchMode = !m.PitchMode
			if m.PitchSentence == "" {
				sentence := m.Note.GetSentence()
				m.PitchSentence = core.ParseJpSentence(sentence)
			}
		case "l":
			if m.PitchMode {
				if m.PitchCursor < len(m.PitchSentence)-3 {
					m.PitchCursor += 3
					if m.PitchCursor < len(m.PitchSentence)-1 && m.PitchSentence[m.PitchCursor:m.PitchCursor+3] == "　" {
						m.PitchCursor += 3
					}
				}
			}
		case "h":
			if m.PitchMode {
				if m.PitchCursor > 3 && m.PitchSentence[m.PitchCursor-3:m.PitchCursor] == "　" {
					m.PitchCursor -= 3
				}
				m.PitchCursor -= 3
				if m.PitchCursor < 0 {
					m.PitchCursor = 0
				}
			}
		case "j", "k":
			if m.PitchMode {
				return m, nil
			}
		case "a":
			if m.PitchMode {
				m.PitchDrops = append(m.PitchDrops, m.PitchCursor)
			}
		case "u":
			if m.PitchMode {
				if len(m.PitchDrops) > 0 {
					m.PitchDrops = m.PitchDrops[:len(m.PitchDrops)-1]
				}
			}
		case "g":
			// See in anki gui
			core.App.AnkiConnect.GuiBrowse(fmt.Sprintf("nid:%d", m.Note.NoteID))
		}
	}

	return m, cmd
}

func (m Model) View() string {
	sentence := m.Note.GetSentence()
	morphs := m.Note.GetMorphs()
	if morphs == "" {
		morphs = "-"
	}

	var newSentence string
	if m.PitchMode {
		newSentence += "pitch: "

		// Add pitch drops
		for i, c := range m.PitchSentence {

			// Set color to see the cursor
			if i == m.PitchCursor {
				newSentence += lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(string(c))
			} else {
				newSentence += string(c)
			}

			for _, drop := range m.PitchDrops {
				if i == drop {
					newSentence += "↓"
				}
			}
		}

		newSentence += "\n"
	}

	width := 99 // It has to be multiple of 3 (ideally, because of japanese characters)
	height := core.App.AvailableHeight - lipgloss.Height(m.help.View(HelpKeys))

	// if sentence is longer than the width, edit sentence and add ... at the end
	if len(sentence) > width {
		sentence = sentence[:width-3] + "[...]"
	}

	// Center image, but align left image and text
	b := lipgloss.JoinVertical(
		lipgloss.Top,
		lipgloss.PlaceHorizontal(width, lipgloss.Center, m.Image.View()),
		lipgloss.JoinVertical(lipgloss.Top, "morphs: "+morphs, "sentence: "+sentence, newSentence, "tags: "+strings.Join(m.Note.Tags, ", ")),
	)

	var info string
	if m.PitchMode && m.PitchCursor-3 >= 0 {
		info = "Note\n"
	}

	renderImage := b
	main := lipgloss.Place(core.App.AvailableWidth, height, lipgloss.Center, lipgloss.Center, info+renderImage)

	return lipgloss.JoinVertical(lipgloss.Top, main, m.help.View(HelpKeys))
}

func (m *Model) SetNote(note *models.Note) {
	m.Note = note
	m.PitchMode = false
	m.PitchCursor = 0
	m.PitchDrops = []int{}
	m.PitchSentence = ""
}
