package lang

var RU = map[string]string{
	// Menu items
	"menu_title":        "Меню",
	"connections_title": "Соединения",
	"menu_add":          "Добавить соединение",
	"menu_language":     "Язык",
	"menu_edit_config":  "Редактировать конфиг",
	"menu_exit":         "Выход",

	// Buttons
	"btn_ok":     "OK",
	"btn_cancel": "Отмена",
	"btn_save":   "Сохранить",

	// Forms
	"form_server":   "SSH сервер",
	"form_port":     "Порт",
	"form_comment":  "Комментарий",
	"form_username": "Имя пользователя",
	"title_add":     "Добавить соединение",
	"title_edit":    "Редактировать соединение",

	// Messages
	"msg_no_connections": "Нет сохраненных соединений",
	"msg_enter_server":   "Введите адрес сервера",
	"msg_enter_comment":  "Введите комментарий",
	"msg_conn_exists":    "Такое соединение уже существует",
	"msg_invalid_port":   "Порт должен быть числом от 1 до 65535",
	"msg_connecting":     "Подключение к %s\n",
	"msg_conn_error":     "Ошибка подключения к %s: %v\n",

	// Dialog messages
	"dlg_connect": "Подключиться к %s?",
	"dlg_edit":    "Редактировать соединение %s?",
	"dlg_delete":  "Удалить соединение %s?",
	"dlg_add":     "Добавить новое соединение?",

	// SSH login check
	"dlg_ssh_checking": "Проверка входа по SSH: %s…",
	"ssh_ok":           "%s\n\nВход по SSH выполнен за %.1f с",
	"ssh_denied":       "%s\n\nSSH-сервер отвечает, но войти не удалось:\n%s",
	"ssh_failed":       "%s\n\nНе удалось подключиться:\n%s",
	"ssh_timeout":      "%s\n\nНет ответа за %d с",

	// Help text
	"help_text": " Управление:\n ↑↓ - Навигация по списку              Tab - Переключить раздел\n Enter - Подключиться                  Ctrl+Enter, Ctrl+T - Проверить вход по SSH\n Ctrl+N - Добавить соединение          Ctrl+E - Редактировать соединение\n Del - Удалить соединение              Ctrl+R - Проверить доступность\n Ctrl+C - Выход",

	// Error messages
	"msg_config_dir_error":  "Ошибка создания директории конфигурации: %v\n",
	"msg_save_error":        "Ошибка сохранения соединений: %v\n",
	"msg_write_error":       "Ошибка записи файла: %v\n",
	"msg_read_error":        "Ошибка чтения файла: %v\n",
	"msg_parse_error":       "Ошибка разбора файла: %v\n",
	"msg_config_open_error": "Ошибка открытия конфига: %v\n",
	"msg_app_error":         "Ошибка запуска приложения: %v\n",

	// Language code
	"language_code": "ru",
}
