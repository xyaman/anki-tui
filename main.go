package main

import (
	"fmt"

	"github.com/xyaman/anki-tui/core"
	"github.com/xyaman/anki-tui/ui"
)

func main() {

	core.App = core.NewAnkiTui()

  ui.ShowMainPage()
  
  if err := core.App.Tview.Run(); err != nil {
    fmt.Println("Error: ", err)
    return
  }
}
