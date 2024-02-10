package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/gopxl/beep"
	"github.com/gopxl/beep/mp3"
	"github.com/gopxl/beep/speaker"
	"github.com/rivo/tview"
	"golang.design/x/clipboard"
)

const (
	nextCard = 1
	prevCard = -1
	reset    = 0
	reload   = 2
)

const (
	infoMsg  = 1
	errorMsg = 2
)

type App struct {
	tviewApp    *tview.Application
	AnkiConnect *AnkiConnect

	// Config struct + form fields
	Config *Config

	BaseView *tview.Flex

	LeftPanel tview.Primitive
	TopBox    *tview.TextView
	MidBox    *tview.Flex
	BottomBox *tview.TextView

	RightPanel *tview.TextView

	CardImageView *tview.Image
	modalIsActive bool

	TotalWords     int
	DeckName       string
	collectionPath string

	NotesId     []int
	NoteIndex   int
	CurrentNote *Note

	// Fields tracking
	CurrentImageValue string
	CurrentAudioValue string
}

func NewApp(config *Config, ankiconnect *AnkiConnect) (*App, error) {

	// Init Speaker
	sr := beep.SampleRate(48000)
	speaker.Init(sr, sr.N(time.Second/2))

	collectionPath, err := ankiconnect.GetMediaDirPath()
	if err != nil {
		return nil, err
	}

	// Interface

	topbox := tview.NewTextView().SetScrollable(false).SetDynamicColors(true)
	topbox.SetBorder(true).SetTitle("Anki-morph").SetTitleAlign(tview.AlignLeft)

	bottombox := tview.NewTextView().SetScrollable(false)

	cardimageview := tview.NewImage()

	leftpanel := tview.NewFlex().SetDirection(0).AddItem(topbox, 6, 1, false).AddItem(cardimageview, 0, 5, false).AddItem(bottombox, 3, 1, false)

	rightpanel := tview.NewTextView().SetScrollable(false)

	rightpanel.SetBorder(true).SetTitle("Keys")

	mainView := tview.NewFlex().SetDirection(1).AddItem(leftpanel, 0, 4, false).AddItem(rightpanel, 0, 1, false)
	tviewApp := tview.NewApplication().SetRoot(mainView, true)

	app := &App{
		tviewApp:    tviewApp,
		AnkiConnect: ankiconnect,

		Config: config,

		BaseView: mainView,

		LeftPanel:  leftpanel,
		RightPanel: rightpanel,

		TopBox:         topbox,
		BottomBox:      bottombox,
		CardImageView:  cardimageview,
		NotesId:        nil,
		NoteIndex:      0,
		collectionPath: collectionPath,
	}

	app.writeInRightPanel()

	// Keys
	mainView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if app.modalIsActive {
			return event
		}

		switch event.Key() {
		case tcell.KeyRight:
			app.moveCard(nextCard)
		case tcell.KeyLeft:
			app.moveCard(prevCard)
		}

		switch event.Rune() {
		case 's':
			lastAddedCard, err := app.AnkiConnect.GetLastAddedCard()
			if err != nil {
				app.UpdateTopText(fmt.Sprintf("Can't get last added card: %v", err), errorMsg)
			}

			app.modalIsActive = true

			confimationModal := tview.NewModal()
			mainView.AddItem(confimationModal, 0, 0, false)
			confimationModal = confimationModal.SetText(fmt.Sprintf("Audio and Picture will be added to Note ID: %d", lastAddedCard.NoteID)).
				AddButtons([]string{"Continue", "Cancel"}).
				SetDoneFunc(func(buttonIndex int, buttonLabel string) {
					if buttonLabel == "Continue" {
						app.AddAudioAndPictureToLastCard()
						app.UpdateTopText("Audio and picture added succesfully!", infoMsg)
					}

					mainView = mainView.RemoveItem(confimationModal)
					app.modalIsActive = false
					app.tviewApp.SetFocus(mainView)
				})

			app.tviewApp.SetFocus(confimationModal)
		case 'd':

			confirmationModal := tview.NewModal()
			app.modalIsActive = true

			app.BaseView.AddItem(confirmationModal, 0, 0, false)
			confirmationModal = confirmationModal.SetText("Are you sure you want to remove this note?").
				AddButtons([]string{"Yes", "No"}).
				SetDoneFunc(func(buttonIndex int, buttonLabel string) {
					if buttonLabel == "Yes" {
						err := app.AnkiConnect.DeleteNotes([]int{app.CurrentNote.NoteID})
						if err != nil {
							app.UpdateTopText(fmt.Sprintf("Error: %v", err), errorMsg)
						} else {
							app.NotesId = append(app.NotesId[:app.NoteIndex], app.NotesId[app.NoteIndex+1:]...)
							app.moveCard(reset)
							app.UpdateTopText("Note removed succesfully!", infoMsg)
						}
					}

					app.BaseView.RemoveItem(confirmationModal)
					app.modalIsActive = false
					app.tviewApp.SetFocus(app.BaseView)

				})

			app.tviewApp.SetFocus(confirmationModal)

		case 'k':
			app.SetCardAsKnown()
		case 'o':
			lastAddedCard, err := app.AnkiConnect.GetLastAddedCard()
			if err != nil {
				app.UpdateTopText(fmt.Sprintf("Can't get last added card: %v", err), errorMsg)
			}
			app.AnkiConnect.guiBrowse(fmt.Sprintf("nid:%d", lastAddedCard.NoteID))
		case 'c':
			app.showConfigForm()
			// Send backspace event
			app.tviewApp.QueueEvent(tcell.NewEventKey(tcell.KeyBackspace, 0, tcell.ModNone))
		case 'r':
			app.playAudio()
		}
		return event
	})

	// Start config form
	app.showConfigForm()

	return app, nil
}

