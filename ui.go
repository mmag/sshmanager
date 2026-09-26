package main

import (
	"fmt"
	"slices"
	"strings"

	"sshman/lang"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	formWidth    = 100
	tabBarHeight = 1
	helpHeight   = 8
)

var (
	app             *tview.Application
	connectionsList *tview.List
	menuList        *tview.List
	helpText        *tview.TextView
	// mainScreen is true while the main layout is shown rather than a dialog or form
	mainScreen bool
)

// setupDebianTheme configures the Debian installer color scheme
func setupDebianTheme() {
	tview.Styles.PrimitiveBackgroundColor = tcell.ColorNavy
	tview.Styles.ContrastBackgroundColor = tcell.ColorDarkRed
	tview.Styles.MoreContrastBackgroundColor = tcell.ColorRed
	tview.Styles.BorderColor = tcell.ColorWhite
	tview.Styles.TitleColor = tcell.ColorWhite
	tview.Styles.GraphicsColor = tcell.ColorWhite
	tview.Styles.PrimaryTextColor = tcell.ColorWhite
	tview.Styles.SecondaryTextColor = tcell.ColorLightGray
	tview.Styles.TertiaryTextColor = tcell.ColorGreen
	tview.Styles.InverseTextColor = tcell.ColorBlack
	tview.Styles.ContrastSecondaryTextColor = tcell.ColorWhite
}

func newList() *tview.List {
	list := tview.NewList().ShowSecondaryText(false)
	list.SetBorder(true).SetTitleAlign(tview.AlignLeft)
	list.SetBackgroundColor(tcell.ColorNavy)
	list.SetMainTextColor(tcell.ColorWhite)
	list.SetSelectedTextColor(tcell.ColorWhite)
	list.SetSelectedBackgroundColor(tcell.ColorDarkRed)
	return list
}

// setupUI creates the main screen widgets
func setupUI() {
	tabBar = NewTabBar()
	connectionsList = newList()
	connectionsList.SetUseStyleTags(true, false)
	menuList = newList()

	helpText = tview.NewTextView().SetTextAlign(tview.AlignLeft)
	helpText.SetBackgroundColor(tcell.ColorNavy)
	helpText.SetTextColor(tcell.ColorWhite)

	applyLanguage()
	refreshConnectionsList(0)

	// Auto-scroll when reaching visible area boundaries
	connectionsList.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		_, _, _, visibleItems := connectionsList.GetInnerRect()
		offset, _ := connectionsList.GetOffset()
		if index < offset {
			connectionsList.SetOffset(index, 0)
		} else if index >= offset+visibleItems {
			connectionsList.SetOffset(index-visibleItems+1, 0)
		}
	})

	// Only the focused list shows its selection
	connectionsList.SetFocusFunc(func() {
		connectionsList.SetSelectedBackgroundColor(tcell.ColorDarkRed)
		menuList.SetSelectedBackgroundColor(tcell.ColorNavy)
	})
	menuList.SetFocusFunc(func() {
		menuList.SetSelectedBackgroundColor(tcell.ColorDarkRed)
		connectionsList.SetSelectedBackgroundColor(tcell.ColorNavy)
	})
}

// applyLanguage updates titles, menu items and help text for the current language
func applyLanguage() {
	connectionsList.SetTitle(currentLang["connections_title"])
	menuList.SetTitle(currentLang["menu_title"])
	helpText.SetText(currentLang["help_text"])

	menuList.Clear()
	menuList.AddItem(" "+currentLang["menu_add"], "", 0, func() { showConnectionForm(-1) })
	menuList.AddItem(" "+currentLang["menu_tab_add"], "", 0, func() { showTabForm(-1) })
	menuList.AddItem(" "+currentLang["menu_tab_rename"], "", 0, func() { showTabForm(currentTab) })
	menuList.AddItem(" "+currentLang["menu_tab_delete"], "", 0, deleteTab)
	menuList.AddItem(" "+currentLang["menu_language"], "", 0, switchLanguage)
	menuList.AddItem(" "+currentLang["menu_edit_config"], "", 0, openConfig)
	menuList.AddItem(" "+currentLang["menu_exit"], "", 0, app.Stop)
}

// layoutHeights returns the heights of the connections list and the menu including borders.
// The list fits the largest tab, so switching tabs doesn't resize the window.
func layoutHeights() (connections, menu int) {
	largest := 0
	for _, tab := range config.Tabs {
		largest = max(largest, len(tab.Connections))
	}
	return largest + 3, menuList.GetItemCount() + 2
}

func mainLayout() *tview.Flex {
	connectionsHeight, menuHeight := layoutHeights()
	return tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(tabBar, tabBarHeight, 0, false).
		AddItem(connectionsList, connectionsHeight, 0, true).
		AddItem(menuList, menuHeight, 0, false).
		AddItem(helpText, helpHeight, 0, false)
}

