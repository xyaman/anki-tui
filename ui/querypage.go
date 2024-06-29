package ui

import (
	"errors"
	"fmt"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gopxl/beep"
	"github.com/gopxl/beep/generators"
	"github.com/gopxl/beep/speaker"

	"github.com/xyaman/anki-tui/core"
	"github.com/xyaman/anki-tui/models"
	"github.com/xyaman/anki-tui/ui/components/cardviewer"
	"github.com/xyaman/anki-tui/ui/components/modal"
)

const (
	mineModal   = "MineModal"
	deleteModal = "DeleteModal"
)

type QueryPage struct {
	table table.Model

	// Fetch cursor
	currentEnd int

	searchNotes     []models.Note
	morphNotes      []models.Note
	prevNotesCursor int

	help       help.Model
	notePage   cardviewer.Model
	configPage QueryPageConfig

	isConfig bool
	isNote   bool

	audioCtrl *beep.Ctrl
}

func NewQueryPage() QueryPage {

	t := table.New(
		table.WithFocused(true),
		table.WithColumns([]table.Column{
			{Title: "#", Width: 4},
			{Title: "Sentence", Width: 50},
			{Title: "Morphs", Width: 20},
			{Title: "Tags", Width: 50},
			{Title: "Source", Width: 50},
		}))

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	return QueryPage{
		table:       t,
		searchNotes: []models.Note{},
		help:        help.New(),
		morphNotes:  []models.Note{},
		notePage:    cardviewer.New(),
		configPage:  NewQueryPageConfig(),
		isConfig:    false,
		currentEnd:  100,
	}
}

// Init is called when the program starts. And it automically fetches notes
// based on the minning query.
// TODO: Don't fetch notes until is opened/needed.
func (m QueryPage) Init() tea.Cmd {
	if core.App.Config.MinningQuery == "" {
		return m.configPage.Init()
	}

	return tea.Batch(FetchNotes(core.App.Config.MinningQuery, 0, 100, false, false), m.configPage.Init())
}

