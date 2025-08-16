package lang

var EN = map[string]string{
	// Menu items
	"menu_title":        "Menu",
	"connections_title": "Connections",
	"menu_add":          "Add connection",
	"menu_language":     "Language",
	"menu_edit_config":  "Edit config",
	"menu_exit":         "Exit",

	// Buttons
	"btn_ok":     "OK",
	"btn_cancel": "Cancel",
	"btn_save":   "Save",

	// Forms
	"form_server":   "SSH server",
	"form_port":     "Port",
	"form_comment":  "Comment",
	"form_username": "Username",
	"title_add":     "Add connection",
	"title_edit":    "Edit connection",

	// Messages
	"msg_no_connections": "No saved connections",
	"msg_enter_server":   "Enter server address",
	"msg_enter_comment":  "Enter comment",
	"msg_conn_exists":    "Connection already exists",
	"msg_connecting":     "Connecting to %s\n",
	"msg_conn_error":     "Connection error to %s: %v\n",

	// Dialog messages
	"dlg_connect": "Connect to %s?",
	"dlg_edit":    "Edit connection %s?",
	"dlg_delete":  "Delete connection %s?",
	"dlg_add":     "Add new connection?",

	// Context menu
	"ctx_connect": "Connect",
	"ctx_edit":    "Edit",
	"ctx_cancel":  "Cancel",
	"ctx_actions": "Actions for %s",

	// Help text
	"help_text": " Controls:                    \n ↑↓ - Navigate list           Tab - Switch section\n Enter - Connect              Ctrl+E - Edit connection\n Ctrl+N - Add connection      Del - Delete connection\n Ctrl+R - Refresh window      Ctrl+C - Exit",

	// Error messages
	"msg_config_dir_error":  "Error creating config directory: %v\n",
	"msg_save_error":        "Error saving connections: %v\n",
	"msg_write_error":       "Error writing file: %v\n",
	"msg_read_error":        "Error reading file: %v\n",
	"msg_parse_error":       "Error parsing file: %v\n",
	"msg_config_open_error": "Error opening config: %v\n",
	"msg_app_error":         "Application error: %v\n",
	
	// Language code
	"language_code": "en",
}
