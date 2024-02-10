package main

import (
	"fmt"
)

// Settings
const (
	inputDeck        = "konosuba-s1"
	valueField       = "am-unknowns"
	sentenceField    = "Expression"
	excludeTag       = "am-known-manually"
	knownTag         = "am-known-manually"
	imageField       = "Screenshot"
	audioField       = "Audio_Sentence"
	outputImageField = "Picture"
	outputAudioField = "SentenceAudio"
	continueOption   = "y"
	exitOption       = "n"
	knownOption      = "k"
)

func main() {

	client := NewAnkiConnect("http://localhost:8765", 6)
	config, err := loadConfig()
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}

	app, err := NewApp(config, client)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}

	if err := app.tviewApp.Run(); err != nil {
		fmt.Println("Error: ", err)
		return
	}
}
