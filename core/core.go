package core

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gopxl/beep"
	"github.com/gopxl/beep/speaker"
)

var (
	App *AnkiTui
)

type InfoLog struct {
	Text    string
	Seconds int
	Type    string
}

func Log(log InfoLog) tea.Cmd {
	return func() tea.Msg {
		return log
	}
}

// ---------------------

type AnkiTui struct {
	Config          *Config
	AnkiConnect     *AnkiConnect
	ExternalSources []ExternalSource
	CollectionPath  string

	Height          int
	Width           int
	AvailableHeight int
	AvailableWidth  int
}

func NewAnkiTui() *AnkiTui {

	sr := beep.SampleRate(48000)
	speaker.Init(sr, sr.N(time.Second/2))

	// TODO: Handle error
	config, _ := LoadConfig()

	ankiconnect := NewAnkiConnect("http://localhost:8765", 6)
	collectionPath, err := ankiconnect.GetMediaDirPath()
	if err != nil {
		panic("Error getting the collection path")
	}

	return &AnkiTui{
		Config:         config,
		AnkiConnect:    NewAnkiConnect("http://localhost:8765", 6),
		CollectionPath: collectionPath,
		ExternalSources: []ExternalSource{
			NewBrigadaSource("f34a3113-e164-4981-bd69-c58430fd64a1"),
		},
	}
}
