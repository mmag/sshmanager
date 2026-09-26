/*
* SSH Connection Manager
 */
package main

import (
	"fmt"
	"log"
	"unicode"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// russianToLatin maps letters of the Russian layout to the Latin letters on the same keys
var russianToLatin = map[rune]rune{
	'й': 'q', 'ц': 'w', 'у': 'e', 'к': 'r', 'е': 't', 'н': 'y', 'г': 'u', 'ш': 'i', 'щ': 'o', 'з': 'p',
	'ф': 'a', 'ы': 's', 'в': 'd', 'а': 'f', 'п': 'g', 'р': 'h', 'о': 'j', 'л': 'k', 'д': 'l',
	'я': 'z', 'ч': 'x', 'с': 'c', 'м': 'v', 'и': 'b', 'т': 'n', 'ь': 'm',
}

// latinCtrlKey turns Ctrl with a Russian letter into Ctrl with the Latin letter on the same key.
// Terminals with the extended keyboard protocol report the letter of the current layout,
// so Ctrl+E arrives as Ctrl+У.
func latinCtrlKey(event *tcell.EventKey) *tcell.EventKey {
	if event.Key() != tcell.KeyRune || event.Modifiers() != tcell.ModCtrl {
		return event
	}
	if r, ok := russianToLatin[unicode.ToLower(event.Rune())]; ok {
		return tcell.NewEventKey(tcell.KeyRune, r, tcell.ModCtrl)
	}
	return event
}

// isSSHCheckKey matches Ctrl+Enter and the Ctrl+T fallback. Terminals that tell
// Ctrl+Enter apart from Enter usually send it as Ctrl+J, the rest send plain Enter.
func isSSHCheckKey(event *tcell.EventKey) bool {
	switch event.Key() {
	case tcell.KeyCtrlJ, tcell.KeyCtrlT:
		return true
	case tcell.KeyEnter:
		return event.Modifiers()&tcell.ModCtrl != 0
	}
	return false
}

// handleKey processes shortcuts of the main screen; dialogs and forms receive keys untouched
func handleKey(event *tcell.EventKey) *tcell.EventKey {
	if converted := latinCtrlKey(event); converted != event {
		// tview stops on Ctrl+C only when the event wasn't replaced
		if converted.Key() == tcell.KeyCtrlC {
			app.Stop()
			return nil
		}
		event = converted
	}
	if !mainScreen {
		return event
	}

	focus := app.GetFocus()
	if isSSHCheckKey(event) {
		if index := selectedConnection(); focus == connectionsList && index >= 0 {
			checkLogin(currentConnections()[index])
		}
		return nil
	}

	switch event.Key() {
	case tcell.KeyCtrlR:
		// Redraw window and recheck hosts
		refreshConnectionsList(connectionsList.GetCurrentItem())
		checkHosts(allConnections())
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
	case tcell.KeyLeft:
		switchTab((currentTab + len(config.Tabs) - 1) % len(config.Tabs))
		return nil
	case tcell.KeyRight:
		switchTab((currentTab + 1) % len(config.Tabs))
		return nil
	case tcell.KeyRune:
		// 1-9 jump to a tab
		if r := event.Rune(); r >= '1' && r <= '9' {
			switchTab(int(r - '1'))
			return nil
		}
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
			confirm(fmt.Sprintf(currentLang["dlg_edit"], currentConnections()[index].Server), func() {
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
	checkHosts(allConnections())
	startPeriodicChecks()
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
