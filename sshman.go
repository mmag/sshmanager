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

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type SSHConnection struct {
	Server  string `json:"server"`
	Comment string `json:"comment"`
	Port    string `json:"port"`
}

var sshConnections []SSHConnection
var configDir = filepath.Join(os.Getenv("HOME"), "sshman")
var configFilePath = filepath.Join(configDir, "sshman.json")
var menuList *tview.List
var helpText *tview.TextView

// Add constants for form dimensions
const (
	formWidth  = 100 // increase width for better readability
	formHeight = 60  // decrease height for compactness
)

// createMainLayout creates and returns the main application layout with the specified heights
// for connections list, menu, and help sections based on screen height
func createMainLayout(app *tview.Application, connectionsList *tview.List) *tview.Flex {
	// Устанавливаем фиксированную высоту терминала
	const screenHeight = 80

	// Вычисляем высоту меню (количество пунктов + рамка)
	menuHeight := menuList.GetItemCount() + 2

	// Вычисляем высоту справки (количество строк + рамка)
	helpHeight := 8 // 6 строк текста + рамка сверху и снизу

	// Вычисляем высоту списка соединений
	connectionsHeight := screenHeight/2 - menuHeight - helpHeight

	// Создаем вертикальный flex для списков и справки
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

	log.Printf("Подключение к %s\n", connection.Server)
	err := cmd.Run()
	if err != nil {
		log.Printf("Ошибка подключения к %s: %v\n", connection.Server, err)
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
		AddInputField("SSH сервер", "", 30, nil, func(text string) {
			if text == "" {
				return
			}
			if isConnectionExists(text) {
				errorText.SetText("Connection already exists")
				return
			}
			errorText.SetText("")
		}).
		AddInputField("Порт", "", 5, nil, nil).
		AddInputField("Комментарий", "", 30, nil, nil).
		AddButton("Сохранить", func() {
			server := form.GetFormItem(0).(*tview.InputField).GetText()
			port := form.GetFormItem(1).(*tview.InputField).GetText()
			comment := form.GetFormItem(2).(*tview.InputField).GetText()

			if server == "" {
				errorText.SetText("Введите адрес сервера")
				return
			}
			if comment == "" {
				errorText.SetText("Введите комментарий")
				return
			}

			if !isConnectionExists(server) {
				connection := SSHConnection{Server: server, Port: port, Comment: comment}
				sshConnections = append(sshConnections, connection)

				// Добавляем новый пункт
				newIndex := connectionsList.GetItemCount()
				mainText, _ := connectionsList.GetItemText(0)
				if newIndex == 1 && mainText == "Нет сохраненных соединений" {
					connectionsList.Clear()
					newIndex = 0
				}
				connectionsList.AddItem(fmt.Sprintf("%s - %s", server, comment), "", 0, func() {
					showMessage(app, connectionsList, server)
				})
				saveConnections()

				// Возвращаемся к основному экрану
				app.SetRoot(centerWidget(createMainLayout(app, connectionsList)), true)
				connectionsList.SetCurrentItem(newIndex)
			}
		}).
		AddButton("Отмена", func() {
			app.SetRoot(centerWidget(createMainLayout(app, connectionsList)), true)
		})

	// Создаем flex для формы с текстом ошибки
	formFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(form, 0, 1, true).
		AddItem(errorText, 1, 0, false)

	formFlex.SetBorder(true).
		SetTitle("Добавить соединение").
		SetTitleAlign(tview.AlignLeft)

	// Устанавливаем форму как активный виджет
	app.SetRoot(centerWidget(formFlex), true)
	app.SetFocus(form)
}

// saveConnections writes the current connections list to the configuration file
// Creates the config directory if it doesn't exist
func saveConnections() {
	// Ensure the config directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		log.Printf("Ошибка создания директории конфигурации: %v\n", err)
		return
	}

	// Используем MarshalIndent вместо Marshal для форматированного вывода
	data, err := json.MarshalIndent(sshConnections, "", "    ")
	if err != nil {
		log.Printf("Ошибка сохранения соединений: %v\n", err)
		return
	}

	// Добавляем перевод строки в конец файла
	data = append(data, '\n')

	err = ioutil.WriteFile(configFilePath, data, 0644)
	if err != nil {
		log.Printf("Ошибка записи файла: %v\n", err)
	}
}

