package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/xyaman/anki-tui/core"
)

var (
	focusedStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorStyle         = focusedStyle.Copy()
	noStyle             = lipgloss.NewStyle()
	helpStyle           = blurredStyle.Copy()
	cursorModeHelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	focusedButton = focusedStyle.Copy().Render("[ Save ]")
	blurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Save"))
)

const (
	MinningQuery = iota
	SearchQuery
	MorphFieldName
	SentenceFieldName
	ImageFieldName
	AudioFieldName
	KnownTag
	MinningImageFieldName
	MinningAudioFieldName
	PlayAudioAutomatically
)

var labels = []string{
	"Minning Query           ",
	"Search Query            ",
	"Morph Field Name        ",
	"Sentence Field Name     ",
	"Image Field Name        ",
	"Audio Field Name        ",
	"Known Tag               ",
	"Minning Image Field Name",
	"Minning Audio Field Name",
	"Play Audio Automatically",
}

type QueryPageConfig struct {
	inputs  []textinput.Model
	focused int
}

func NewQueryPageConfig() QueryPageConfig {
	var inputs = make([]textinput.Model, 9)
	for i := 0; i < 9; i++ {
		inputs[i] = textinput.New()
		inputs[i].Prompt = labels[i] + ": "
	}

	inputs[0].TextStyle = focusedStyle
	inputs[0].PromptStyle = focusedStyle

	inputs[MinningQuery].SetValue(core.App.Config.MinningQuery)
	inputs[SearchQuery].SetValue(core.App.Config.SearchQuery)
	inputs[MorphFieldName].SetValue(core.App.Config.MorphFieldName)
	inputs[SentenceFieldName].SetValue(core.App.Config.SentenceFieldName)
	inputs[ImageFieldName].SetValue(core.App.Config.ImageFieldName)
	inputs[AudioFieldName].SetValue(core.App.Config.AudioFieldName)
	inputs[KnownTag].SetValue(core.App.Config.KnownTag)
	inputs[MinningImageFieldName].SetValue(core.App.Config.MinningImageFieldName)
	inputs[MinningAudioFieldName].SetValue(core.App.Config.MinningAudioFieldName)

	return QueryPageConfig{
		inputs: inputs,
	}
}

func (m QueryPageConfig) Init() tea.Cmd {
	return textinput.Blink
}

func (m *QueryPageConfig) Save() error {
  core.App.Config.MinningQuery = m.inputs[MinningQuery].Value()
  core.App.Config.SearchQuery = m.inputs[SearchQuery].Value()
  core.App.Config.MorphFieldName = m.inputs[MorphFieldName].Value()
  core.App.Config.SentenceFieldName = m.inputs[SentenceFieldName].Value()
  core.App.Config.ImageFieldName = m.inputs[ImageFieldName].Value()
  core.App.Config.AudioFieldName = m.inputs[AudioFieldName].Value()
  core.App.Config.KnownTag = m.inputs[KnownTag].Value()
  core.App.Config.MinningImageFieldName = m.inputs[MinningImageFieldName].Value()
  core.App.Config.MinningAudioFieldName = m.inputs[MinningAudioFieldName].Value()
  return core.App.Config.Save()
}

func (m QueryPageConfig) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds = make([]tea.Cmd, len(m.inputs))

	switch msg := msg.(type) {
	case tea.KeyMsg:

		switch msg.String() {
		case "tab", "ctrl+n", "enter":
      if msg.String() == "enter" && m.focused == len(m.inputs) {
        return m, nil
      }
			m.focused++
      // Because we have the button
			if m.focused > len(m.inputs) {
				m.focused = 0
			}

		case "shift+tab", "ctrl+p":
			m.focused--
      if m.focused < 0 {
        m.focused = len(m.inputs)
      }
		}
	}
	for i := 0; i < len(m.inputs); i++ {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
		m.inputs[i].Blur()
		m.inputs[i].TextStyle = noStyle
		m.inputs[i].PromptStyle = noStyle

    if i == m.focused {
      m.inputs[i].Focus()
      m.inputs[i].TextStyle = focusedStyle
      m.inputs[i].PromptStyle = focusedStyle
    }
	}

	return m, tea.Batch(cmds...)
}

func (m QueryPageConfig) View() string {
	var b strings.Builder
	for i := range m.inputs {
		b.WriteString(m.inputs[i].View())
		if i < len(m.inputs)-1 {
			b.WriteString("\n")
		}
	}

  button := &blurredButton
	if m.focused == len(m.inputs) {
		button = &focusedButton
	}
	fmt.Fprintf(&b, "\n\n%s\n\n", *button)

	return docStyle.Render(b.String())
}