func (a *App) writeInRightPanel() {
	a.RightPanel.Clear()
	fmt.Fprintf(a.RightPanel, "Left Arrow: Previous Card\n\n")
	fmt.Fprintf(a.RightPanel, "Right Arrow: Next Card\n\n")
	fmt.Fprintf(a.RightPanel, "k: Mark as known (tag: %s)\n\n", excludeTag)
	fmt.Fprintf(a.RightPanel, "d: Delete note\n\n")
	fmt.Fprintf(a.RightPanel, "r: (Re)play audio\n\n")
	fmt.Fprintf(a.RightPanel, "s: Add audio and image to last card created\n\n")
	fmt.Fprintf(a.RightPanel, "o: Open last added card in anki\n\n")
	fmt.Fprintf(a.RightPanel, "c: Open config\n\n")
}

func (a *App) showConfigForm() error {

	// Hide/remove all other views
	a.BaseView.RemoveItem(a.LeftPanel)
	a.BaseView.RemoveItem(a.RightPanel)

	configForm := tview.NewForm()

	configForm.AddInputField("Query", a.Config.Query, 0, nil, nil).
		AddInputField("Morph Field Name", a.Config.MorphFieldName, 0, nil, nil).
		AddInputField("Sentence Field Name", a.Config.SentenceFieldName, 0, nil, nil).
		AddInputField("Audio Field Name", a.Config.AudioFieldName, 0, nil, nil).
		AddInputField("Image Field Name", a.Config.ImageFieldName, 0, nil, nil).
		AddInputField("Known Tag", a.Config.KnownTag, 0, nil, nil).
		AddInputField("Minning Audio Field Name", a.Config.MinningAudioFieldName, 0, nil, nil).
		AddInputField("Minning Image Field Name", a.Config.MinningImageFieldName, 0, nil, nil).
		AddCheckbox("Play Audio Automatically", a.Config.PlayAudioAutomatically, nil).
		AddButton("Save", func() {
			defer a.tviewApp.SetFocus(a.BaseView)

			a.Config.Query = configForm.GetFormItemByLabel("Query").(*tview.InputField).GetText()
			a.Config.MorphFieldName = configForm.GetFormItemByLabel("Morph Field Name").(*tview.InputField).GetText()
			a.Config.SentenceFieldName = configForm.GetFormItemByLabel("Sentence Field Name").(*tview.InputField).GetText()
			a.Config.AudioFieldName = configForm.GetFormItemByLabel("Audio Field Name").(*tview.InputField).GetText()
			a.Config.ImageFieldName = configForm.GetFormItemByLabel("Image Field Name").(*tview.InputField).GetText()
			a.Config.KnownTag = configForm.GetFormItemByLabel("Known Tag").(*tview.InputField).GetText()
			a.Config.MinningAudioFieldName = configForm.GetFormItemByLabel("Minning Audio Field Name").(*tview.InputField).GetText()
			a.Config.MinningImageFieldName = configForm.GetFormItemByLabel("Minning Image Field Name").(*tview.InputField).GetText()
			a.Config.PlayAudioAutomatically = configForm.GetFormItemByLabel("Play Audio Automatically").(*tview.Checkbox).IsChecked()

			err := a.Config.save()
			if err != nil {
				a.UpdateTopText(fmt.Sprintf("Error saving config: %v", err), errorMsg)
			}

			a.modalIsActive = false
			a.BaseView.RemoveItem(configForm)
			a.moveCard(reload)

			// add back all other views
			a.BaseView.AddItem(a.LeftPanel, 0, 4, false)
			a.BaseView.AddItem(a.RightPanel, 0, 1, false)

		}).
		AddButton("Cancel", func() {
			a.modalIsActive = false
			a.BaseView.RemoveItem(configForm)
			a.tviewApp.SetFocus(a.BaseView)

			// add back all other views
			a.BaseView.AddItem(a.LeftPanel, 0, 4, false)
			a.BaseView.AddItem(a.RightPanel, 0, 1, false)
		})

	// Add form to main view
	a.BaseView.AddItem(configForm, 0, 10, true)
	a.tviewApp.SetFocus(configForm)
	a.modalIsActive = true

	return nil
}

