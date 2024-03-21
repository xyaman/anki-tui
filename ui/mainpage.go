package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/xyaman/anki-tui/core"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

type MainPage struct {
	list list.Model
}

func NewMainPage() MainPage {
	items := []list.Item{
		item{title: "Query", desc: "Show cards based on a query"},
		item{title: "Morphs", desc: "Learn unknown morphs"},
	}

	m := MainPage{list: list.New(items, list.NewDefaultDelegate(), 0, 0)}
	m.list.Title = "Anki TUI"
	m.list.DisableQuitKeybindings()

	return m
}

func (m MainPage) Init() tea.Cmd {
	return nil
}

func (m MainPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "enter" {
			switch m.list.Index() {
			case 0:
				return m, GoToPanel(QueryPanel)
			}
		}
	}

	h, v := docStyle.GetFrameSize()
	m.list.SetSize(core.App.AvailableWidth-h, core.App.AvailableHeight-v)

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m MainPage) View() string {
	return docStyle.Render(m.list.View())
}
