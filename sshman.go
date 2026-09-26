/*
* SSH Connection Manager
 */
package main

import (
	"fmt"
	"log"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// handleKey processes shortcuts of the main screen; dialogs and forms receive keys untouched
func handleKey(event *tcell.EventKey) *tcell.EventKey {
	if !mainScreen {
		return event
	}

	focus := app.GetFocus()
	switch event.Key() {
	case tcell.KeyCtrlR:
		// Redraw window and recheck hosts
		refreshConnectionsList(connectionsList.GetCurrentItem())
		checkHosts(config.Connections)
		showMain()
		app.SetFocus(focus)
		return nil
	case tcell.KeyTab:
		// Switch between lists
		if focus == connectionsList {
			app.SetFocus(menuList)
		} else {
			app.SetFocus(connectionsList)
		}
		return nil
	case tcell.KeyDown:
		// Wrap around at the end
		if list, ok := focus.(*tview.List); ok && list.GetCurrentItem() == list.GetItemCount()-1 {
			list.SetCurrentItem(0)
			return nil
		}
	case tcell.KeyUp:
		// Wrap around at the beginning
		if list, ok := focus.(*tview.List); ok && list.GetCurrentItem() == 0 {
			list.SetCurrentItem(list.GetItemCount() - 1)
			return nil
		}
	case tcell.KeyCtrlE:
		if index := selectedConnection(); focus == connectionsList && index >= 0 {
			confirm(fmt.Sprintf(currentLang["dlg_edit"], config.Connections[index].Server), func() {
				showConnectionForm(index)
			})
		}
		return nil
	case tcell.KeyCtrlN:
		confirm(currentLang["dlg_add"], func() {
			showConnectionForm(-1)
		})
		return nil
	case tcell.KeyDelete:
		if index := selectedConnection(); focus == connectionsList && index >= 0 {
			deleteConnection(index)
		}
		return nil
	}
	return event
}

// startApp loads the configuration and shows the main screen
func startApp() {
	setupDebianTheme()
	loadConfig()
	setupUI()

	app.SetInputCapture(handleKey)
	checkHosts(config.Connections)
	showMain()
}

// main initializes and runs the SSH connection manager application
func main() {
	app = tview.NewApplication()
	startApp()

	if err := app.EnableMouse(true).Run(); err != nil {
		log.Fatalf(currentLang["msg_app_error"], err)
	}
}
