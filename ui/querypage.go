package ui

import (
	"fmt"
	"image"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/gopxl/beep"
	"github.com/gopxl/beep/generators"
	"github.com/gopxl/beep/mp3"
	"github.com/gopxl/beep/speaker"
	"github.com/rivo/tview"
	"golang.design/x/clipboard"

	"github.com/ikawaha/kagome-dict/ipa"
	"github.com/ikawaha/kagome/v2/tokenizer"

	"github.com/xyaman/anki-tui/core"
	"github.com/xyaman/anki-tui/models"
)

type LoadEvent int

const (
	NextNote LoadEvent = iota
	PreviousNote
	FullReload
	Refresh
)

type QueryPage struct {
	// Maybe add the posibility to use table to show the results
	Table      *tview.Table
	TableMode  bool
	QueryLimit int

	// TODO: Change the type
	MinningView tview.Primitive
	TopBar      *tview.TextView
	CardImage   *tview.Image

	CardInfo       *tview.TextView
	SentenceCursor int
	PitchDrops     []int
	// This string is saved in case we want to copy to the clipboard
	PitchSentence string

	// Status info
	Query       string
	NotesId     []int
	NotesInfo   []models.Note
	NoteCursor  int
	CurrentNote *models.Note

	NoteSentence  string
	NoteMorphs    string
	NoteAudioPath string
	NoteImagePath string

	// Audio
	AudioCtrl *beep.Ctrl

	KeysBuffer []rune
}

func ShowQueryPage() {

	queryPage := newQueryPage()
	core.App.Tview.SetFocus(queryPage.MinningView)
	core.App.PageHolder.AddAndSwitchToPage(core.QueryPageID, queryPage.MinningView, true)
}

func newQueryPage() *QueryPage {

	topbar := tview.NewTextView().SetScrollable(false)

	cardimage := tview.NewImage()
	cardinfo := tview.NewTextView().SetScrollable(false).SetDynamicColors(true)

	table := tview.NewTable().
		SetSelectable(true, false).
		SetSelectedStyle(tcell.StyleDefault.Foreground(tcell.ColorGreen).Background(tcell.ColorBlack)).SetFixed(1, 1).
    SetSeparator('|').
    SetBordersColor(tcell.ColorGray)

	tablemode := true

  minningview := tview.NewFlex().SetDirection(0).
  AddItem(topbar, 3, 1, false).
  AddItem(tview.NewFlex().
    AddItem(nil, 0, 1, false).
    AddItem(table, 0, 3, true).
    AddItem(nil, 0, 1, false), 0, 1, true)

	queryPage := &QueryPage{
		Table:      table,
		TableMode:  tablemode,
		QueryLimit: 50,

		MinningView: minningview,
		TopBar:      topbar,
		CardImage:   cardimage,
		CardInfo:    cardinfo,
	}

	minningview.SetInputCapture(queryPage.setupInputCapture)
	// cardinfo.SetInputCapture(queryPage.setupCardInfoInput)
	// queryPage.load(FullReload)

	queryPage.fetchAllNotes()

	return queryPage
}

func (qp *QueryPage) fetchAllNotes() {
	// Fetch all ids
	notesResult, err := core.App.AnkiConnect.FindNotes(core.App.Config.MinningQuery)
	if err != nil {
		panic(err)
	}

	if len(notesResult.Result) == 0 {
		return
	}

	qp.NotesId = notesResult.Result[:qp.QueryLimit]
	fmt.Fprintf(qp.TopBar, "Total notes: %d", len(qp.NotesId))

	notes, err := core.App.AnkiConnect.NotesInfo(qp.NotesId)
	if err != nil {
		panic(err)
	}

	qp.NotesInfo = notes.Result

	// Add headers
	qp.Table.SetCell(0, 0, tview.NewTableCell("#").SetAlign(tview.AlignCenter))
	qp.Table.SetCell(0, 1, tview.NewTableCell("Sentence").SetAlign(tview.AlignCenter))
	qp.Table.SetCell(0, 2, tview.NewTableCell("Morphs").SetAlign(tview.AlignCenter))
	qp.Table.SetCell(0, 3, tview.NewTableCell("Tags").SetAlign(tview.AlignCenter))

	for i, note := range qp.NotesInfo {
		sentence, morphs := qp.getFields(&note)
		qp.Table.SetCell(i+1, 0, tview.NewTableCell(fmt.Sprintf("#%d", i+1)).SetAlign(tview.AlignCenter))
		qp.Table.SetCell(i+1, 2, tview.NewTableCell(morphs).SetAlign(tview.AlignCenter))
		qp.Table.SetCell(i+1, 1, tview.NewTableCell(sentence).SetAlign(tview.AlignCenter))
		qp.Table.SetCell(i+1, 3, tview.NewTableCell(strings.Join(note.Tags, ", ")).SetAlign(tview.AlignCenter))
	}

  qp.Table.Select(1, 0)
}