// centerWidget places the widget in the middle of the screen
// in a box of the same size as the main layout
func centerWidget(widget tview.Primitive) *tview.Flex {
	connectionsHeight, menuHeight := layoutHeights()
	height := tabBarHeight + connectionsHeight + menuHeight + helpHeight

	flex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().
			SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(widget, height, 0, true).
			AddItem(nil, 0, 1, false), formWidth, 0, true).
		AddItem(nil, 0, 1, false)
	flex.SetBackgroundColor(tcell.ColorNavy)
	return flex
}

// showMain displays the main screen with the connections list focused
func showMain() {
	mainScreen = true
	app.SetRoot(centerWidget(mainLayout()), true)
	app.SetFocus(connectionsList)
}

// showDialog displays a dialog or form in place of the main screen
func showDialog(widget tview.Primitive) {
	mainScreen = false
	app.SetRoot(centerWidget(widget), true)
}

// showModal displays a message with buttons; done receives the pressed
// button label, or an empty string when the dialog is closed with Esc
func showModal(text string, buttons []string, done func(label string)) {
	modal := tview.NewModal()
	modal.SetBackgroundColor(tcell.ColorNavy)
	modal.SetTextColor(tcell.ColorWhite)
	modal.SetButtonBackgroundColor(tcell.ColorDarkRed)
	modal.SetButtonTextColor(tcell.ColorWhite)
	modal.
		SetText(text).
		AddButtons(buttons).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			done(buttonLabel)
		})
	showDialog(modal)
}

// confirm asks an OK/Cancel question and runs onOK on confirmation,
// otherwise returns to the main screen
func confirm(text string, onOK func()) {
	showModal(text, []string{currentLang["btn_ok"], currentLang["btn_cancel"]}, func(label string) {
		if label == currentLang["btn_ok"] {
			onOK()
			return
		}
		showMain()
	})
}

// newForm creates a form in the Debian installer colors; Esc returns to the main screen
func newForm() *tview.Form {
	form := tview.NewForm()
	form.SetBackgroundColor(tcell.ColorNavy)
	form.SetFieldBackgroundColor(tcell.ColorDarkBlue)
	form.SetFieldTextColor(tcell.ColorWhite)
	form.SetLabelColor(tcell.ColorWhite)
	form.SetButtonBackgroundColor(tcell.ColorDarkRed)
	form.SetButtonTextColor(tcell.ColorWhite)
	form.SetCancelFunc(showMain)
	return form
}

// showForm displays the form in a titled frame with an error line below it
func showForm(title string, form *tview.Form, errorText *tview.TextView) {
	frame := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(form, 0, 1, true).
		AddItem(errorText, 1, 0, false)

	frame.SetBorder(true).
		SetTitle(title).
		SetTitleAlign(tview.AlignLeft).
		SetBackgroundColor(tcell.ColorNavy).
		SetBorderColor(tcell.ColorWhite).
		SetTitleColor(tcell.ColorWhite)
	showDialog(frame)
}

func newErrorText() *tview.TextView {
	errorText := tview.NewTextView()
	errorText.SetTextColor(tcell.ColorYellow).SetBackgroundColor(tcell.ColorNavy)
	return errorText
}

func inputText(form *tview.Form, index int) string {
	return strings.TrimSpace(form.GetFormItem(index).(*tview.InputField).GetText())
}

// formatConnectionLine formats connection info with dots between address and description
func formatConnectionLine(conn SSHConnection) string {
	serverPart := conn.Server
	if conn.Port != "" && conn.Port != "22" {
		serverPart = serverPart + ":" + conn.Port
	}
	if conn.Username != "" {
		serverPart = conn.Username + "@" + serverPart
	}
	serverPart = tview.Escape(serverPart)
	comment := tview.Escape(conn.Comment)

	// formWidth - 2 (borders) - 2 (status symbol + space)
	totalWidth := formWidth - 4
	dotsCount := totalWidth - tview.TaggedStringWidth(serverPart) - tview.TaggedStringWidth(comment)

	// If both parts fit with at least 3 dots, use dots
	if dotsCount >= 3 {
		return fmt.Sprintf("%s %s%s%s", statusSymbol(conn.Server), serverPart, strings.Repeat(".", dotsCount), comment)
	}
	return fmt.Sprintf("%s %s - %s", statusSymbol(conn.Server), serverPart, comment)
}

// refreshConnectionsList rebuilds the connections list and selects the given item
func refreshConnectionsList(selected int) {
	connectionsList.Clear()

	connections := currentConnections()
	if len(connections) == 0 {
		connectionsList.AddItem(currentLang["msg_no_connections"], "", 0, nil)
		return
	}

	for i, conn := range connections {
		connectionsList.AddItem(formatConnectionLine(conn), "", 0, func() { connectTo(i) })
	}
	connectionsList.SetCurrentItem(max(0, min(selected, len(connections)-1)))
}