// loadConnections reads and parses the SSH connections from the configuration file
// Silently handles the case when the config file doesn't exist
func loadConnections() {
	data, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("Ошибка чтения файла: %v\n", err)
		}
		return
	}
	err = json.Unmarshal(data, &sshConnections)
	if err != nil {
		log.Printf("Ошибка разбора файла: %v\n", err)
	}
}

// showMessage displays a confirmation dialog before establishing an SSH connection
// Suspends the application while the SSH connection is active
func showMessage(app *tview.Application, list *tview.List, server string) {
	modal := tview.NewModal().
		SetText(fmt.Sprintf("Подключиться к %s?", server)).
		AddButtons([]string{"OK", "Отмена"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "OK" {
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
		log.Printf("Ошибка открытия конфига: %v\n", err)
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
		SetText(fmt.Sprintf("Удалить соединение %s?", server)).
		AddButtons([]string{"OK", "Отмена"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "OK" {
				// Удаляем из слайса
				sshConnections = append(sshConnections[:index], sshConnections[index+1:]...)
				// Удаляем из списка
				list.RemoveItem(index)
				// Сохраняем изменения
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
		AddInputField("SSH сервер", connection.Server, 30, nil, func(text string) {
			if text == "" {
				return
			}
			if text != connection.Server && isConnectionExists(text) {
				errorText.SetText("Connection already exists")
				return
			}
			errorText.SetText("")
		}).
		AddInputField("Порт", connection.Port, 5, nil, nil).
		AddInputField("Комментарий", connection.Comment, 30, nil, nil).
		AddButton("Сохранить", func() {
			server := form.GetFormItem(0).(*tview.InputField).GetText()
			port := form.GetFormItem(1).(*tview.InputField).GetText()
			comment := form.GetFormItem(2).(*tview.InputField).GetText()

			if server == "" {
				errorText.SetText("Введите адрес сервера")
				return
			}
			if comment == "" {
				errorText.SetText("Введите комментарий")
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
		AddButton("Отмена", func() {
			app.SetRoot(centerWidget(createMainLayout(app, connectionsList)), true)
		})

	formFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(form, 0, 1, true).
		AddItem(errorText, 1, 0, false)

	formFlex.SetBorder(true).SetTitle("Редактировать соединение").SetTitleAlign(tview.AlignLeft)
	app.SetRoot(centerWidget(formFlex), true)
}

// Добавим функцию для отображения контекстного меню
func showContextMenu(app *tview.Application, connectionsList *tview.List, index int) {
	if index < 0 || index >= len(sshConnections) {
		return
	}

	server := sshConnections[index].Server
	contextMenu := tview.NewList().
		AddItem("Подключить", "", 0, func() {
			showMessage(app, connectionsList, server)
		}).
		AddItem("Редактировать", "", 0, func() {
			editConnection(app, connectionsList, index)
		}).
		AddItem("Отмена", "", 0, func() {
			app.SetRoot(centerWidget(createMainLayout(app, connectionsList)), true)
		})

	contextMenu.SetBorder(true).
		SetTitle(fmt.Sprintf("Действия для %s", server)).
		SetTitleAlign(tview.AlignLeft)

	// Создаем flex с фиксированной высотой для контекстного меню
	menuFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(contextMenu, 6, 0, true). // Устанавливаем высоту 6 строк
		AddItem(nil, 0, 1, false)

	app.SetRoot(centerWidget(menuFlex), true)
}

// main initializes and runs the SSH connection manager application
// Sets up the UI, loads configuration and handles user input
func main() {
	app := tview.NewApplication()

	// Load connections from file
	loadConnections()

	// Create connections list
	connectionsList := tview.NewList().ShowSecondaryText(false)
	connectionsList.SetTitle("Соединения").SetBorder(true).SetTitleAlign(tview.AlignLeft)

	// Create menu
	menuList = tview.NewList().ShowSecondaryText(false)
	menuList.SetTitle("Меню").SetBorder(true).SetTitleAlign(tview.AlignLeft)

	// Add existing connections
	if len(sshConnections) == 0 {
		connectionsList.AddItem("No saved connections", "", 0, nil)
	} else {
		for _, conn := range sshConnections {
			server := conn.Server
			comment := conn.Comment
			connectionsList.AddItem(fmt.Sprintf("%s - %s", server, comment), "", 0, func() {
				// При нажатии Enter показываем контекстное меню
				//showContextMenu(app, connectionsList, index)
				showMessage(app, connectionsList, server)
			})
		}
	}

	// Add menu items
	menuList.AddItem("Add connection", "", 0, func() {
		addConnection(app, connectionsList)
	})
	menuList.AddItem("Редактировать конфиг", "", 0, func() {
		openConfig()
	})
	menuList.AddItem("Выход", "", 0, func() {
		app.Stop()
	})

	// Update help text
	helpText = tview.NewTextView().
		SetText("Controls:\n" +
			"↑↓ - Navigate list\n" +
			"Enter - Connect\n" +
			"Ctrl+E - Edit connection\n" +
			"Ctrl+N - Add connection\n" +
			"Del - Delete connection\n" +
			"Ctrl+C - Exit").
		SetTextAlign(tview.AlignLeft)
	helpText.SetBorder(true).SetTitle("Помощь").SetTitleAlign(tview.AlignLeft)

	// Устанавливаем фиксированную высоту терминала
	const screenHeight = 80

	// Вычисляем высоту меню (количество пунктов + рамка)
	menuHeight := menuList.GetItemCount() + 2

	// Вычисляем высоту справки (количество строк + рамка)
	helpHeight := 8 // 6 строк текста + рамка сверху и снизу

	// Вычисляем высоту списка соединений
	connectionsHeight := screenHeight/2 - menuHeight - helpHeight

	// Создаем вертикальный flex для списков и справки
	tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(connectionsList, connectionsHeight, 0, true).
		AddItem(menuList, menuHeight, 0, false).
		AddItem(helpText, helpHeight, 0, false)

	// Устанавливаем свойства прокрутки для списка соединений
	connectionsList.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		// Автоматическая прокрутка при достижении границ видимой области
		offset, _ := connectionsList.GetOffset()
		if index < offset {
			connectionsList.SetOffset(index, 0)
		} else if index >= offset+connectionsHeight-2 { // -2 для учета границ
			connectionsList.SetOffset(index-connectionsHeight+3, 0)
		}
	})

	// Центрируем общий контейнер
	flex := centerWidget(createMainLayout(app, connectionsList))

	// Обновляем обработчик клавиш в main()
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Получаем текущий активный примитив
		primitive := app.GetFocus()

		// Если активно модальное окно или форма ввода, пропускаем обработку клавиш
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
				// Если достигли конца списка соединений
				if connectionsList.GetCurrentItem() == connectionsList.GetItemCount()-1 {
					app.SetFocus(menuList)
					menuList.SetCurrentItem(0)
					return nil
				}
			} else if app.GetFocus() == menuList {
				// Если достигли конца меню
				if menuList.GetCurrentItem() == menuList.GetItemCount()-1 {
					app.SetFocus(connectionsList)
					connectionsList.SetCurrentItem(0)
					return nil
				}
			}
		case tcell.KeyUp:
			if app.GetFocus() == connectionsList {
				// Если в начале списка соединений
				if connectionsList.GetCurrentItem() == 0 {
					app.SetFocus(menuList)
					menuList.SetCurrentItem(menuList.GetItemCount() - 1)
					return nil
				}
			} else if app.GetFocus() == menuList {
				// Если в начале меню
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
						SetText(fmt.Sprintf("Редактировать соединение %s?", sshConnections[currentIndex].Server)).
						AddButtons([]string{"OK", "Отмена"}).
						SetDoneFunc(func(buttonIndex int, buttonLabel string) {
							if buttonLabel == "OK" {
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
			// Показываем окно добавления нового соединения
			modal := tview.NewModal().
				SetText("Добавить новое соединение?").
				AddButtons([]string{"OK", "Отмена"}).
				SetDoneFunc(func(buttonIndex int, buttonLabel string) {
					app.SetRoot(centerWidget(createMainLayout(app, connectionsList)), true)
					if buttonLabel == "OK" {
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

	// Запуск приложения с flex контейнером
	if err := app.SetRoot(flex, true).EnableMouse(true).Run(); err != nil {
		log.Fatalf("Ошибка запуска приложения: %v\n", err)
	}
}