func (m QueryPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	var k string

	// Handle configpage events
	if m.isConfig {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			k = msg.String()

			// If user press enter in Save button (last input)
			// save the config
			if k == "enter" && m.configPage.focused == len(m.configPage.inputs) {
				err := m.configPage.Save()
				if err != nil {
					panic(err)
				}
				m.isConfig = false
				m.searchNotes = []models.Note{}
				m.currentEnd = 100
				m.table.SetRows([]table.Row{})

				if core.App.Config.MinningQuery == "" {
					return m, core.Log(core.InfoLog{Type: "info", Text: "Minning query is empty.", Seconds: 3})
				}

				cmds := []tea.Cmd{FetchNotes(core.App.Config.MinningQuery, 0, 100, false, false)}

				if core.App.Config.SearchQuery == "" {
					cmds = append(cmds, core.Log(core.InfoLog{Type: "info", Text: "Search query is empty.", Seconds: 3}))
				}

				return m, tea.Batch(cmds...)

				// Exit config without saving if user presses esc
			} else if k == "esc" {
				m.isConfig = false
				return m, nil
			}

			newConfigPage, cmd := m.configPage.Update(msg)
			configPage, _ := newConfigPage.(QueryPageConfig)
			m.configPage = configPage

			return m, cmd
		}
	}

	// Handle notePage & cardview events
	switch msg := msg.(type) {
	case tea.KeyMsg:
		k = msg.String()

		switch k {
		// Exit NotePage, Config and MorphMode block
		// The order is:
		// 1. Exit notePage
		// 2. if not, exit morphMode and reset
		case "esc":
			if m.isNote {
				m.isNote = false
				return m, nil
			}

			isMorphMode := len(m.morphNotes) > 0
			if isMorphMode {
				m.morphNotes = []models.Note{}

				// Update table
				m.setNotesToTable(m.searchNotes)
				m.table.SetCursor(m.prevNotesCursor)

			}
			return m, nil

		// Enter Morphmode
		// "m" it will look for morphs in the local notes
		// "e" it will look for morphs in the external notes (BrigadaSOS, ImmersionKit, etc)
		case "m", "e":
			isMorphMode := len(m.morphNotes) > 0

			// If we are already in morph mode, we get the current note in the morphs array
			// If not, we get the current note in the searchNotes array
			morphs := m.searchNotes[m.table.Cursor()].GetMorphs()
			if isMorphMode {
				morphs = m.morphNotes[m.table.Cursor()].GetMorphs()
			}

			// Dont enter morph mode if there are no morphs in the selected note
			// if morphs == "" && !isMorphMode

			// If morphs is empty, we dont make any request
			if morphs == "" {
				return m, core.Log(core.InfoLog{Text: "The selected note has no morphs", Type: "Info", Seconds: 2})
			}

			// Local search
			if k == "m" {
				query := core.App.Config.SearchQuery + " " + strings.ReplaceAll(morphs, " ", " or ")
				return m, tea.Batch(
					FetchNotes(query, 0, 100, true, false),
					core.Log(core.InfoLog{Text: "Fetching morphs...", Type: "Info", Seconds: 1}),
				)

				// External search
			} else if k == "e" {
				return m, tea.Batch(
					FetchNotes(morphs, 0, 100, true, true),
					core.Log(core.InfoLog{Text: "[external] Fetching morphs...", Type: "Info", Seconds: 5}),
				)
			}
		case "c":
			m.isConfig = true
			return m, textinput.Blink

		case "p":
			note := m.searchNotes[m.table.Cursor()]
			if len(m.morphNotes) > 0 {
				note = m.morphNotes[m.table.Cursor()]
			}
			m.playAudio(&note)
			return m, nil

		case "o":
			m.showCardViewer()

			return m, nil

		case "ctrl+k":
			err := m.setCardAsKnown()
			if err != nil {
				return m, core.Log(core.InfoLog{Type: "error", Text: "Error when setting card as known", Seconds: 3})
			} else {
				return m, core.Log(core.InfoLog{Type: "info", Text: fmt.Sprintf("Card set as known (%s)", core.App.Config.KnownTag), Seconds: 2})
			}
		case "d":
			note := m.searchNotes[m.table.Cursor()]
			if len(m.morphNotes) > 0 {
				note = m.morphNotes[m.table.Cursor()]
			}

			// Show modal
			modal := modal.New(deleteModal, m.table.Cursor(), true)
			modal.Text = fmt.Sprintf("Delete note?\n\n%s", note.GetSentence())
			modal.OkText = "Confirm"
			modal.CancelText = "Cancel"
			return m, ShowModal(modal)

		case "ctrl+n":
			note := m.searchNotes[m.table.Cursor()]
			if len(m.morphNotes) > 0 {
				note = m.morphNotes[m.table.Cursor()]
			}

			// Show modal
			sentence := note.GetSentence()
			modal := modal.New(mineModal, m.table.Cursor(), true)
			modal.Text = fmt.Sprintf("Add image and sentence to last added card?\n\n%s", sentence)
			modal.OkText = "Yes"
			modal.CancelText = "No"
			return m, ShowModal(modal)
		case "y":
			sentence := m.searchNotes[m.table.Cursor()].GetSentence()
			if len(m.morphNotes) > 0 {
				sentence = m.morphNotes[m.table.Cursor()].GetSentence()
			}
			clipboard.WriteAll(sentence)

			// if user moves, update the note. Unless the note is in pitch mode
			// then pass the movements to the table too
		case "j", "k":
			if !m.notePage.PitchMode {
				var cmd tea.Cmd
				cmds := make([]tea.Cmd, 0)
				m.table, cmd = m.table.Update(msg)
				cmds = append(cmds, cmd)

				m.showCardViewer()
			}

			var cmd tea.Cmd
			m.notePage, cmd = m.notePage.Update(msg)
			return m, cmd
		}

	case tea.WindowSizeMsg:
		// We don't want to use the whole height
		// We have header
		m.table.SetHeight(core.App.AvailableHeight - 5 - lipgloss.Height(m.help.View(cardviewer.HelpKeys)))

		return m, nil

	case FetchNotesMsg:
		// Length is 0 when:
		// 1. First time fetching notes
		// 2. Config is updated
		isReload := len(m.searchNotes) == 0

		var notes []models.Note

		if msg.morphs {

			// This will run when we enter to morph mode
			if msg.morphs && len(m.morphNotes) == 0 {
				m.prevNotesCursor = m.table.Cursor()
			}

			m.morphNotes = msg.notes
			notes = m.morphNotes
			m.table.SetCursor(0)
		} else {
			m.searchNotes = append(m.searchNotes, msg.notes...)
			notes = m.searchNotes
		}

		if len(notes) == 0 {

			if msg.morphs {
				m.table.SetCursor(m.prevNotesCursor)
			}

			// Teest
			return m, core.Log(core.InfoLog{Type: "info", Text: "No notes were found with that query", Seconds: 4})
		}

		// Update table
		m.setNotesToTable(notes)

		if isReload {
			m.table.SetCursor(0)
		}

		// Update NotePage
		if m.isNote {
			m.showCardViewer()
		}

		return m, nil

	case modal.OkMsg:
		switch msg.ID {
		case deleteModal:
			err := core.App.AnkiConnect.DeleteNotes([]int{m.searchNotes[m.table.Cursor()].NoteID})
			if err != nil {
				return m, core.Log(core.InfoLog{Type: "error", Text: fmt.Sprintf("%s", err), Seconds: 3})
			} else {
				noteCursor := msg.Cursor
				m.currentEnd -= 1
				if len(m.morphNotes) > 0 {
					m.morphNotes = append(m.morphNotes[:noteCursor], m.morphNotes[noteCursor+1:]...)
					m.setNotesToTable(m.morphNotes)
				} else {
					m.searchNotes = append(m.searchNotes[:noteCursor], m.searchNotes[noteCursor+1:]...)
					m.setNotesToTable(m.searchNotes)
				}

				// if current cursor is the same as the deleted note
				// and the notepage is being used, update it
				if m.table.Cursor() == noteCursor && m.isNote {
					m.showCardViewer()
				}

				// return m, core.Log(core.InfoLog{Type: "info", Text: "Note deleted", Seconds: 2})
				return m, tea.Batch(
					core.Log(core.InfoLog{Type: "info", Text: "Note deleted", Seconds: 2}),
					HideModal(),
				)
			}

		case mineModal:
			note := m.searchNotes[msg.Cursor]
			if len(m.morphNotes) > 0 {
				note = m.morphNotes[msg.Cursor]
			}
			err := addImageAndSentenceToLastCard(&note)
			if err != nil {
				return m, core.Log(core.InfoLog{Type: "error", Text: fmt.Sprintf("%s", err), Seconds: 3})
			} else {
				return m, core.Log(core.InfoLog{Type: "info", Text: "Image and sentence added to last added card", Seconds: 2})
			}

		}
	case modal.CancelMsg:
		return m, HideModal()
	}

	// If the table is at the end, fetch more notes
	// This also works when NotePage is visible
	if m.table.Cursor() == m.currentEnd-1 {
		m.currentEnd += 100
		return m, FetchNotes(core.App.Config.MinningQuery, m.currentEnd, m.currentEnd+100, false, false)
	}

	// handle table
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m QueryPage) View() string {

	if m.isConfig {
		renderConfig := baseStyle.Render(m.configPage.View())
		return lipgloss.Place(core.App.AvailableWidth, core.App.AvailableHeight, lipgloss.Center, lipgloss.Center, renderConfig)
	}

	if m.isNote {
		return m.notePage.View()
	}

	topbarinfo := fmt.Sprintf("Query: %s \nTotal: %d\n\n", core.App.Config.MinningQuery, len(m.searchNotes))

	var b strings.Builder
	b.WriteString(topbarinfo)
	b.WriteString(m.table.View())

	main := lipgloss.PlaceHorizontal(core.App.AvailableWidth, lipgloss.Center, b.String())
	return lipgloss.JoinVertical(lipgloss.Top, main, m.help.View(cardviewer.HelpKeys))
}

