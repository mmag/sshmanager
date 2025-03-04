/*
* SSH Connection Manager
 */
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"sshmanager/lang"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Update the config structure by adding a new type
type Config struct {
	Connections []SSHConnection `json:"connections"`
	Language    string          `json:"language"`
}

type SSHConnection struct {
	Server  string `json:"server"`
	Comment string `json:"comment"`
	Port    string `json:"port"`
}

// Update global variables
var (
	sshConnections []SSHConnection
	configDir      = filepath.Join(os.Getenv("HOME"), "sshman")
	configFilePath = filepath.Join(configDir, "sshman.json")
	menuList       *tview.List
	helpText       *tview.TextView
	config         Config // Add config variable
)

// Add constants for form dimensions
const (
	formWidth  = 100 // increase width for better readability
	formHeight = 60  // decrease height for compactness
)

// Add global variables
var (
	currentLang       = lang.EN // Default language
	connectionsHeight = 0       // Height of connections list
)

// createMainLayout creates and returns the main application layout with the specified heights
// for connections list, menu, and help sections based on screen height
func createMainLayout(app *tview.Application, connectionsList *tview.List) *tview.Flex {
	// Set fixed terminal height
	const screenHeight = 80

	// Calculate menu height (items count + border)
	menuHeight := menuList.GetItemCount() + 2

	// Calculate help height (text lines + border)
	helpHeight := 8 // 6 text lines + top and bottom borders
	// Calculate connections list height
	connectionsHeight = screenHeight/2 - menuHeight - helpHeight
	connectionsHeight := screenHeight/2 - menuHeight - helpHeight

	// Create vertical flex for lists and help text
	return tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(connectionsList, connectionsHeight, 0, true).
		AddItem(menuList, menuHeight, 0, false).
		AddItem(helpText, helpHeight, 0, false)
}

// sshConnect establishes an SSH connection to the specified server using the saved configuration
// It supports custom port specification and handles connection errors
func sshConnect(server string) {
	var connection SSHConnection
	for _, conn := range sshConnections {
		if conn.Server == server {
			connection = conn
			break
		}
	}
	sshCommand := "ssh"
	if connection.Port != "" {
		sshCommand = fmt.Sprintf("ssh -p %s", connection.Port)
	}

	cmd := exec.Command("sh", "-c", fmt.Sprintf("%s %s", sshCommand, connection.Server))
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Printf(currentLang["msg_connecting"], connection.Server)
	err := cmd.Run()
	if err != nil {
		log.Printf(currentLang["msg_conn_error"], connection.Server, err)
	}
}

// centerWidget centers the provided widget in the screen with specified form dimensions
func centerWidget(widget tview.Primitive) *tview.Flex {
	widget.SetRect(0, 0, formWidth, formHeight)
	return tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().
			SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(widget, formHeight, 0, true).
			AddItem(nil, 0, 1, false), formWidth, 0, true).
		AddItem(nil, 0, 1, false)
}

// isConnectionExists checks if a connection with the given server address already exists
// Returns true if the connection exists, false otherwise
func isConnectionExists(server string) bool {
	for _, conn := range sshConnections {
		if conn.Server == server {
			return true
		}
	}
	return false
}

// addConnection displays a form for adding a new SSH connection
// Validates input and saves the new connection to the configuration
func addConnection(app *tview.Application, connectionsList *tview.List) {
	var form *tview.Form
	errorText := tview.NewTextView().SetText("")
	form = tview.NewForm().
		AddInputField(currentLang["form_server"], "", 30, nil, func(text string) {
			if text == "" {
				return
			}
			if isConnectionExists(text) {
				errorText.SetText(currentLang["msg_conn_exists"])
				return
			}
			errorText.SetText("")
		}).
		AddInputField(currentLang["form_port"], "", 5, nil, nil).
		AddInputField(currentLang["form_comment"], "", 30, nil, nil).
		AddButton(currentLang["btn_save"], func() {
			server := form.GetFormItem(0).(*tview.InputField).GetText()
			port := form.GetFormItem(1).(*tview.InputField).GetText()
			comment := form.GetFormItem(2).(*tview.InputField).GetText()

			if server == "" {
				errorText.SetText(currentLang["msg_enter_server"])
				return
			}
			if comment == "" {
				errorText.SetText(currentLang["msg_enter_comment"])
				return
			}

			if !isConnectionExists(server) {
				connection := SSHConnection{Server: server, Port: port, Comment: comment}
				sshConnections = append(sshConnections, connection)

				// Add new item
				newIndex := connectionsList.GetItemCount()
				mainText, _ := connectionsList.GetItemText(0)
				if newIndex == 1 && mainText == currentLang["msg_no_connections"] {
					connectionsList.Clear()
					newIndex = 0
				}
				connectionsList.AddItem(fmt.Sprintf("%s - %s", server, comment), "", 0, func() {
					showMessage(app, connectionsList, server)
				})
				saveConnections()

				// Return to main screen
				app.SetRoot(centerWidget(createMainLayout(app, connectionsList)), true)
				connectionsList.SetCurrentItem(newIndex)
			}
		}).
		AddButton(currentLang["btn_cancel"], func() {
			app.SetRoot(centerWidget(createMainLayout(app, connectionsList)), true)
		})

	// Create flex for form with error text
	formFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(form, 0, 1, true).
		AddItem(errorText, 1, 0, false)

	formFlex.SetBorder(true).
		SetTitle(currentLang["title_add"]).
		SetTitleAlign(tview.AlignLeft)

	// Set form as active widget
	app.SetRoot(centerWidget(formFlex), true)
	app.SetFocus(form)
}

