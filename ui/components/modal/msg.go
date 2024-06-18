package modal

import tea "github.com/charmbracelet/bubbletea"

type OkMsg struct {
	ID   string
	Cursor int
}

type CancelMsg struct {
	ID string
}

func SendOkMsg(kind string, cursor int) tea.Cmd {
	return func() tea.Msg {
		return OkMsg{ID: kind, Cursor: cursor}
	}
}

func SendCancelMsg(kind string) tea.Cmd {
	return func() tea.Msg {
		return CancelMsg{ID: kind}
	}
}
