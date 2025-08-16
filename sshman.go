/*
* SSH Connection Manager
 */
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"sshman/lang"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Update the config structure by adding a new type
type Config struct {
	Connections []SSHConnection `json:"connections"`
	Language    string          `json:"language"`
}

type SSHConnection struct {
	Server   string `json:"server"`
	Comment  string `json:"comment"`
	Port     string `json:"port"`
	Username string `json:"username,omitempty"`
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

// Add constants for dimensions
const (
	formWidth     = 100 // increase width for better readability
	formHeight    = 60  // decrease height for compactness
	contextHeight = 6   // Context menu height
)

// Add global variables
var (
	currentLang = lang.EN // Default language
)

// createMainLayout creates and returns the main application layout with the specified heights
// for connections list, menu, and help sections based on screen height
func createMainLayout(app *tview.Application, connectionsList *tview.List) *tview.Flex {
	// Calculate menu height (items count + border)
	menuHeight := menuList.GetItemCount() + 2

	// Calculate help height (text lines + border)
	helpHeight := 8 // 6 text lines + top and bottom borders
	// Calculate connections list height (number of connections + 1 + border)
	connectionsHeight := len(sshConnections) + 3 // +1 for extra row, +2 for borders

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
	
	// Build the target with username if provided
	target := connection.Server
	if connection.Username != "" {
		target = fmt.Sprintf("%s@%s", connection.Username, connection.Server)
	}

	cmd := exec.Command("sh", "-c", fmt.Sprintf("%s %s", sshCommand, target))
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Printf(currentLang["msg_connecting"], connection.Server)
	err := cmd.Run()
	if err != nil {
		log.Printf(currentLang["msg_conn_error"], connection.Server, err)
	}
}

// formatConnectionLine formats connection info with dots between address and description
func formatConnectionLine(conn SSHConnection) string {
	// Build server address with port and username if needed
	serverPart := conn.Server
	if conn.Port != "" && conn.Port != "22" {
		serverPart = fmt.Sprintf("%s:%s", conn.Server, conn.Port)
	}
	if conn.Username != "" {
		serverPart = fmt.Sprintf("%s@%s", conn.Username, serverPart)
	}
	
	// Calculate available width - experimentally determined to fit the list width
	totalWidth := formWidth - 4
	serverLen := len(serverPart)
	commentLen := len(conn.Comment)
	
	// If both parts fit with at least 3 dots, use dots
	if serverLen + commentLen + 3 <= totalWidth {
		dotsCount := totalWidth - serverLen - commentLen
		dots := strings.Repeat(".", dotsCount)
		return fmt.Sprintf(" %s%s%s ", serverPart, dots, conn.Comment)
	}
	
	// If too long, just use simple format
	return fmt.Sprintf("%s - %s", serverPart, conn.Comment)
}

// centerWidget centers the provided widget in the screen with dynamic dimensions
func centerWidget(app *tview.Application, widget tview.Primitive) *tview.Flex {
	// Use reasonable defaults for screen size
	// tview will handle actual centering based on current terminal size
	screenWidth, screenHeight := 120, 40
	
	// Calculate widget dimensions based on screen size
	widgetWidth := formWidth
	if screenWidth < formWidth {
		widgetWidth = screenWidth - 4 // Leave some margin
	}
	
	// Calculate total height needed for layout
	menuHeight := menuList.GetItemCount() + 2
	helpHeight := 8
	connectionsHeight := len(sshConnections) + 3
	totalHeight := menuHeight + helpHeight + connectionsHeight
	
	widgetHeight := totalHeight
	if screenHeight < totalHeight {
		widgetHeight = screenHeight - 4 // Leave some margin
	}
	
	widget.SetRect(0, 0, widgetWidth, widgetHeight)
	flex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().
			SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(widget, widgetHeight, 0, true).
			AddItem(nil, 0, 1, false), widgetWidth, 0, true).
		AddItem(nil, 0, 1, false)
	flex.SetBackgroundColor(tcell.ColorNavy)
	return flex
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
	errorText.SetTextColor(tcell.ColorYellow).SetBackgroundColor(tcell.ColorNavy)
	form = tview.NewForm()
	form.SetBackgroundColor(tcell.ColorNavy)
	form.SetFieldBackgroundColor(tcell.ColorDarkBlue)
	form.SetFieldTextColor(tcell.ColorWhite)
	form.SetLabelColor(tcell.ColorWhite)
	form.SetButtonBackgroundColor(tcell.ColorDarkRed)
	form.SetButtonTextColor(tcell.ColorWhite)
	
	form.
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
		AddInputField(currentLang["form_username"], "", 20, nil, nil).
		AddButton(currentLang["btn_save"], func() {
			server := form.GetFormItem(0).(*tview.InputField).GetText()
			port := form.GetFormItem(1).(*tview.InputField).GetText()
			comment := form.GetFormItem(2).(*tview.InputField).GetText()
			username := form.GetFormItem(3).(*tview.InputField).GetText()

			if server == "" {
				errorText.SetText(currentLang["msg_enter_server"])
				return
			}
			if comment == "" {
				errorText.SetText(currentLang["msg_enter_comment"])
				return
			}

			if !isConnectionExists(server) {
				connection := SSHConnection{Server: server, Port: port, Comment: comment, Username: username}
				sshConnections = append(sshConnections, connection)

				// Add new item
				newIndex := connectionsList.GetItemCount()
				mainText, _ := connectionsList.GetItemText(0)
				if newIndex == 1 && mainText == currentLang["msg_no_connections"] {
					connectionsList.Clear()
					newIndex = 0
				}
				// Format with dots
				newConn := SSHConnection{Server: server, Port: port, Comment: comment, Username: username}
				displayText := formatConnectionLine(newConn)
				connectionsList.AddItem(displayText, "", 0, func() {
					showMessage(app, connectionsList, server)
				})
				saveConnections()

				// Return to main screen
				app.SetRoot(centerWidget(app, createMainLayout(app, connectionsList)), true)
				connectionsList.SetCurrentItem(newIndex)
			}
		}).
		AddButton(currentLang["btn_cancel"], func() {
			app.SetRoot(centerWidget(app, createMainLayout(app, connectionsList)), true)
		})

	// Create flex for form with error text
	formFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(form, 0, 1, true).
		AddItem(errorText, 1, 0, false)

	formFlex.SetBorder(true).
		SetTitle(currentLang["title_add"]).
		SetTitleAlign(tview.AlignLeft).
		SetBackgroundColor(tcell.ColorNavy).
		SetBorderColor(tcell.ColorWhite).
		SetTitleColor(tcell.ColorWhite)

	// Set form as active widget
	app.SetRoot(centerWidget(app, formFlex), true)
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
	err = os.WriteFile(configFilePath, data, 0644)
	if err != nil {
		log.Printf(currentLang["msg_write_error"], err)
	}
}

