package main

import (
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"

	"sshman/lang"
)

type Config struct {
	Tabs []Tab `json:"tabs"`
	// Connections is the list from configs without tabs, moved to the first tab on load
	Connections []SSHConnection `json:"connections,omitempty"`
	Language    string          `json:"language"`
	// CheckInterval is the period of host availability checks in seconds, 0 disables them
	CheckInterval int `json:"check_interval"`
}

type Tab struct {
	Name        string          `json:"name"`
	Connections []SSHConnection `json:"connections,omitempty"`
}

type SSHConnection struct {
	Server   string `json:"server"`
	Comment  string `json:"comment"`
	Port     string `json:"port"`
	Username string `json:"username,omitempty"`
}

var (
	configDir      = filepath.Join(os.Getenv("HOME"), "sshman")
	configFilePath = filepath.Join(configDir, "sshman.json")
	config         = Config{CheckInterval: 15}
	currentLang    = lang.EN
)

// loadConfig reads and parses the configuration file
// Silently handles the case when the config file doesn't exist
func loadConfig() {
	data, err := os.ReadFile(configFilePath)
	switch {
	case err == nil:
		if err := json.Unmarshal(data, &config); err != nil {
			log.Printf(currentLang["msg_parse_error"], err)
		}
	case !os.IsNotExist(err):
		log.Printf(currentLang["msg_read_error"], err)
	}

	if config.Language == "ru" {
		currentLang = lang.RU
	}

	// The first tab always exists and receives connections from configs without tabs
	if len(config.Tabs) == 0 {
		config.Tabs = []Tab{{Name: currentLang["tab_default"]}}
	}
	config.Tabs[0].Connections = append(config.Connections, config.Tabs[0].Connections...)
	config.Connections = nil
}

// saveConfig writes the configuration to disk
// Creates the config directory if it doesn't exist
func saveConfig() {
	if err := os.MkdirAll(configDir, 0755); err != nil {
		log.Printf(currentLang["msg_config_dir_error"], err)
		return
	}

	config.Language = currentLang["language_code"]

	data, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		log.Printf(currentLang["msg_save_error"], err)
		return
	}

	data = append(data, '\n')
	if err := os.WriteFile(configFilePath, data, 0644); err != nil {
		log.Printf(currentLang["msg_write_error"], err)
	}
}

// openConfig opens the configuration file in the default system editor
func openConfig() {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", configFilePath)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", configFilePath)
	default:
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "xdg-open"
		}
		cmd = exec.Command(editor, configFilePath)
	}
	if err := cmd.Run(); err != nil {
		log.Printf(currentLang["msg_config_open_error"], err)
	}
}

// currentConnections returns the connections of the current tab
func currentConnections() []SSHConnection {
	return config.Tabs[currentTab].Connections
}

// allConnections returns the connections of all tabs
func allConnections() []SSHConnection {
	var connections []SSHConnection
	for _, tab := range config.Tabs {
		connections = append(connections, tab.Connections...)
	}
	return connections
}

// connectionExists reports whether any tab has a connection with the given server address
func connectionExists(server string) bool {
	for _, tab := range config.Tabs {
		for _, conn := range tab.Connections {
			if conn.Server == server {
				return true
			}
		}
	}
	return false
}

func tabExists(name string) bool {
	for _, tab := range config.Tabs {
		if tab.Name == name {
			return true
		}
	}
	return false
}

// isValidPort checks that port string is empty (use default) or a number 1–65535
func isValidPort(port string) bool {
	if port == "" {
		return true
	}
	n, err := strconv.Atoi(port)
	if err != nil {
		return false
	}
	return n >= 1 && n <= 65535
}
