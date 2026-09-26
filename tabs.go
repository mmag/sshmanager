package main

import (
	"fmt"
	"slices"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const maxTabNameLength = 20

var (
	tabBar     *TabBar
	currentTab int
)

// TabBar shows tab names in one row; clicking a name switches to that tab
type TabBar struct {
	*tview.Box
	// spans holds the [start, end) columns of each drawn tab relative to the bar
	spans [][2]int
}

func NewTabBar() *TabBar {
	bar := &TabBar{Box: tview.NewBox()}
	bar.SetBackgroundColor(tcell.ColorNavy)
	return bar
}

func (b *TabBar) Draw(screen tcell.Screen) {
	b.DrawForSubclass(screen, b)
	x, y, width, _ := b.GetInnerRect()

	b.spans = b.spans[:0]
	pos := 0
	for i, tab := range config.Tabs {
		style := "[white:darkblue]"
		if i == currentTab {
			style = "[white:darkred]"
		}
		_, drawn := tview.Print(screen, style+" "+tview.Escape(tab.Name)+" ", x+pos, y, width-pos, tview.AlignLeft, tcell.ColorWhite)
		b.spans = append(b.spans, [2]int{pos, pos + drawn})
		pos += drawn + 1
	}
}

func (b *TabBar) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
	return b.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
		if !b.InRect(event.Position()) {
			return false, nil
		}
		// Two quick clicks come as a double click even on different tabs
		if action == tview.MouseLeftClick || action == tview.MouseLeftDoubleClick {
			x, _ := event.Position()
			rectX, _, _, _ := b.GetInnerRect()
			for i, span := range b.spans {
				if x-rectX >= span[0] && x-rectX < span[1] {
					switchTab(i)
					break
				}
			}
		}
		// Keep focus on the lists
		return true, nil
	})
}

// switchTab shows the connections of the tab with the given index
func switchTab(index int) {
	if index < 0 || index >= len(config.Tabs) {
		return
	}
	currentTab = index
	refreshConnectionsList(0)
}

func tabNames() []string {
	names := make([]string, len(config.Tabs))
	for i, tab := range config.Tabs {
		names[i] = tab.Name
	}
	return names
}

// showTabForm displays a form for creating (index -1) or renaming a tab
func showTabForm(index int) {
	var name string
	title := currentLang["title_tab_add"]
	if index >= 0 {
		name = config.Tabs[index].Name
		title = currentLang["title_tab_rename"]
	}

	errorText := newErrorText()
	form := newForm()
	form.
		AddInputField(currentLang["form_tab_name"], name, maxTabNameLength, tview.InputFieldMaxLength(maxTabNameLength), nil).
		AddButton(currentLang["btn_save"], func() {
			newName := inputText(form, 0)
			switch {
			case newName == "":
				errorText.SetText(currentLang["msg_enter_tab_name"])
				return
			case newName != name && tabExists(newName):
				errorText.SetText(currentLang["msg_tab_exists"])
				return
			}

			if index < 0 {
				config.Tabs = append(config.Tabs, Tab{Name: newName})
				index = len(config.Tabs) - 1
			} else {
				config.Tabs[index].Name = newName
			}
			saveConfig()
			switchTab(index)
			showMain()
		}).
		AddButton(currentLang["btn_cancel"], showMain)

	showForm(title, form, errorText)
}

// deleteTab asks for confirmation and removes the current tab. Its connections
// are either deleted too or moved to the first tab, which can't be deleted.
func deleteTab() {
	if currentTab == 0 {
		showModal(currentLang["msg_first_tab"], []string{currentLang["btn_ok"]}, func(string) {
			showMain()
		})
		return
	}

	index := currentTab
	tab := config.Tabs[index]
	if len(tab.Connections) == 0 {
		confirm(fmt.Sprintf(currentLang["dlg_tab_delete"], tab.Name), func() {
			removeTab(index, false)
		})
		return
	}

	deleteAll := currentLang["btn_tab_delete_all"]
	moveToFirst := fmt.Sprintf(currentLang["btn_tab_move"], config.Tabs[0].Name)
	showModal(fmt.Sprintf(currentLang["dlg_tab_delete_conns"], tab.Name, len(tab.Connections)),
		[]string{deleteAll, moveToFirst, currentLang["btn_cancel"]},
		func(label string) {
			switch label {
			case deleteAll:
				removeTab(index, false)
			case moveToFirst:
				removeTab(index, true)
			default:
				showMain()
			}
		})
}

// removeTab deletes the tab, moving its connections to the first tab if requested
func removeTab(index int, moveConnections bool) {
	tab := config.Tabs[index]
	if moveConnections {
		config.Tabs[0].Connections = append(config.Tabs[0].Connections, tab.Connections...)
	} else {
		for _, conn := range tab.Connections {
			delete(hostStatuses, conn.Server)
		}
	}
	config.Tabs = slices.Delete(config.Tabs, index, index+1)
	saveConfig()

	if moveConnections {
		switchTab(0)
	} else {
		switchTab(index - 1)
	}
	showMain()
}
