package modal

import tea "github.com/charmbracelet/bubbletea"

type OkMsg struct {
	Kind   string
	Cursor int
}

type CancelMsg struct {
	Kind string
}

func SendOkMsg(kind string, cursor int) tea.Cmd {
	return func() tea.Msg {
		return OkMsg{Kind: kind, Cursor: cursor}
	}
}

func SendCancelMsg(kind string) tea.Cmd {
	return func() tea.Msg {
		return CancelMsg{Kind: kind}
	}
}