func (qp *QueryPage) playAudio(note *models.Note) {

	// Stop previous audio
	if qp.audioCtrl != nil {
		speaker.Lock()
		qp.audioCtrl.Paused = true
		speaker.Unlock()
	}

	reader, streamer := note.GetAudio(core.App.CollectionPath)
	if streamer == nil {
		return
	}

	speaker.Lock()
	qp.audioCtrl = &beep.Ctrl{Streamer: streamer}
	speaker.Unlock()

	// Add a small silence delay
	silence := generators.Silence(12000)

	speaker.Play(beep.Seq(silence, qp.audioCtrl, beep.Callback(func() {
		streamer.Close()
		reader.Close()
	})))
}

func (qp *QueryPage) setNotesToTable(notes []models.Note) {
	rows := make([]table.Row, len(notes))
	for i, note := range notes {
		sentence := note.GetSentence()
		morphs := note.GetMorphs()
		rows[i] = table.Row{
			fmt.Sprintf("#%d", i+1),
			sentence,
			morphs,
			strings.Join(note.Tags, ", "),
			note.GetSource(),
		}
	}
	qp.table.SetRows(rows)
}

func (m *QueryPage) showCardViewer() {
	m.isNote = true

	note := m.searchNotes[m.table.Cursor()]
	if len(m.morphNotes) > 0 {
		note = m.morphNotes[m.table.Cursor()]
	}

	prevNote := 0
	if m.notePage.Note != nil {
		prevNote = m.notePage.Note.NoteID
	}
	m.notePage.SetNote(&note)
	m.notePage.Image.SetSize(50, 50)
	image := note.GetImage(core.App.CollectionPath)
	m.notePage.Image.SetImage(image)

	if core.App.Config.PlayAudioAutomatically && note.NoteID != prevNote {
		m.playAudio(&note)
	}
}

