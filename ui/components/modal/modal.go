package modal

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	Text       string
	OkText     string
	CancelText string
	IsVisible  bool

	// kind of id
	Kind     string
	Cursor   int
	selected int
}

func New(kind string, cursor int) Model {
	model := Model{
		Text:       "",
		OkText:     "Ok",
		CancelText: "Cancel",
		Kind:       kind,
		Cursor:     cursor,
	}
	return model
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "right", "h", "l":
			m.selected = 1 - m.selected
		case "enter":
			m.IsVisible = false
			if m.selected == 0 {
				return m, SendOkMsg(m.Kind, m.Cursor)
			} else {
				return m, SendCancelMsg(m.Kind)
			}
		}
	}

	return m, nil
}

func (m Model) View() string {
	if !m.IsVisible {
		return ""
	}

	blurredStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Padding(0, 3).MarginTop(1)
	focusedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Padding(0, 3).MarginTop(1).MarginRight(2).Underline(true)

	okText := blurredStyle.Render(m.OkText)
	cancelText := blurredStyle.Render(m.CancelText)

	if m.selected == 0 {
		okText = focusedStyle.Render(fmt.Sprintf("[ %s ]", m.OkText))
	} else {
		cancelText = focusedStyle.Render(fmt.Sprintf("[ %s ]", m.CancelText))
	}

	text := lipgloss.NewStyle().Width(50).Align(lipgloss.Center).Render(m.Text)
	buttons := lipgloss.JoinHorizontal(lipgloss.Top, okText, cancelText)
	return lipgloss.JoinVertical(lipgloss.Center, text, buttons)
}
