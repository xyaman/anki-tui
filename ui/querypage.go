package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gopxl/beep"
	"github.com/gopxl/beep/generators"
	"github.com/gopxl/beep/mp3"
	"github.com/gopxl/beep/speaker"
	"github.com/xyaman/anki-tui/core"
	"github.com/xyaman/anki-tui/models"
)

type QueryPage struct {
	table table.Model

	// Fetch cursor
	currentEnd int

	notes           []models.Note
	morphNotes      []models.Note
	prevNotesCursor int

	notePage   NotePage
	configPage QueryPageConfig
	isConfig   bool
	isNote     bool

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
		table:      t,
		notes:      []models.Note{},
		morphNotes: []models.Note{},
		notePage:   NewNotePage(),
		configPage: NewQueryPageConfig(),
		isConfig:   false,
		currentEnd: 100,
	}
}

func (m QueryPage) Init() tea.Cmd {
	if core.App.Config.MinningQuery == "" {
		return m.configPage.Init()
	}

	return tea.Batch(FetchNotes(core.App.Config.MinningQuery, 0, 100, false), m.configPage.Init())
}

func (m QueryPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	var k string

	switch msg := msg.(type) {
	case tea.KeyMsg:
		k = msg.String()

		// Exit NotePage or MorphMode
		// The order is:
		// 1. Exit notePage
		// 2. if not, exit morphMode
		if k == "esc" && len(m.morphNotes) > 0 && !m.isConfig {
			if m.isNote {
				m.isNote = false
				return m, nil
			}
			m.morphNotes = []models.Note{}

			// Update table
			rows := make([]table.Row, len(m.notes))
			for i, note := range m.notes {
				sentence, morphs, _ := getNoteFields(&note)
				rows[i] = table.Row{
					fmt.Sprintf("#%d", i+1),
					sentence,
					morphs,
					strings.Join(note.Tags, ", "),
				}
			}
			m.table.SetRows(rows)
			m.table.SetCursor(m.prevNotesCursor)

			// update	image
			return m, nil
		}

		// Enter with m
		// Exit with esc
		if (k == "m") && !m.isConfig {

			isMorphMode := len(m.morphNotes) > 0

			// When entering morph mode, we need to get the morphs using "normal" notes
			// When we are already in morph mode, we need to get the morphs using the morph notes
			_, morphs, _ := getNoteFields(&m.notes[m.table.Cursor()])
			if isMorphMode {
				_, morphs, _ = getNoteFields(&m.morphNotes[m.table.Cursor()])
			}

			// Dont enter morph mode if there are no morphs in the selected note
			if morphs == "" && !isMorphMode {
				return m, core.Log(core.InfoLog{Text: "The selected note has no morphs", Type: "Info", Seconds: 2})
			}

			query := core.App.Config.SearchQuery + " " + strings.ReplaceAll(morphs, " ", " or ")
			return m, tea.Batch(
				FetchNotes(query, 0, 100, true),
				core.Log(core.InfoLog{Text: "Fetching morphs...", Type: "Info", Seconds: 1}),
			)

		} else if k == "esc" && (m.isConfig || m.isNote) {
			m.isConfig = false
			m.isNote = false
			return m, nil
		}

		if !m.isConfig {
			switch k {
			case "c":
				m.isConfig = true
				return m, textinput.Blink
			case "p":
				note := m.notes[m.table.Cursor()]
				if len(m.morphNotes) > 0 {
					note = m.morphNotes[m.table.Cursor()]
				}
				m.playAudio(&note)
			case "o":
				if m.isConfig {
					break
				}
				isMorphMode := len(m.morphNotes) > 0
				note := m.notes[m.table.Cursor()]
				if isMorphMode {
					note = m.morphNotes[m.table.Cursor()]
				}
				m.notePage.note = &note
				m.isNote = true

				_, _, image := getNoteFields(&note)
				m.notePage.imagepath = image
				m.notePage.SetNote(&note)

				// TODO: Change this to global
				m.notePage.image.SetSize(50, 50)
				m.notePage.image.SetImage(image)

				// play audio
				if core.App.Config.PlayAudioAutomatically {
					m.playAudio(&note)
				}

				return m, nil

			case "ctrl+k":
				err := m.setCardAsKnown()
				if err != nil {
					return m, tea.Batch(core.Log(core.InfoLog{Type: "error", Text: "Error when setting card as known", Seconds: 2}))
				} else {
					return m, tea.Batch(core.Log(core.InfoLog{Type: "info", Text: "Card set as known", Seconds: 2}))
				}
			}
		}

	case tea.WindowSizeMsg:
		// We don't want to use the whole height
		// We have header
		m.table.SetHeight(core.App.AvailableHeight - 5)

		return m, nil

	case FetchNotesMsg:
		// Length is 0 when:
		// 1. First time fetching notes
		// 2. Config is updated
		isReload := len(m.notes) == 0

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
			m.notes = append(m.notes, msg.notes...)
			notes = m.notes
		}

		if len(notes) == 0 {

			if msg.morphs {
				m.table.SetCursor(m.prevNotesCursor)
			}

			// Teest
			return m, core.Log(core.InfoLog{Type: "info", Text: "No notes were found with that query", Seconds: 2})
		}

		// Update table
		rows := make([]table.Row, len(notes))
		for i, note := range notes {
			sentence, morphs, _ := getNoteFields(&note)
			rows[i] = table.Row{
				fmt.Sprintf("#%d", i+1),
				sentence,
				morphs,
				strings.Join(note.Tags, ", "),
			}
		}
		m.table.SetRows(rows)

		if isReload {
			m.table.SetCursor(0)
		}

		// Update NotePage
		if m.isNote {
			note := m.notes[m.table.Cursor()]
			if msg.morphs {
				note = m.morphNotes[m.table.Cursor()]
			}
			_, _, image := getNoteFields(&note)
			m.notePage.imagepath = image
			m.notePage.SetNote(&note)
			m.notePage.image.SetImage(image)

			if core.App.Config.PlayAudioAutomatically {
				m.playAudio(&note)
			}
		}

		return m, nil
	}

	// If the table is at the end, fetch more notes
	// This also works when NotePage is visible
	if m.table.Cursor() == m.currentEnd-1 {
		m.currentEnd += 100
		return m, FetchNotes(core.App.Config.MinningQuery, m.currentEnd, m.currentEnd+100, false)
	}

	if m.isConfig {
		// Exit
		if k == "enter" && m.configPage.focused == len(m.configPage.inputs) {
			err := m.configPage.Save()
			if err != nil {
				panic(err)
			}
			m.isConfig = false
			m.notes = []models.Note{}
			m.currentEnd = 100
			m.table.SetRows([]table.Row{})

			if core.App.Config.MinningQuery == "" {
				return m, core.Log(core.InfoLog{Type: "info", Text: "Minning query is empty.", Seconds: 2})
			}

			return m, FetchNotes(core.App.Config.MinningQuery, 0, 100, false)
		}

		newConfigPage, cmd := m.configPage.Update(msg)
		configPage, _ := newConfigPage.(QueryPageConfig)
		m.configPage = configPage

		return m, cmd
	}

	if m.isNote {

		// if user moves, update the note
		if k == "j" || k == "k" {
			var cmd tea.Cmd
			cmds := make([]tea.Cmd, 0)
			m.table, cmd = m.table.Update(msg)
			cmds = append(cmds, cmd)

			// update note
			note := m.notes[m.table.Cursor()]
			if len(m.morphNotes) > 0 {
				note = m.morphNotes[m.table.Cursor()]
			}
			_, _, image := getNoteFields(&note)
			m.notePage.SetNote(&note)
			m.notePage.image.SetImage(image)

			// play audio if it is enabled
			if core.App.Config.PlayAudioAutomatically {
				m.playAudio(&note)
			}
		}

		var cmd tea.Cmd
		m.notePage, cmd = m.notePage.Update(msg)
		return m, cmd
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

	topbarinfo := fmt.Sprintf("Query: %s \nTotal: %d\n\n", core.App.Config.MinningQuery, len(m.notes))

	var b strings.Builder
	b.WriteString(topbarinfo)
	b.WriteString(m.table.View())
	return lipgloss.PlaceHorizontal(core.App.AvailableWidth, lipgloss.Center, b.String())
}

