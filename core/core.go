package core

import (
	"fmt"
	"time"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/speaker"
	"github.com/rivo/tview"
)

var (
	App *AnkiTui
)

type AnkiTui struct {
	Config     *Config
	Tview      *tview.Application
	PageHolder *tview.Pages
	StatusBar  *tview.TextView

	AnkiConnect    *AnkiConnect
	CollectionPath string
}

func NewAnkiTui() *AnkiTui {

	sr := beep.SampleRate(48000)
	speaker.Init(sr, sr.N(time.Second/2))

  // tview.Styles.PrimitiveBackgroundColor = tcell.ColorDefault

	// TODO: Handle error
	config, _ := loadConfig()

	tviewApp := tview.NewApplication()
	pageHolder := tview.NewPages()
	statusBar := tview.NewTextView()

	fmt.Fprint(statusBar, "Welcome to AnkiTUI")

	grid := tview.NewGrid().SetRows(0, 1).SetColumns(0).
		AddItem(pageHolder, 0, 0, 1, 1, 0, 0, true).
		AddItem(statusBar, 1, 0, 1, 1, 0, 0, false)

	tviewApp.SetRoot(grid, true).SetFocus(pageHolder)

	ankiconnect := NewAnkiConnect("http://localhost:8765", 6)
	collectionPath, err := ankiconnect.GetMediaDirPath()
	if err != nil {
		panic("Error getting the collection path")
	}

	return &AnkiTui{
		Config:         config,
		Tview:          tviewApp,
		PageHolder:     pageHolder,
		AnkiConnect:    NewAnkiConnect("http://localhost:8765", 6),
		CollectionPath: collectionPath,
	}
}