// loadConnections reads and parses the SSH connections from the configuration file
// Silently handles the case when the config file doesn't exist
func loadConnections() {
	data, err := os.ReadFile(configFilePath)
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
	modal := tview.NewModal()
	modal.SetBackgroundColor(tcell.ColorNavy)
	modal.SetTextColor(tcell.ColorWhite)
	modal.SetButtonBackgroundColor(tcell.ColorDarkRed)
	modal.SetButtonTextColor(tcell.ColorWhite)
	modal.
		SetText(fmt.Sprintf(currentLang["dlg_connect"], server)).
		AddButtons([]string{currentLang["btn_ok"], currentLang["btn_cancel"]}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == currentLang["btn_ok"] {
				app.Suspend(func() {
					sshConnect(server)
				})
			}
			app.SetRoot(centerWidget(app, createMainLayout(app, list)), true)
		})
	app.SetRoot(centerWidget(app, modal), true)
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
	modal := tview.NewModal()
	modal.SetBackgroundColor(tcell.ColorNavy)
	modal.SetTextColor(tcell.ColorWhite)
	modal.SetButtonBackgroundColor(tcell.ColorDarkRed)
	modal.SetButtonTextColor(tcell.ColorWhite)
	modal.
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
			app.SetRoot(centerWidget(app, createMainLayout(app, list)), true)
		})
	app.SetRoot(centerWidget(app, modal), true)
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
	errorText.SetTextColor(tcell.ColorYellow).SetBackgroundColor(tcell.ColorNavy)
	form = tview.NewForm()
	form.SetBackgroundColor(tcell.ColorNavy)
	form.SetFieldBackgroundColor(tcell.ColorDarkBlue)
	form.SetFieldTextColor(tcell.ColorWhite)
	form.SetLabelColor(tcell.ColorWhite)
	form.SetButtonBackgroundColor(tcell.ColorDarkRed)
	form.SetButtonTextColor(tcell.ColorWhite)
	
	form.
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
		AddInputField(currentLang["form_username"], connection.Username, 20, nil, nil).
		AddButton(currentLang["btn_save"], func() {
			server := form.GetFormItem(0).(*tview.InputField).GetText()
			port := form.GetFormItem(1).(*tview.InputField).GetText()
			comment := form.GetFormItem(2).(*tview.InputField).GetText()
			username := form.GetFormItem(3).(*tview.InputField).GetText()

			if server == "" {
				errorText.SetText(currentLang["msg_enter_server"])
				return
			}
			if comment == "" {
				errorText.SetText(currentLang["msg_enter_comment"])
				return
			}

			if server == connection.Server || !isConnectionExists(server) {
				updatedConn := SSHConnection{Server: server, Port: port, Comment: comment, Username: username}
				sshConnections[index] = updatedConn
				connectionsList.RemoveItem(index)
				// Format with dots
				displayText := formatConnectionLine(updatedConn)
				connectionsList.InsertItem(index, displayText, "", 0, func() {
					showMessage(app, connectionsList, server)
				})
				saveConnections()
				app.SetRoot(centerWidget(app, createMainLayout(app, connectionsList)), true)
			}
		}).
		AddButton(currentLang["btn_cancel"], func() {
			app.SetRoot(centerWidget(app, createMainLayout(app, connectionsList)), true)
		})

	formFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(form, 0, 1, true).
		AddItem(errorText, 1, 0, false)

	formFlex.SetBorder(true).
		SetTitle(currentLang["title_edit"]).
		SetTitleAlign(tview.AlignLeft).
		SetBackgroundColor(tcell.ColorNavy).
		SetBorderColor(tcell.ColorWhite).
		SetTitleColor(tcell.ColorWhite)
	app.SetRoot(centerWidget(app, formFlex), true)
}