// saveConnections writes the current connections list to the configuration file
// Creates the config directory if it doesn't exist
func saveConnections() {
	// Ensure the config directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		log.Printf(currentLang["msg_config_dir_error"], err)
		return
	}

	// Update config before saving
	config.Connections = sshConnections
	config.Language = "en"
	if currentLang["language_code"] == "ru" {
		config.Language = "ru"
	}

	// Use MarshalIndent for formatted output
	data, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		log.Printf(currentLang["msg_save_error"], err)
		return
	}

	data = append(data, '\n')
	err = ioutil.WriteFile(configFilePath, data, 0644)
	if err != nil {
		log.Printf(currentLang["msg_write_error"], err)
	}
}

// loadConnections reads and parses the SSH connections from the configuration file
// Silently handles the case when the config file doesn't exist
func loadConnections() {
	data, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf(currentLang["msg_read_error"], err)
		}
		return
	}

	err = json.Unmarshal(data, &config)
	if err != nil {
		log.Printf(currentLang["msg_parse_error"], err)
		return
	}

	// Set connections from config
	sshConnections = config.Connections

	// Set language from config
	if config.Language == "ru" {
		currentLang = lang.RU
	} else {
		currentLang = lang.EN
	}
}

// showMessage displays a confirmation dialog before establishing an SSH connection
// Suspends the application while the SSH connection is active
func showMessage(app *tview.Application, list *tview.List, server string) {
	modal := tview.NewModal().
		SetText(fmt.Sprintf(currentLang["dlg_connect"], server)).
		AddButtons([]string{currentLang["btn_ok"], currentLang["btn_cancel"]}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == currentLang["btn_ok"] {
				app.Suspend(func() {
					sshConnect(server)
				})
			}
			app.SetRoot(centerWidget(createMainLayout(app, list)), true)
		})
	app.SetRoot(centerWidget(modal), true)
}

// openConfig opens the configuration file in the default system editor
func openConfig() {
	cmd := exec.Command("open", configFilePath)
	err := cmd.Run()
	if err != nil {
		log.Printf(currentLang["msg_config_open_error"], err)
	}
}

// deleteConnection shows a confirmation dialog and removes the selected connection
// Updates both the UI list and the saved configuration
func deleteConnection(app *tview.Application, list *tview.List, index int) {
	if index < 0 || index >= len(sshConnections) {
		return
	}

	server := sshConnections[index].Server
	modal := tview.NewModal().
		SetText(fmt.Sprintf(currentLang["dlg_delete"], server)).
		AddButtons([]string{currentLang["btn_ok"], currentLang["btn_cancel"]}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == currentLang["btn_ok"] {
				// Remove from slice
				sshConnections = append(sshConnections[:index], sshConnections[index+1:]...)
				// Remove from list
				list.RemoveItem(index)
				// Save changes
				saveConnections()
			}
			app.SetRoot(centerWidget(createMainLayout(app, list)), true)
		})
	app.SetRoot(centerWidget(modal), true)
}

