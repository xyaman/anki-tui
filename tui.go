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

const (
	MINNING_PAGE = "minning"
	CONFIG_PAGE  = "config"
)

type App struct {
	tviewApp    *tview.Application
	AnkiConnect *AnkiConnect

	// Config struct + form fields
	Config *Config

	MinningView *tview.Flex
	Pages       *tview.Pages

	LeftPanel tview.Primitive
	TopBox    *tview.TextView
	MidBox    *tview.Flex
	BottomBox *tview.TextView

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

	SearchQuery     string
	SearchMorphView *tview.Flex
}

func NewApp(config *Config, ankiconnect *AnkiConnect) (*App, error) {

	// Init Speaker
	sr := beep.SampleRate(48000)
	speaker.Init(sr, sr.N(time.Second/2))

	collectionPath, err := ankiconnect.GetMediaDirPath()
	if err != nil {
		return nil, err
	}

	pages := tview.NewPages()

	tviewApp := tview.NewApplication().SetRoot(pages, true).SetFocus(pages)

	app := &App{
		tviewApp:    tviewApp,
		AnkiConnect: ankiconnect,

		Config: config,

		NotesId:        nil,
		NoteIndex:      0,
		collectionPath: collectionPath,

		Pages: pages,
		// Views are initialized in setupViews
	}

	app.setupViews()

	// Keys
	app.MinningView.SetInputCapture(app.minningViewInput)
	app.moveCard(reload)

	return app, nil
}

func (a *App) setupViews() {

	// MainView
	topbox := tview.NewTextView().SetScrollable(false).SetDynamicColors(true)
	topbox.SetBorder(true).SetTitle("anki-tui").SetTitleAlign(tview.AlignLeft)

	bottombox := tview.NewTextView().SetScrollable(false)
	cardimageview := tview.NewImage()
	leftpanel := tview.NewFlex().SetDirection(0).AddItem(topbox, 6, 1, false).AddItem(cardimageview, 0, 5, false).AddItem(bottombox, 3, 1, false)
	mainView := tview.NewFlex().SetDirection(1).AddItem(leftpanel, 0, 4, false)

	a.MinningView = mainView
	a.LeftPanel = leftpanel
	a.TopBox = topbox
	a.BottomBox = bottombox
	a.CardImageView = cardimageview

	// Config Form
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
				a.TopBox.SetText(fmt.Sprintf("Error saving config: %v", err))
			}

			a.moveCard(reload)
			a.Pages.SwitchToPage(MINNING_PAGE)
		}).
		AddButton("Cancel", func() {
			a.Pages.SwitchToPage(MINNING_PAGE)
		})

	a.Pages.AddPage(MINNING_PAGE, mainView, true, true)
	a.Pages.AddPage(CONFIG_PAGE, configForm, true, false)
}