func (qp *QueryPage) setCardAsKnown() error {
	note := qp.searchNotes[qp.table.Cursor()]
	if len(qp.morphNotes) > 0 {
		note = qp.morphNotes[qp.table.Cursor()]
	}
	note.Tags = append(note.Tags, core.App.Config.KnownTag)
	return core.App.AnkiConnect.AddTags(note.NoteID, core.App.Config.KnownTag)
}

// Add image and sentence to last added card
func addImageAndSentenceToLastCard(note *models.Note) error {

	image := note.GetImageValue()
	audio := note.GetAudioValue()

	if note.GetSource() != "Anki" {
		var err error
		image, err = note.DownloadImage(core.App.CollectionPath)
		if err != nil {
			return err
		}
		audio, err = note.DownloadAudio(core.App.CollectionPath)
		if err != nil {
			return err
		}
	}

	// Get last added card
	lastAddedCard, err := core.App.AnkiConnect.GetLastAddedCard()
	if err != nil {
		return err
	}

	if audio == "" {
		return errors.New("No audio field found, check settings")
	} else if image == "" {
		return errors.New("No image field found, check settings")
	}

	err = core.App.AnkiConnect.UpdateNoteFields(lastAddedCard.NoteID, models.Fields{
		core.App.Config.MinningAudioFieldName: audio,
		core.App.Config.MinningImageFieldName: image,
	})

	// Add tcore.App. (except 1T, MT, 0T)
	for _, tag := range note.Tags {
		if tag != "1T" && tag != "MT" && tag != "0T" {
			err = core.App.AnkiConnect.AddTags(lastAddedCard.NoteID, tag)
			if err != nil {
				return err
			}
		}
	}

	return err
}