func (a *App) moveCard(action int) {
	switch action {
	case nextCard:
		if a.NoteIndex == len(a.NotesId)-1 {
			return
		}
		a.NoteIndex++
	case prevCard:
		if a.NoteIndex == 0 {
			return
		}
		a.NoteIndex--
	case reset:
		if len(a.NotesId) == 0 {
			return
		}
	case reload:
		a.writeInRightPanel()

		// Get Notes
		notes, err := a.AnkiConnect.FindNotes(a.Config.Query)
		if err != nil {
			a.UpdateTopText(fmt.Sprintf("Error: %v", err), errorMsg)
		}

		if len(notes.Result) == 0 {
			a.UpdateTopText("No cards found! Try changing the query (c)", errorMsg)
			return
		}

		a.NotesId = notes.Result
		a.NoteIndex = 0
	}

	currentNotes, err := a.AnkiConnect.NotesInfo(a.NotesId[a.NoteIndex : a.NoteIndex+1])
	if err != nil {
		fmt.Println("Error:", err, a)
		return
	}

	currentNote := currentNotes.Result[0]
	a.CurrentNote = &currentNote

	// Update TopBox with current note info
	a.UpdateTopText("Card loaded sucessfully!", infoMsg)

	// Update ImageView
	imageFieldsNames := strings.Split(a.Config.ImageFieldName, ",")
	for index, imageFieldName := range imageFieldsNames {
		pictureField, exists := currentNote.Fields[strings.Trim(imageFieldName, " ")]
		if !exists {
			if index == len(imageFieldsNames)-1 {
				a.UpdateTopText(fmt.Sprintf("Error: image field ({%s}) not found!", a.Config.ImageFieldName), errorMsg)
			}
		} else {
			a.CurrentImageValue = pictureField.(map[string]interface{})["value"].(string)
			picture := pictureField.(map[string]interface{})["value"].(string)
			picture = picture[10 : len(picture)-2]
			a.SetCardImage(filepath.Join(a.collectionPath, picture))
			break
		}
	}

	// Update bottom box with sentence info
	a.BottomBox.Clear()

	morphFieldsNames := strings.Split(a.Config.MorphFieldName, ",")
	for index, morphFieldName := range morphFieldsNames {
		morphField, exists := currentNote.Fields[strings.Trim(morphFieldName, " ")]
		if !exists {
			if index == len(morphFieldsNames)-1 {
				a.UpdateTopText(fmt.Sprintf("Error: morph field ({%s}) not found!", a.Config.MorphFieldName), errorMsg)
			}
		} else {
			fmt.Fprintf(a.BottomBox, "Morph: %s\n", morphField.(map[string]interface{})["value"])
		}
	}

	sentenceFieldsNames := strings.Split(a.Config.SentenceFieldName, ",")
	for index, sentenceFieldName := range sentenceFieldsNames {
		sentenceField, exists := currentNote.Fields[strings.Trim(sentenceFieldName, " ")]
		if !exists {
			if index == len(sentenceFieldsNames)-1 {
				a.UpdateTopText(fmt.Sprintf("Error: sentence field ({%s}) not found!", a.Config.SentenceFieldName), errorMsg)
			}
		} else {
			fmt.Fprintf(a.BottomBox, "Sentence: %s\n", sentenceField.(map[string]interface{})["value"])
			clipboard.Write(clipboard.FmtText, []byte(sentenceField.(map[string]interface{})["value"].(string)))
		}
	}

	if a.Config.PlayAudioAutomatically {
		a.playAudio()
	}
}

