package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/xyaman/anki-tui/core"
	"github.com/xyaman/anki-tui/models"
	"github.com/xyaman/anki-tui/ui/components/image"
)

type keyMap struct {
	NextNote  key.Binding
	PrevNote  key.Binding
	PlayAudio key.Binding
	SeeInAnki key.Binding
	Return    key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.NextNote, k.PrevNote, k.PlayAudio, k.SeeInAnki, k.Return}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		k.ShortHelp(),
	}
}

var keys = keyMap{
	NextNote: key.NewBinding(
		key.WithKeys("j"),
		key.WithHelp("j", "Next note"),
	),
	PrevNote: key.NewBinding(
		key.WithKeys("k"),
		key.WithHelp("k", "Previous note"),
	),
	PlayAudio: key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p", "Play audio"),
	),
	SeeInAnki: key.NewBinding(
		key.WithKeys("g"),
		key.WithHelp("g", "See in Anki"),
	),
	Return: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "Return to main menu"),
	),
}

type NotePage struct {
	image image.Model
	help  help.Model
	keys  keyMap

	note      *models.Note
	imagepath string

	// Pitch
	pitchMode     bool
	pitchCursor   int
	pitchDrops    []int
	pitchSentence string
}

func NewNotePage() NotePage {

	image := image.New()

	return NotePage{
		image: image,
		help:  help.New(),
		note:  nil,
	}
}

func (m NotePage) Init() tea.Cmd {
	return nil
}

func (m NotePage) Update(msg tea.Msg) (NotePage, tea.Cmd) {

	var cmd tea.Cmd
	image, cmd := m.image.Update(msg)
	m.image = image

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "i":
			m.pitchMode = !m.pitchMode
			if m.pitchSentence == "" {
				sentence, _, _ := getNoteFields(m.note)
				m.pitchSentence = core.ParseJpSentence(sentence)
			}
		case "l":
			if m.pitchMode {
				if m.pitchCursor < len(m.pitchSentence)-3 {
					m.pitchCursor += 3
					if m.pitchCursor < len(m.pitchSentence)-1 && m.pitchSentence[m.pitchCursor:m.pitchCursor+3] == "　" {
						m.pitchCursor += 3
					}
				}

			}
		case "h":
			if m.pitchMode {
				if m.pitchCursor > 3 && m.pitchSentence[m.pitchCursor-3:m.pitchCursor] == "　" {
					m.pitchCursor -= 3
				}
				m.pitchCursor -= 3
				if m.pitchCursor < 0 {
					m.pitchCursor = 0
				}
			}
		case "a":
			if m.pitchMode {
				m.pitchDrops = append(m.pitchDrops, m.pitchCursor)
			}
		case "u":
			if m.pitchMode {
				if len(m.pitchDrops) > 0 {
					m.pitchDrops = m.pitchDrops[:len(m.pitchDrops)-1]
				}
			}
		case "g":
			// See in anki gui
			core.App.AnkiConnect.GuiBrowse(fmt.Sprintf("nid:%d", m.note.NoteID))
		}
	}

	return m, cmd
}

func (m NotePage) View() string {

	sentence, morphs, _ := getNoteFields(m.note)
	if morphs == "" {
		morphs = "-"
	}

	var newSentence string
	if m.pitchMode {
		newSentence += "pitch: "

		// Add pitch drops
		for i, c := range m.pitchSentence {

			// Set color to see the cursor
			if i == m.pitchCursor {
				newSentence += lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(string(c))
			} else {
				newSentence += string(c)
			}

			for _, drop := range m.pitchDrops {
				if i == drop {
					newSentence += "↓"
				}
			}
		}

		newSentence += "\n"
	}

	width := 99 // It has to be multiple of 3 (ideally, because of japanese characters)
	height := core.App.AvailableHeight - lipgloss.Height(m.help.View(keys))

	// if sentence is longer than the width, edit sentence and add ... at the end
	if len(sentence) > width {
		sentence = sentence[:width-3] + "[...]"
	}

	// Center image, but align left image and text
	b := lipgloss.JoinVertical(
		lipgloss.Top,
		lipgloss.PlaceHorizontal(width, lipgloss.Center, m.image.View()),
		lipgloss.JoinVertical(lipgloss.Top, "morphs: "+morphs, "sentence: "+sentence, newSentence, "tags: "+strings.Join(m.note.Tags, ", ")),
	)

	var info string
	if m.pitchMode && m.pitchCursor-3 >= 0 {
		info = "Note\n"
	}

	renderImage := b
	main := lipgloss.Place(core.App.AvailableWidth, height, lipgloss.Center, lipgloss.Center, info+renderImage)

	return lipgloss.JoinVertical(lipgloss.Top, main, m.help.View(keys))
}

func (m *NotePage) SetNote(note *models.Note) {
	m.note = note
	m.pitchMode = false
	m.pitchCursor = 0
	m.pitchDrops = []int{}
	m.pitchSentence = ""
}