// Add language switching function
func switchLanguage(app *tview.Application, connectionsList *tview.List) {
	modal := tview.NewModal()
	modal.SetBackgroundColor(tcell.ColorNavy)
	modal.SetTextColor(tcell.ColorWhite)
	modal.SetButtonBackgroundColor(tcell.ColorDarkRed)
	modal.SetButtonTextColor(tcell.ColorWhite)
	modal.
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
			menuList.AddItem(" "+currentLang["menu_add"], "", 0, func() {
				addConnection(app, connectionsList)
			})
			menuList.AddItem(" "+currentLang["menu_language"], "", 0, func() {
				switchLanguage(app, connectionsList)
			})
			menuList.AddItem(" "+currentLang["menu_edit_config"], "", 0, func() {
				openConfig()
			})
			menuList.AddItem(" "+currentLang["menu_exit"], "", 0, func() {
				app.Stop()
			})

			// Save config with new language
			saveConnections()

			app.SetRoot(centerWidget(app, createMainLayout(app, connectionsList)), true)
		})
	app.SetRoot(centerWidget(app, modal), true)
}

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

// main initializes and runs the SSH connection manager application
// Sets up the UI, loads configuration and handles user input
func main() {
	app := tview.NewApplication()
	
	// Apply Debian installer theme
	setupDebianTheme()

	// Load connections from file
	loadConnections()

	// Create connections list
	connectionsList := tview.NewList().ShowSecondaryText(false)
	connectionsList.SetTitle(currentLang["connections_title"]).SetBorder(true).SetTitleAlign(tview.AlignLeft)
	connectionsList.SetBackgroundColor(tcell.ColorNavy)
	connectionsList.SetMainTextColor(tcell.ColorWhite)
	connectionsList.SetSelectedTextColor(tcell.ColorWhite)
	connectionsList.SetSelectedBackgroundColor(tcell.ColorDarkRed)

	// Create menu
	menuList = tview.NewList().ShowSecondaryText(false)
	menuList.SetTitle(currentLang["menu_title"]).SetBorder(true).SetTitleAlign(tview.AlignLeft)
	menuList.SetBackgroundColor(tcell.ColorNavy)
	menuList.SetMainTextColor(tcell.ColorWhite)
	menuList.SetSelectedTextColor(tcell.ColorWhite)
	menuList.SetSelectedBackgroundColor(tcell.ColorDarkRed)

	// Add existing connections
	if len(sshConnections) == 0 {
		connectionsList.AddItem(currentLang["msg_no_connections"], "", 0, nil)
	} else {
		for i, conn := range sshConnections {
			index := i // Capture index to avoid closure issue
			displayText := formatConnectionLine(conn)
			connectionsList.AddItem(displayText, "", 0, func() {
				currentIndex := index
				showMessage(app, connectionsList, sshConnections[currentIndex].Server)
			})
		}
	}

	// Add menu items with left padding
	menuList.AddItem(" "+currentLang["menu_add"], "", 0, func() {
		addConnection(app, connectionsList)
	})
	menuList.AddItem(" "+currentLang["menu_language"], "", 0, func() {
		switchLanguage(app, connectionsList)
	})
	menuList.AddItem(" "+currentLang["menu_edit_config"], "", 0, func() {
		openConfig()
	})
	menuList.AddItem(" "+currentLang["menu_exit"], "", 0, func() {
		app.Stop()
	})

	// Update help text
	helpText = tview.NewTextView().
		SetText(currentLang["help_text"]).
		SetTextAlign(tview.AlignLeft)
	helpText.SetBackgroundColor(tcell.ColorNavy)
	helpText.SetTextColor(tcell.ColorWhite)
	helpText.SetBorderColor(tcell.ColorWhite)

	// Set scrolling properties for connections list
	connectionsList.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		// Auto-scroll when reaching visible area boundaries
		// Get current connections list height
		_, _, _, listHeight := connectionsList.GetInnerRect()
		visibleItems := listHeight
		
		offset, _ := connectionsList.GetOffset()
		if index < offset {
			connectionsList.SetOffset(index, 0)
		} else if index >= offset+visibleItems {
			connectionsList.SetOffset(index-visibleItems+1, 0)
		}
	})

	// Add focus change handlers to manage selection colors
	connectionsList.SetFocusFunc(func() {
		if connectionsList.GetItemCount() > 0 && connectionsList.GetCurrentItem() < 0 {
			connectionsList.SetCurrentItem(0)
		}
		// Active colors - red background
		connectionsList.SetSelectedTextColor(tcell.ColorWhite)
		connectionsList.SetSelectedBackgroundColor(tcell.ColorDarkRed)
		
		// Make menu inactive - same color as background
		menuList.SetSelectedTextColor(tcell.ColorWhite)
		menuList.SetSelectedBackgroundColor(tcell.ColorNavy)
	})

	menuList.SetFocusFunc(func() {
		if menuList.GetItemCount() > 0 && menuList.GetCurrentItem() < 0 {
			menuList.SetCurrentItem(0)
		}
		// Active colors - red background
		menuList.SetSelectedTextColor(tcell.ColorWhite)
		menuList.SetSelectedBackgroundColor(tcell.ColorDarkRed)
		
		// Make connections list inactive - same color as background
		connectionsList.SetSelectedTextColor(tcell.ColorWhite)
		connectionsList.SetSelectedBackgroundColor(tcell.ColorNavy)
	})

	// Center main container
	flex := centerWidget(app, createMainLayout(app, connectionsList))

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
		if _, ok := primitive.(*tview.Form); ok {
			return event
		}

		// Only handle Tab for main lists, not for forms
		if event.Key() == tcell.KeyTab {
			// Check if we're in the main interface (connectionsList or menuList have focus)
			if app.GetFocus() != connectionsList && app.GetFocus() != menuList {
				return event
			}
		}

		switch event.Key() {
		case tcell.KeyCtrlC:
			app.Stop()
		case tcell.KeyCtrlR:
			// Refresh/redraw window - recreate layout and center it
			currentFocus := app.GetFocus()
			app.SetRoot(centerWidget(app, createMainLayout(app, connectionsList)), true)
			// Restore focus to the previously focused element
			if currentFocus == connectionsList {
				app.SetFocus(connectionsList)
			} else if currentFocus == menuList {
				app.SetFocus(menuList)
			} else {
				app.SetFocus(connectionsList) // Default to connections list
			}
			return nil
		case tcell.KeyTab:
			// Switch between lists
			if app.GetFocus() == connectionsList {
				app.SetFocus(menuList)
			} else {
				app.SetFocus(connectionsList)
			}
			return nil
		case tcell.KeyDown:
			if app.GetFocus() == connectionsList {
				// Wrap around at the end
				if connectionsList.GetCurrentItem() == connectionsList.GetItemCount()-1 {
					connectionsList.SetCurrentItem(0)
					return nil
				}
			} else if app.GetFocus() == menuList {
				// Wrap around at the end
				if menuList.GetCurrentItem() == menuList.GetItemCount()-1 {
					menuList.SetCurrentItem(0)
					return nil
				}
			}
		case tcell.KeyUp:
			if app.GetFocus() == connectionsList {
				// Wrap around at the beginning
				if connectionsList.GetCurrentItem() == 0 {
					connectionsList.SetCurrentItem(connectionsList.GetItemCount() - 1)
					return nil
				}
			} else if app.GetFocus() == menuList {
				// Wrap around at the beginning
				if menuList.GetCurrentItem() == 0 {
					menuList.SetCurrentItem(menuList.GetItemCount() - 1)
					return nil
				}
			}
		case tcell.KeyCtrlE:
			if app.GetFocus() == connectionsList && connectionsList.GetItemCount() > 0 {
				currentIndex := connectionsList.GetCurrentItem()
				if currentIndex >= 0 && currentIndex < len(sshConnections) {
					modal := tview.NewModal()
	modal.SetBackgroundColor(tcell.ColorNavy)
	modal.SetTextColor(tcell.ColorWhite)
	modal.SetButtonBackgroundColor(tcell.ColorDarkRed)
	modal.SetButtonTextColor(tcell.ColorWhite)
	modal.
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
								app.SetRoot(centerWidget(app, lists), true)
							}
						})
					app.SetRoot(centerWidget(app, modal), true)
				}
			}
			return nil
		case tcell.KeyCtrlN:
			// Show add new connection window
			modal := tview.NewModal()
	modal.SetBackgroundColor(tcell.ColorNavy)
	modal.SetTextColor(tcell.ColorWhite)
	modal.SetButtonBackgroundColor(tcell.ColorDarkRed)
	modal.SetButtonTextColor(tcell.ColorWhite)
	modal.
				SetText(currentLang["dlg_add"]).
				AddButtons([]string{currentLang["btn_ok"], currentLang["btn_cancel"]}).
				SetDoneFunc(func(buttonIndex int, buttonLabel string) {
					app.SetRoot(centerWidget(app, createMainLayout(app, connectionsList)), true)
					if buttonLabel == currentLang["btn_ok"] {
						addConnection(app, connectionsList)
					}
				})
			app.SetRoot(centerWidget(app, modal), true)
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