func (a *App) playAudio() {

	audioFieldsNames := strings.Split(a.Config.AudioFieldName, ",")
	for index, audioFieldName := range audioFieldsNames {
		audioField, exists := a.CurrentNote.Fields[strings.Trim(audioFieldName, " ")]
		if !exists {
			if index == len(audioFieldsNames)-1 {
				a.UpdateTopText(fmt.Sprintf("Error: audio field ({%s}) not found!", a.Config.AudioFieldName), errorMsg)
			}
		} else {
			a.CurrentAudioValue = audioField.(map[string]interface{})["value"].(string)
			audio := audioField.(map[string]interface{})["value"].(string)
			audioFile, err := os.Open(filepath.Join(a.collectionPath, audio[7:len(audio)-1]))
			if err != nil {
				a.UpdateTopText(fmt.Sprintf("Error when playing the audio: %v", err), errorMsg)
			}
			streamer, _, err := mp3.Decode(audioFile)
			if err != nil {
				a.UpdateTopText(fmt.Sprintf("Error when playing the audio: %v", err), errorMsg)
			}

			speaker.Clear()
			speaker.Play(beep.Seq(streamer, beep.Callback(func() {
				defer streamer.Close()
			})))
		}
	}
}

func (a *App) UpdateTopText(msg string, kind int) {
	a.TopBox.Clear()
	fmt.Fprintf(a.TopBox, "Query: %s\n", a.Config.Query)
	fmt.Fprintf(a.TopBox, "Total Notes: %d/%d\n", a.NoteIndex+1, len(a.NotesId))

	if kind == infoMsg {
		fmt.Fprintf(a.TopBox, "[green]%s", msg)
	} else {
		fmt.Fprintf(a.TopBox, "[red]%s", msg)
		return
	}
}

func (a *App) SetCardImage(path string) {
	a.CardImageView, _ = SetImageToBox(path, a.CardImageView)
}

func (a *App) SetCardAsKnown() error {
	// Get current note
	currentNotes, err := a.AnkiConnect.NotesInfo(a.NotesId[a.NoteIndex : a.NoteIndex+1])
	if err != nil {
		return err
	}
	currentNote := currentNotes.Result[0]

	// Update note with known tag
	err = a.AnkiConnect.AddTags(currentNote.NoteID, a.Config.KnownTag)
	if err != nil {
		a.UpdateTopText(fmt.Sprintf("Error: %v", err), errorMsg)
		return err
	}
	a.UpdateTopText("Card marked as known!", infoMsg)

	return nil
}

func (a *App) AddAudioAndPictureToLastCard() error {
	// Get last added card
	lastAddedCard, err := a.AnkiConnect.GetLastAddedCard()
	if err != nil {
		return err
	}

	err = a.AnkiConnect.UpdateNoteFields(lastAddedCard.NoteID, Fields{
		a.Config.MinningAudioFieldName: a.CurrentAudioValue,
		a.Config.MinningImageFieldName: a.CurrentImageValue,
	})

	return err
}

func SetImageToBox(path string, c *tview.Image) (*tview.Image, error) {
	reader, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	m, _, err := image.Decode(reader)
	if err != nil {
		return nil, err
	}
	return c.SetImage(m), nil
}
