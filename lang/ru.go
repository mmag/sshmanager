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
	"form_server":  "SSH сервер",
	"form_port":    "Порт",
	"form_comment": "Комментарий",
	"title_add":    "Добавить соединение",
	"title_edit":   "Редактировать соединение",

	// Messages
	"msg_no_connections": "Нет сохраненных соединений",
	"msg_enter_server":   "Введите адрес сервера",
	"msg_enter_comment":  "Введите комментарий",
	"msg_conn_exists":    "Такое соединение уже существует",
	"msg_connecting":     "Подключение к %s\n",
	"msg_conn_error":     "Ошибка подключения к %s: %v\n",

	// Dialog messages
	"dlg_connect": "Подключиться к %s?",
	"dlg_edit":    "Редактировать соединение %s?",
	"dlg_delete":  "Удалить соединение %s?",
	"dlg_add":     "Добавить новое соединение?",

	// Context menu
	"ctx_connect": "Подключить",
	"ctx_edit":    "Редактировать",
	"ctx_cancel":  "Отмена",
	"ctx_actions": "Действия для %s",

	// Help text
	"help_text": "Управление:\n↑↓ - Навигация по списку\nEnter - Подключиться\nCtrl+E - Редактировать соединение\nCtrl+N - Добавить соединение\nDel - Удалить соединение\nCtrl+C - Выход",

	// Error messages
	"msg_config_dir_error":  "Ошибка создания директории конфигурации: %v\n",
	"msg_save_error":        "Ошибка сохранения соединений: %v\n",
	"msg_write_error":       "Ошибка записи файла: %v\n",
	"msg_read_error":        "Ошибка чтения файла: %v\n",
	"msg_parse_error":       "Ошибка разбора файла: %v\n",
	"msg_config_open_error": "Ошибка открытия конфига: %v\n",
	"msg_app_error":         "Ошибка запуска приложения: %v\n",
}