func (a *App) minningViewInput(event *tcell.EventKey) *tcell.EventKey {
	if a.modalIsActive {
		return event
	}

	switch event.Key() {
	case tcell.KeyRight:
		a.moveCard(nextCard)
	case tcell.KeyLeft:
		a.moveCard(prevCard)
	}

	switch event.Rune() {
	case 'a':
		// copy morph to clipboard
		clipboard.Write(clipboard.FmtText, []byte(a.CurrentNote.Fields[a.Config.MorphFieldName].(map[string]interface{})["value"].(string)))
	case 's':
		lastAddedCard, err := a.AnkiConnect.GetLastAddedCard()
		if err != nil {
			a.UpdateTopText(fmt.Sprintf("Can't get last added card: %v", err), errorMsg)
		}

		a.modalIsActive = true

		confimationModal := tview.NewModal()
		a.MinningView.AddItem(confimationModal, 0, 0, false)
		confimationModal = confimationModal.SetText(fmt.Sprintf("Audio and Picture will be added to Note ID: %d", lastAddedCard.NoteID)).
			AddButtons([]string{"Continue", "Cancel"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				if buttonLabel == "Continue" {
					a.AddAudioAndPictureToLastCard()
					a.UpdateTopText("Audio and picture added succesfully!", infoMsg)
				}

				a.MinningView = a.MinningView.RemoveItem(confimationModal)
				a.modalIsActive = false
				a.tviewApp.SetFocus(a.MinningView)
			})

		a.tviewApp.SetFocus(confimationModal)
	case 'd':

		confirmationModal := tview.NewModal()
		a.modalIsActive = true

		a.MinningView.AddItem(confirmationModal, 0, 0, false)
		confirmationModal = confirmationModal.SetText("Are you sure you want to remove this note?").
			AddButtons([]string{"Yes", "No"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				if buttonLabel == "Yes" {
					err := a.AnkiConnect.DeleteNotes([]int{a.CurrentNote.NoteID})
					if err != nil {
						a.UpdateTopText(fmt.Sprintf("Error: %v", err), errorMsg)
					} else {
						a.NotesId = append(a.NotesId[:a.NoteIndex], a.NotesId[a.NoteIndex+1:]...)
						a.moveCard(reset)
						a.UpdateTopText("Note removed succesfully!", infoMsg)
					}
				}

				a.MinningView.RemoveItem(confirmationModal)
				a.modalIsActive = false
				a.tviewApp.SetFocus(a.MinningView)

			})

		a.tviewApp.SetFocus(confirmationModal)
	case 'k':
		a.SetCardAsKnown()
	case 'o':
		lastAddedCard, err := a.AnkiConnect.GetLastAddedCard()
		if err != nil {
			a.UpdateTopText(fmt.Sprintf("Can't get last added card: %v", err), errorMsg)
		}
		a.AnkiConnect.guiBrowse(fmt.Sprintf("nid:%d", lastAddedCard.NoteID))
	case 'c':
		// app.showConfigForm()
		// Send backspace event
		// app.tviewApp.QueueEvent(tcell.NewEventKey(tcell.KeyBackspace, 0, tcell.ModNone))
		a.Pages.SwitchToPage(CONFIG_PAGE)
	case 'r':
		a.playAudio()
	case '?':
		a.showKeysBindings()
	case '/':
		a.showSearchBar()
	}
	return event
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
		// Get Notes
		notes, err := a.AnkiConnect.FindNotes(fmt.Sprintf("%s %s", a.Config.Query, a.SearchQuery))
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
	fmt.Fprintf(a.TopBox, "Query: %s %s\n", a.Config.Query, a.SearchQuery)
	fmt.Fprintf(a.TopBox, "Total Notes: %d/%d\n", a.NoteIndex+1, len(a.NotesId))

	if kind == infoMsg {
		fmt.Fprintf(a.TopBox, "[green]%s", msg)
	} else {
		fmt.Fprintf(a.TopBox, "[red]%s", msg)
		return
	}
}

func (a *App) SetCardImage(path string) error {

	reader, err := os.Open(path)
	if err != nil {
		return err
	}
	defer reader.Close()
	m, _, err := image.Decode(reader)
	if err != nil {
		return err
	}
	a.CardImageView.SetImage(m)

	return nil
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

// Show keys bindings modal
func (a *App) showKeysBindings() {
	// Hide/remove all other views
	a.MinningView.RemoveItem(a.LeftPanel)

	keysModal := tview.NewModal()
	keysModal.SetText("Left Arrow: Previous Card\nRight Arrow: Next Card\nk: Mark as known\nd: Delete note\nr: (Re)play audio\ns: Add audio and image to last card created\no: Open last added card in anki\nc: Open config\na: copy morph to clipboard").
		AddButtons([]string{"Close"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			a.MinningView.RemoveItem(keysModal)
			a.tviewApp.SetFocus(a.MinningView)

			// add back all other views
			a.MinningView.AddItem(a.LeftPanel, 0, 4, false)
		})

	// align to left

	// Add modal to main view
	a.MinningView.AddItem(keysModal, 0, 10, true)
	a.tviewApp.SetFocus(keysModal)
}

// Show search bar (it adds value to query and reloads cards)
func (a *App) showSearchBar() {
	// Hide/remove all other views
	a.MinningView.RemoveItem(a.LeftPanel)

	searchBar := tview.NewInputField()
	searchBar.SetLabel("Search: ").SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			a.SearchQuery = searchBar.GetText()
			a.modalIsActive = false
			a.MinningView.RemoveItem(searchBar)
			a.moveCard(reload)

			// add back all other views
			a.MinningView.AddItem(a.LeftPanel, 0, 4, false)
			a.tviewApp.SetFocus(a.MinningView)
		}
	})

	searchBar.SetText("")
	a.tviewApp.QueueEvent(tcell.NewEventKey(tcell.KeyBackspace, 0, tcell.ModNone))

	// Add form to main view
	a.MinningView.AddItem(searchBar, 0, 10, true)
	a.tviewApp.SetFocus(searchBar)
	a.modalIsActive = true
}
