package modal

import tea "github.com/charmbracelet/bubbletea"

type OkMsg struct {
	ID   string
	Cursor int
}

type CancelMsg struct {
	ID string
}

func SendOkMsg(id string, cursor int) tea.Cmd {
	return func() tea.Msg {
		return OkMsg{ID: id, Cursor: cursor}
	}
}

func SendCancelMsg(id string) tea.Cmd {
	return func() tea.Msg {
		return CancelMsg{ID: id}
	}
}