// editConnection displays a form for editing an existing SSH connection
// Validates input and updates both the UI and saved configuration
func editConnection(app *tview.Application, connectionsList *tview.List, index int) {
	if index < 0 || index >= len(sshConnections) {
		return
	}

	connection := sshConnections[index]
	var form *tview.Form
	errorText := tview.NewTextView().SetText("")
	form = tview.NewForm().
		AddInputField(currentLang["form_server"], connection.Server, 30, nil, func(text string) {
			if text == "" {
				return
			}
			if text != connection.Server && isConnectionExists(text) {
				errorText.SetText(currentLang["msg_conn_exists"])
				return
			}
			errorText.SetText("")
		}).
		AddInputField(currentLang["form_port"], connection.Port, 5, nil, nil).
		AddInputField(currentLang["form_comment"], connection.Comment, 30, nil, nil).
		AddButton(currentLang["btn_save"], func() {
			server := form.GetFormItem(0).(*tview.InputField).GetText()
			port := form.GetFormItem(1).(*tview.InputField).GetText()
			comment := form.GetFormItem(2).(*tview.InputField).GetText()

			if server == "" {
				errorText.SetText(currentLang["msg_enter_server"])
				return
			}
			if comment == "" {
				errorText.SetText(currentLang["msg_enter_comment"])
				return
			}

			if server == connection.Server || !isConnectionExists(server) {
				sshConnections[index] = SSHConnection{Server: server, Port: port, Comment: comment}
				connectionsList.RemoveItem(index)
				connectionsList.InsertItem(index, fmt.Sprintf("%s - %s", server, comment), "", 0, func() {
					//showContextMenu(app, connectionsList, index)
					showMessage(app, connectionsList, server)
				})
				saveConnections()
				app.SetRoot(centerWidget(createMainLayout(app, connectionsList)), true)
			}
		}).
		AddButton(currentLang["btn_cancel"], func() {
			app.SetRoot(centerWidget(createMainLayout(app, connectionsList)), true)
		})

	formFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(form, 0, 1, true).
		AddItem(errorText, 1, 0, false)

	formFlex.SetBorder(true).SetTitle(currentLang["title_edit"]).SetTitleAlign(tview.AlignLeft)
	app.SetRoot(centerWidget(formFlex), true)
}

// Добавим функцию для отображения контекстного меню
func showContextMenu(app *tview.Application, connectionsList *tview.List, index int) {
	if index < 0 || index >= len(sshConnections) {
		return
	}

	server := sshConnections[index].Server
	contextMenu := tview.NewList().
		AddItem(currentLang["ctx_connect"], "", 0, func() {
			showMessage(app, connectionsList, server)
		}).
		AddItem(currentLang["ctx_edit"], "", 0, func() {
			editConnection(app, connectionsList, index)
		}).
		AddItem(currentLang["ctx_cancel"], "", 0, func() {
			app.SetRoot(centerWidget(createMainLayout(app, connectionsList)), true)
		})

	contextMenu.SetBorder(true).
		SetTitle(fmt.Sprintf(currentLang["ctx_actions"], server)).
		SetTitleAlign(tview.AlignLeft)

	// Create flex with fixed height for context menu
	menuFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(contextMenu, 6, 0, true). // Устанавливаем высоту 6 строк
		AddItem(nil, 0, 1, false)

	app.SetRoot(centerWidget(menuFlex), true)
}

// Add language switching function
func switchLanguage(app *tview.Application, connectionsList *tview.List) {
	modal := tview.NewModal().
		SetText("Select language / Выберите язык").
		AddButtons([]string{"English", "Русский"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "English" {
				currentLang = lang.EN
			} else {
				currentLang = lang.RU
			}

			// Update UI with new language
			connectionsList.SetTitle(currentLang["connections_title"])
			menuList.SetTitle(currentLang["menu_title"])
			helpText.SetText(currentLang["help_text"])

			// Update menu items
			menuList.Clear()
			menuList.AddItem(currentLang["menu_add"], "", 0, func() {
				addConnection(app, connectionsList)
			})
			menuList.AddItem(currentLang["menu_language"], "", 0, func() {
				switchLanguage(app, connectionsList)
			})
			menuList.AddItem(currentLang["menu_edit_config"], "", 0, func() {
				openConfig()
			})
			menuList.AddItem(currentLang["menu_exit"], "", 0, func() {
				app.Stop()
			})

			// Save config with new language
			saveConnections()

			app.SetRoot(centerWidget(createMainLayout(app, connectionsList)), true)
		})
	app.SetRoot(centerWidget(modal), true)
}

