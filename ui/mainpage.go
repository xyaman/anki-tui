package ui

import (
	"github.com/rivo/tview"
	"github.com/xyaman/anki-tui/core"
)

type MainPage struct {
	Flex *tview.Flex
  List *tview.List
}

func NewMainPage() *MainPage {
  flex := tview.NewFlex()

  list := tview.NewList().
  AddItem("Query", "See all cards based on a query", 'a', func() {
    ShowQueryPage()
  }).
  AddItem("Morph Query", "See all cards based on a morph query", 'b', func() {
    //
  }).
  AddItem("Config", "Change the configuration", 'c', func() {
    ShowConfigPage()
  })

  flex.AddItem(list, 0, 1, true)

  return &MainPage{
    Flex: flex,
    List: list,
  }
}

func ShowMainPage() {
  mainPage := NewMainPage()
  core.App.PageHolder.AddAndSwitchToPage(core.MainPageID, mainPage.Flex, true)
  core.App.Tview.SetFocus(mainPage.List)
}

