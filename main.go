package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/xyaman/anki-tui/core"
	"github.com/xyaman/anki-tui/ui"
)

func main() {

	core.App = core.NewAnkiTui()

	p := tea.NewProgram(ui.NewProgram(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
