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
	"msg_invalid_port":   "Port must be a number between 1 and 65535",
	"msg_connecting":     "Connecting to %s\n",
	"msg_conn_error":     "Connection error to %s: %v\n",

	// Dialog messages
	"dlg_connect": "Connect to %s?",
	"dlg_edit":    "Edit connection %s?",
	"dlg_delete":  "Delete connection %s?",
	"dlg_add":     "Add new connection?",

	// SSH login check
	"dlg_ssh_checking": "Checking SSH login: %s…",
	"ssh_ok":           "%s\n\nSSH login succeeded in %.1f s",
	"ssh_denied":       "%s\n\nSSH server responds, but login failed:\n%s",
	"ssh_failed":       "%s\n\nConnection failed:\n%s",
	"ssh_timeout":      "%s\n\nNo response in %d s",

	// Help text
	"help_text": " Controls:\n ↑↓ - Navigate list           Tab - Switch section\n Enter - Connect              Ctrl+Enter, Ctrl+T - Check SSH login\n Ctrl+N - Add connection      Ctrl+E - Edit connection\n Del - Delete connection      Ctrl+R - Recheck hosts\n Ctrl+C - Exit",

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