func (qp *QueryPage) setupInputCapture(event *tcell.EventKey) *tcell.EventKey {
  
	core.App.Tview.Sync()

	// implement vim-like go to number keybind
	if event.Rune() >= '0' && event.Rune() <= '9' {
		qp.KeysBuffer = append(qp.KeysBuffer, event.Rune())
    qp.TopBar.Clear()
    fmt.Fprintf(qp.TopBar, "KeysBuffer: %s", string(qp.KeysBuffer))
  
    strnumber := string(qp.KeysBuffer)
    number, err := strconv.Atoi(strnumber)
    if err == nil && number > 0 && number <= len(qp.NotesInfo) {
      qp.Table.Select(number, 0)
    }

		return event
	}
  qp.TopBar.Clear()
  fmt.Fprintf(qp.TopBar, "Query: %s", core.App.Config.MinningQuery)
	qp.KeysBuffer = qp.KeysBuffer[:0]

	switch event.Rune() {
	case 'r':
		// get selection
		row, _ := qp.Table.GetSelection()
    if row == 0 {
      return event
    }
		qp.CurrentNote = &qp.NotesInfo[row-1]
		qp.playAudio()
	}

	return event

	switch event.Rune() {
	case 'r':
		qp.playAudio()
	case 'p':
		qp.SentenceCursor = 0
		core.App.Tview.SetFocus(qp.CardInfo)
	// TODO: Change to other place.. refactor
	case 'c':
		ShowConfigPage()
	}

	switch event.Key() {
	case tcell.KeyCtrlK:
		qp.markCardAsKnown()
	}

	return event
}

func (qp *QueryPage) setupCardInfoInput(event *tcell.EventKey) *tcell.EventKey {

	core.App.Tview.Sync()
	return event

	switch event.Key() {
	case tcell.KeyEsc:
		core.App.Tview.SetFocus(qp.MinningView)
		qp.load(Refresh)
		return event
	}
	// TODO: If next character rune is 1 length then move the cursor to the next character
	// otherwise move 3 (japanese character for example)
	switch event.Rune() {
	case 'l':
		if qp.SentenceCursor+3 > len(qp.PitchSentence) {
			return event
		}
		qp.SentenceCursor += 3

	case 'h':
		if qp.SentenceCursor-3 < 0 {
			return event
		}
		qp.SentenceCursor -= 3
	case 'a':
		// if there is already a cursor, remove it
		// otherwise append to the array
		for _, pitchDrop := range qp.PitchDrops {
			if pitchDrop == qp.SentenceCursor {
				break
			}
		}
		qp.PitchDrops = append(qp.PitchDrops, qp.SentenceCursor)
	case 'u':
		// remove last pitch
		if len(qp.PitchDrops) > 0 {
			qp.PitchDrops = qp.PitchDrops[:len(qp.PitchDrops)-1]
		}
	case 'y':
		// Copy the sentence including the pitch drops
		sentence, _ := qp.getFields(qp.CurrentNote)
		output := fmt.Sprintf("%s\n %s", sentence, qp.PitchSentence)
		clipboard.Write(clipboard.FmtText, []byte(output))
	}
	qp.setupCardInfo()
	return event
}

func (qp *QueryPage) setupCardInfo() {
	// I want to be able to move the cursor to the next character
	// I also want to show a cursor in the sentence

	qp.CardInfo.Clear()

	sentence, morphs := qp.getFields(qp.CurrentNote)

	fmt.Fprintf(qp.CardInfo, "Morph: %s || Tags: %s\n\n", morphs, strings.Join(qp.CurrentNote.Tags, ", "))
	fmt.Fprintf(qp.CardInfo, "Sentence: %s\n", sentence)

	tagger, err := tokenizer.New(ipa.Dict(), tokenizer.OmitBosEos())
	if err != nil {
		panic(err)
	}

	seg := tagger.Tokenize(sentence)
	var rawSentence string
	for _, token := range seg {
		reading, hasReading := token.Reading()
		if hasReading {
			rawSentence += reading + "　"
		} else {
			rawSentence += fmt.Sprintf("%s　", token.Surface)
		}
	}

	newIndex := 0
	// Add "＼" under every pitch drop in the array
	for _, pitchDrop := range qp.PitchDrops {
		sign := "＼"
		rawSentence = rawSentence[:pitchDrop] + sign + rawSentence[pitchDrop:]
		newIndex += 3
	}

	qp.PitchSentence = rawSentence

	// Add background under the cursor index to show the cursor
	var outputsentence string
	for i, char := range rawSentence {
		if i == qp.SentenceCursor {
			outputsentence += fmt.Sprintf("[#ff0000]%s[white]", string(char))
		} else {
			outputsentence += string(char)
		}
	}

	fmt.Fprintf(qp.CardInfo, "Pitch Drop: %s\n", outputsentence)
}