func (qp *QueryPage) playAudio(note *models.Note) {

	// Stop previous audio
	if qp.audioCtrl != nil {
		speaker.Lock()
		qp.audioCtrl.Paused = true
		speaker.Unlock()
	}

	audioFieldsName := strings.Split(core.App.Config.AudioFieldName, ",")
	for _, fieldName := range audioFieldsName {
		if audioField, ok := note.Fields[fieldName]; ok {
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
			qp.audioCtrl = &beep.Ctrl{Streamer: streamer}
			speaker.Unlock()

			// Add a small silence delay
			silence := generators.Silence(12000)

			speaker.Play(beep.Seq(silence, qp.audioCtrl, beep.Callback(func() {
				streamer.Close()
				reader.Close()
			})))

			break
		}
	}
}

func (qp *QueryPage) setCardAsKnown() error {
	note := qp.notes[qp.table.Cursor()]
	note.Tags = append(note.Tags, core.App.Config.KnownTag)
	return core.App.AnkiConnect.AddTags(note.NoteID, core.App.Config.KnownTag)
}

func getNoteFields(note *models.Note) (string, string, string) {

	var sentence, morphs, image string

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

	imageFieldsName := strings.Split(core.App.Config.ImageFieldName, ",")
	for _, fieldName := range imageFieldsName {
		if imageField, ok := note.Fields[fieldName]; ok {
			imagevalue := imageField.(map[string]interface{})["value"].(string)
			image = filepath.Join(core.App.CollectionPath, imagevalue[10:len(imagevalue)-2])
		}
	}

	return sentence, morphs, image
}
