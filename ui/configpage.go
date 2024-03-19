package ui

import (
	"github.com/rivo/tview"

	"github.com/xyaman/anki-tui/core"
)

type ConfigPage struct {
  TopBar *tview.TextView
	GlobalForm *tview.Form
  QueryForm *tview.Form
}

func ShowConfigPage() {
  configPage := newConfigPage()

  if core.App.PageHolder.HasPage(core.ConfigPageID) {
    core.App.PageHolder.RemovePage(core.ConfigPageID)
  }

  core.App.PageHolder.AddAndSwitchToPage(core.ConfigPageID, configPage.QueryForm, true)
  core.App.Tview.SetFocus(configPage.QueryForm)
}

func newConfigPage() *ConfigPage {
  return &ConfigPage{
    TopBar: tview.NewTextView(),
    // GlobalForm: setupGlobalForm(),
    QueryForm: setupQueryForm(),
  }
}

func setupQueryForm() *tview.Form {

	queryForm := tview.NewForm()
	queryForm.AddInputField("Minning Query", core.App.Config.MinningQuery, 0, nil, nil).
		AddInputField("Search Query", core.App.Config.SearchQuery, 0, nil, nil).
		AddInputField("Morph Field Name", core.App.Config.MorphFieldName, 0, nil, nil).
		AddInputField("Sentence Field Name", core.App.Config.SentenceFieldName, 0, nil, nil).
		AddInputField("Audio Field Name", core.App.Config.AudioFieldName, 0, nil, nil).
		AddInputField("Image Field Name", core.App.Config.ImageFieldName, 0, nil, nil).
		AddInputField("Known Tag", core.App.Config.KnownTag, 0, nil, nil).
		AddInputField("Minning Audio Field Name", core.App.Config.MinningAudioFieldName, 0, nil, nil).
		AddInputField("Minning Image Field Name", core.App.Config.MinningImageFieldName, 0, nil, nil).
		AddCheckbox("Play Audio Automatically", core.App.Config.PlayAudioAutomatically, nil).
		AddButton("Save", func() {
			core.App.Config.MinningQuery = queryForm.GetFormItemByLabel("Minning Query").(*tview.InputField).GetText()
			core.App.Config.SearchQuery = queryForm.GetFormItemByLabel("Search Query").(*tview.InputField).GetText()
			core.App.Config.MorphFieldName = queryForm.GetFormItemByLabel("Morph Field Name").(*tview.InputField).GetText()
			core.App.Config.SentenceFieldName = queryForm.GetFormItemByLabel("Sentence Field Name").(*tview.InputField).GetText()
			core.App.Config.AudioFieldName = queryForm.GetFormItemByLabel("Audio Field Name").(*tview.InputField).GetText()
			core.App.Config.ImageFieldName = queryForm.GetFormItemByLabel("Image Field Name").(*tview.InputField).GetText()
			core.App.Config.KnownTag = queryForm.GetFormItemByLabel("Known Tag").(*tview.InputField).GetText()
			core.App.Config.MinningAudioFieldName = queryForm.GetFormItemByLabel("Minning Audio Field Name").(*tview.InputField).GetText()
			core.App.Config.MinningImageFieldName = queryForm.GetFormItemByLabel("Minning Image Field Name").(*tview.InputField).GetText()
			core.App.Config.PlayAudioAutomatically = queryForm.GetFormItemByLabel("Play Audio Automatically").(*tview.Checkbox).IsChecked()

			err := core.App.Config.Save()
			if err != nil {
			  panic(err)
			}
  
      ShowQueryPage()

    }).
    AddButton("Quit", func() {
      core.App.Tview.Stop()
    })

    return queryForm
}