// main initializes and runs the SSH connection manager application
// Sets up the UI, loads configuration and handles user input
func main() {
	app := tview.NewApplication()

	// Load connections from file
	loadConnections()

	// Create connections list
	connectionsList := tview.NewList().ShowSecondaryText(false)
	connectionsList.SetTitle(currentLang["connections_title"]).SetBorder(true).SetTitleAlign(tview.AlignLeft)

	// Create menu
	menuList = tview.NewList().ShowSecondaryText(false)
	menuList.SetTitle(currentLang["menu_title"]).SetBorder(true).SetTitleAlign(tview.AlignLeft)

	// Add existing connections
	if len(sshConnections) == 0 {
		connectionsList.AddItem(currentLang["msg_no_connections"], "", 0, nil)
	} else {
		for _, conn := range sshConnections {
			server := conn.Server
			comment := conn.Comment
			connectionsList.AddItem(fmt.Sprintf("%s - %s", server, comment), "", 0, func() {
				// On Enter press show context menu
				//showContextMenu(app, connectionsList, index)
				showMessage(app, connectionsList, server)
			})
		}
	}

	// Add menu items
	menuList.AddItem(currentLang["menu_add"], "", 0, func() {
		addConnection(app, connectionsList)
	})
	menuList.AddItem(currentLang["menu_language"], "", 0, func() {
		switchLanguage(app, connectionsList)
	})
	menuList.AddItem(currentLang["menu_edit_config"], "", 0, func() {
		openConfig()
	})
	menuList.AddItem(currentLang["menu_exit"], "", 0, func() {
		app.Stop()
	})

	// Update help text
	helpText = tview.NewTextView().
		SetText(currentLang["help_text"]).
		SetTextAlign(tview.AlignLeft)

	// Set scrolling properties for connections list
	connectionsList.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		// Auto-scroll when reaching visible area boundaries
		offset, _ := connectionsList.GetOffset()
		if index < offset {
			connectionsList.SetOffset(index, 0)
		} else if index >= offset+connectionsHeight-2 { // -2 for borders
			connectionsList.SetOffset(index-connectionsHeight+3, 0)
		}
	})

	// Center main container
	flex := centerWidget(createMainLayout(app, connectionsList))

	// Update key handler in main()
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Get current active primitive
		primitive := app.GetFocus()

		// If modal window or input form is active, skip key handling
		if _, ok := primitive.(*tview.Modal); ok {
			return event
		}
		if _, ok := primitive.(*tview.InputField); ok {
			return event
		}

		switch event.Key() {
		case tcell.KeyCtrlC:
			app.Stop()
		case tcell.KeyDown:
			if app.GetFocus() == connectionsList {
				// If at end of connections list
				if connectionsList.GetCurrentItem() == connectionsList.GetItemCount()-1 {
					app.SetFocus(menuList)
					menuList.SetCurrentItem(0)
					return nil
				}
			} else if app.GetFocus() == menuList {
				// If at end of menu
				if menuList.GetCurrentItem() == menuList.GetItemCount()-1 {
					app.SetFocus(connectionsList)
					connectionsList.SetCurrentItem(0)
					return nil
				}
			}
		case tcell.KeyUp:
			if app.GetFocus() == connectionsList {
				// If at start of connections list
				if connectionsList.GetCurrentItem() == 0 {
					app.SetFocus(menuList)
					menuList.SetCurrentItem(menuList.GetItemCount() - 1)
					return nil
				}
			} else if app.GetFocus() == menuList {
				// If at start of menu
				if menuList.GetCurrentItem() == 0 {
					app.SetFocus(connectionsList)
					connectionsList.SetCurrentItem(connectionsList.GetItemCount() - 1)
					return nil
				}
			}
		case tcell.KeyCtrlE:
			if app.GetFocus() == connectionsList && connectionsList.GetItemCount() > 0 {
				currentIndex := connectionsList.GetCurrentItem()
				if currentIndex >= 0 && currentIndex < len(sshConnections) {
					modal := tview.NewModal().
						SetText(fmt.Sprintf(currentLang["dlg_edit"], sshConnections[currentIndex].Server)).
						AddButtons([]string{currentLang["btn_ok"], currentLang["btn_cancel"]}).
						SetDoneFunc(func(buttonIndex int, buttonLabel string) {
							if buttonLabel == currentLang["btn_ok"] {
								editConnection(app, connectionsList, currentIndex)
							} else {
								lists := tview.NewFlex().
									SetDirection(tview.FlexRow).
									AddItem(connectionsList, 0, 2, true).
									AddItem(menuList, 0, 1, false).
									AddItem(helpText, 0, 1, false)
								app.SetRoot(centerWidget(lists), true)
							}
						})
					app.SetRoot(centerWidget(modal), true)
				}
			}
			return nil
		case tcell.KeyCtrlN:
			// Show add new connection window
			modal := tview.NewModal().
				SetText(currentLang["dlg_add"]).
				AddButtons([]string{currentLang["btn_ok"], currentLang["btn_cancel"]}).
				SetDoneFunc(func(buttonIndex int, buttonLabel string) {
					app.SetRoot(centerWidget(createMainLayout(app, connectionsList)), true)
					if buttonLabel == currentLang["btn_ok"] {
						addConnection(app, connectionsList)
					}
				})
			app.SetRoot(centerWidget(modal), true)
			return nil
		case tcell.KeyDelete:
			if app.GetFocus() == connectionsList && connectionsList.GetItemCount() > 0 {
				deleteConnection(app, connectionsList, connectionsList.GetCurrentItem())
			}
			return nil
		}
		return event
	})

	// Launch application with flex container
	if err := app.SetRoot(flex, true).EnableMouse(true).Run(); err != nil {
		log.Fatalf(currentLang["msg_app_error"], err)
	}
}
