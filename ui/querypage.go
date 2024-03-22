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
	table  table.Model
	width  int
	height int

	currentEnd int

	notes []models.Note

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
		notePage:   NewNotePage(),
		configPage: NewQueryPageConfig(),
		isConfig:   false,
		currentEnd: 100,
	}
}

func (m QueryPage) Init() tea.Cmd {
	return tea.Batch(FetchNotes(core.App.Config.MinningQuery, 0, 100), m.configPage.Init())
}

func (m QueryPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	var k string

	switch msg := msg.(type) {
	case tea.KeyMsg:
		k = msg.String()
		if k == "c" && !m.isConfig {
			m.isConfig = true
			return m, textinput.Blink
		} else if k == "esc" && (m.isConfig || m.isNote) {
			m.isConfig = false
			m.isNote = false
			return m, nil
		}

		switch k {
		case "p":
			note := m.notes[m.table.Cursor()]
			m.playAudio(&note)
		case "o":
      if m.isConfig{
        break
      }
			note := m.notes[m.table.Cursor()]
			m.notePage.note = &note
			m.isNote = true

			_, _, image := getNoteFields(&note)
			m.notePage.imagepath = image
			m.notePage.note = &note

      // TODO: Change this to global
      m.notePage.image.SetSize(50, 50)
      m.notePage.image.SetImage(image)

      // play audio
      if core.App.Config.PlayAudioAutomatically {
        m.playAudio(&note)
      }

      return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		m.table.SetHeight(core.App.AvailableHeight - 10)

    return m, nil
    

	case FetchNotesMsg:
		isReload := len(m.notes) == 0
		m.notes = append(m.notes, msg.notes...)
		if len(m.notes) == 0 {
			// return m, core.Log(core.InfoLog{Type: "info", Text: "No notes found.", Seconds: 2})
			return m, tea.Batch(
				core.Log(core.InfoLog{Type: "info", Text: "No notes found 1!!.", Seconds: 5}),
				core.Log(core.InfoLog{Type: "error", Text: "No notes found.", Seconds: 2}),
			)
		}

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

		if isReload {
			m.table.SetCursor(0)
		}
	}

	// Table
	if m.table.Cursor() == m.currentEnd-1 {
		m.currentEnd += 100
		return m, FetchNotes(core.App.Config.MinningQuery, m.currentEnd, m.currentEnd+100)
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
			return m, FetchNotes(core.App.Config.MinningQuery, 0, 100)
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
      _, _, image := getNoteFields(&note)
      m.notePage.note = &note
      m.notePage.image.SetImage(image)

      // play audio if it is enabled
      if core.App.Config.PlayAudioAutomatically {
        m.playAudio(&note)
      }
    }


		newNotePage, cmd := m.notePage.Update(msg)
		m.notePage = newNotePage
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
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, renderConfig)
	}

	if m.isNote {
		return m.notePage.View()
	}

	topbarinfo := fmt.Sprintf("Query: %s \nTotal: %d\n\n", core.App.Config.MinningQuery, len(m.notes))

	var b strings.Builder
	b.WriteString(topbarinfo)
	b.WriteString(m.table.View())
	renderTable := baseStyle.Render(b.String())
	return lipgloss.PlaceHorizontal(m.width, lipgloss.Center, renderTable)
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