// updateStatusSymbols redraws status symbols without rebuilding the list
func updateStatusSymbols() {
	connections := currentConnections()
	if connectionsList.GetItemCount() != len(connections) {
		return
	}
	for i, conn := range connections {
		connectionsList.SetItemText(i, formatConnectionLine(conn), "")
	}
}

// selectedConnection returns the index of the highlighted connection, or -1 if there is none
func selectedConnection() int {
	index := connectionsList.GetCurrentItem()
	if index < 0 || index >= len(currentConnections()) {
		return -1
	}
	return index
}

// connectTo asks for confirmation and runs ssh with the application suspended
func connectTo(index int) {
	conn := currentConnections()[index]
	confirm(fmt.Sprintf(currentLang["dlg_connect"], conn.Server), func() {
		pauseChecks()
		app.Suspend(func() {
			sshConnect(conn)
		})
		resumeChecks()
		showMain()
	})
}

// showConnectionForm displays a form for adding (index -1) or editing
// a connection of the current tab; the connection can be moved to another tab
func showConnectionForm(index int) {
	var conn SSHConnection
	title := currentLang["title_add"]
	sourceTab := currentTab
	if index >= 0 {
		conn = currentConnections()[index]
		title = currentLang["title_edit"]
	}

	errorText := newErrorText()
	form := newForm()
	form.
		AddInputField(currentLang["form_server"], conn.Server, 30, nil, func(text string) {
			if text != conn.Server && connectionExists(text) {
				errorText.SetText(currentLang["msg_conn_exists"])
			} else {
				errorText.SetText("")
			}
		}).
		AddInputField(currentLang["form_port"], conn.Port, 5, nil, nil).
		AddInputField(currentLang["form_comment"], conn.Comment, 30, nil, nil).
		AddInputField(currentLang["form_username"], conn.Username, 20, nil, nil).
		AddDropDown(currentLang["form_tab"], tabNames(), sourceTab, nil).
		AddButton(currentLang["btn_save"], func() {
			updated := SSHConnection{
				Server:   inputText(form, 0),
				Port:     inputText(form, 1),
				Comment:  inputText(form, 2),
				Username: inputText(form, 3),
			}
			targetTab, _ := form.GetFormItem(4).(*tview.DropDown).GetCurrentOption()

			switch {
			case updated.Server == "":
				errorText.SetText(currentLang["msg_enter_server"])
				return
			case updated.Comment == "":
				errorText.SetText(currentLang["msg_enter_comment"])
				return
			case !isValidPort(updated.Port):
				errorText.SetText(currentLang["msg_invalid_port"])
				return
			case updated.Server != conn.Server && connectionExists(updated.Server):
				errorText.SetText(currentLang["msg_conn_exists"])
				return
			}

			if index >= 0 {
				if updated.Server != conn.Server {
					delete(hostStatuses, conn.Server)
				}
				if targetTab == sourceTab {
					config.Tabs[sourceTab].Connections[index] = updated
				} else {
					config.Tabs[sourceTab].Connections = slices.Delete(config.Tabs[sourceTab].Connections, index, index+1)
					index = -1
				}
			}
			if index < 0 {
				config.Tabs[targetTab].Connections = append(config.Tabs[targetTab].Connections, updated)
				index = len(config.Tabs[targetTab].Connections) - 1
			}
			saveConfig()
			currentTab = targetTab
			refreshConnectionsList(index)
			checkHosts([]SSHConnection{updated})
			showMain()
		}).
		AddButton(currentLang["btn_cancel"], showMain)

	form.GetFormItem(4).(*tview.DropDown).SetListStyles(
		tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorDarkBlue),
		tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorDarkRed),
	)
	showForm(title, form, errorText)
}

// deleteConnection asks for confirmation and removes the connection from the current tab
func deleteConnection(index int) {
	server := currentConnections()[index].Server
	confirm(fmt.Sprintf(currentLang["dlg_delete"], server), func() {
		tab := &config.Tabs[currentTab]
		tab.Connections = slices.Delete(tab.Connections, index, index+1)
		delete(hostStatuses, server)
		saveConfig()
		refreshConnectionsList(index)
		showMain()
	})
}

// switchLanguage lets the user choose the interface language and saves the choice
func switchLanguage() {
	showModal("Select language / Выберите язык", []string{"English", "Русский"}, func(label string) {
		switch label {
		case "English":
			currentLang = lang.EN
		case "Русский":
			currentLang = lang.RU
		default:
			showMain()
			return
		}
		// The first tab follows the language until the user renames it
		if name := config.Tabs[0].Name; name == lang.EN["tab_default"] || name == lang.RU["tab_default"] {
			config.Tabs[0].Name = currentLang["tab_default"]
		}
		applyLanguage()
		refreshConnectionsList(connectionsList.GetCurrentItem())
		saveConfig()
		showMain()
	})
}
