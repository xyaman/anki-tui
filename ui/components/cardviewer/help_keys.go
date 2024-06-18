package cardviewer

import "github.com/charmbracelet/bubbles/key"

// HelpKeyMap are the key bindings for the cardviewer
type HelpKeyMap struct {
	NextNote  key.Binding
	PrevNote  key.Binding
	OpenNote  key.Binding
	PlayAudio key.Binding
	SeeInAnki key.Binding
	Mine      key.Binding
	Pitch     key.Binding
	Return    key.Binding
}

func (k HelpKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.NextNote, k.PrevNote, k.OpenNote, k.PlayAudio, k.Mine, k.Return}
}

func (k HelpKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		k.ShortHelp(),
		{k.Pitch, k.SeeInAnki},
	}
}

var HelpKeys = HelpKeyMap{
	NextNote: key.NewBinding(
		key.WithKeys("j"),
		key.WithHelp("j", "Next note"),
	),
	PrevNote: key.NewBinding(
		key.WithKeys("k"),
		key.WithHelp("k", "Previous note"),
	),
	PlayAudio: key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p", "Play audio"),
	),
	Mine: key.NewBinding(
		key.WithKeys("ctrl-n"),
		key.WithHelp("ctrl-n", "Mine to Anki"),
	),
	OpenNote: key.NewBinding(
		key.WithKeys("o"),
		key.WithHelp("o", "Open note"),
	),
	SeeInAnki: key.NewBinding(
		key.WithKeys("g"),
		key.WithHelp("g", "See in Anki"),
	),
	Pitch: key.NewBinding(
		key.WithKeys("i"),
		key.WithHelp("i", "Pitch mode"),
	),
	Return: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "Return"),
	),
}