func (qp *QueryPage) load(event LoadEvent) {

	if qp.TableMode {
		return
	}

	qp.SentenceCursor = 0

	switch event {
	case NextNote:
		if len(qp.NotesId) == 0 || qp.NoteCursor >= len(qp.NotesId) {
			return
		}
		qp.NoteCursor++
	case PreviousNote:
		if len(qp.NotesId) == 0 || qp.NoteCursor <= 0 {
			return
		}
		qp.NoteCursor--

	// In this case we want to fetch the notes again
	// When this event is called:
	// - The query has changed
	// - The page was just loaded
	case FullReload:
		qp.NoteCursor = 0

		notesResult, err := core.App.AnkiConnect.FindNotes(core.App.Config.MinningQuery)
		if err != nil {
			panic(err)
		}

		if len(notesResult.Result) == 0 {
			return
		}

		qp.NotesId = notesResult.Result
	}

	// TODO: When use table, fetch all notes

	notes, err := core.App.AnkiConnect.NotesInfo(qp.NotesId[qp.NoteCursor : qp.NoteCursor+1])
	if err != nil {
		panic(err)
	}

	qp.CurrentNote = &notes.Result[0]

	// Update ImageView
	imageFieldsName := strings.Split(core.App.Config.ImageFieldName, ",")
	for _, fieldName := range imageFieldsName {
		if imageField, ok := qp.CurrentNote.Fields[fieldName]; ok {
			imageFile := imageField.(map[string]interface{})["value"].(string)
			imageFile = imageFile[10 : len(imageFile)-2]
			reader, err := os.Open(filepath.Join(core.App.CollectionPath, imageFile))
			if err != nil {
				panic(err)
			}
			defer reader.Close()
			m, _, err := image.Decode(reader)
			if err != nil {
				panic(err)
			}
			qp.CardImage.SetImage(m)
		}
	}

	// Update CardInfo
	qp.CardInfo.Clear()

	sentence, morphs := qp.getFields(qp.CurrentNote)
	fmt.Fprintf(qp.CardInfo, "Morph: %s || Tags: %s\n\n", morphs, strings.Join(qp.CurrentNote.Tags, ", "))
	fmt.Fprintf(qp.CardInfo, "Sentence: %s\n", sentence)

	// Update TopBar
	qp.TopBar.Clear()
	fmt.Fprint(qp.TopBar, "Query: "+core.App.Config.MinningQuery)

	if core.App.Config.PlayAudioAutomatically {
		qp.playAudio()
	}
}

func (qp *QueryPage) getFields(note *models.Note) (string, string) {

	var sentence, morphs string

	sentenceFieldsName := strings.Split(core.App.Config.SentenceFieldName, ",")
	for _, fieldName := range sentenceFieldsName {
		if sentenceField, ok := note.Fields[fieldName]; ok {
			sentence = sentenceField.(map[string]interface{})["value"].(string)
			break
		}
	}

	morphFieldsName := strings.Split(core.App.Config.MorphFieldName, ",")
	for _, fieldName := range morphFieldsName {
		if morphField, ok := note.Fields[fieldName]; ok {
			morphs = morphField.(map[string]interface{})["value"].(string)
			break
		}
	}

	return sentence, morphs
}

func (qp *QueryPage) playAudio() {

	// Stop previous audio
	if qp.AudioCtrl != nil {
		speaker.Lock()
		qp.AudioCtrl.Paused = true
		speaker.Unlock()
	}

	audioFieldsName := strings.Split(core.App.Config.AudioFieldName, ",")
	for _, fieldName := range audioFieldsName {
		if audioField, ok := qp.CurrentNote.Fields[fieldName]; ok {
			audioFile := audioField.(map[string]interface{})["value"].(string)
			audioFile = audioFile[7 : len(audioFile)-1]
			reader, err := os.Open(filepath.Join(core.App.CollectionPath, audioFile))
			if err != nil {
				panic(err)
			}
			streamer, _, err := mp3.Decode(reader)
			if err != nil {
				panic(err)
			}

			speaker.Lock()
			qp.AudioCtrl = &beep.Ctrl{Streamer: streamer}
			speaker.Unlock()

			// Add a small silence delay
			silence := generators.Silence(12000)

			speaker.Play(beep.Seq(silence, qp.AudioCtrl, beep.Callback(func() {
				streamer.Close()
				reader.Close()
			})))

			break
		}
	}
}

func (qp *QueryPage) markCardAsKnown() error {

	currentNote := qp.CurrentNote

	err := core.App.AnkiConnect.AddTags(currentNote.NoteID, core.App.Config.KnownTag)
	if err != nil {
		panic(err)
	}

	return nil
}
