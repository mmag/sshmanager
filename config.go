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
	Connections []SSHConnection `json:"connections"`
	Language    string          `json:"language"`
	// CheckInterval is the period of host availability checks in seconds, 0 disables them
	CheckInterval int `json:"check_interval"`
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
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf(currentLang["msg_read_error"], err)
		}
		return
	}

	if err := json.Unmarshal(data, &config); err != nil {
		log.Printf(currentLang["msg_parse_error"], err)
		return
	}

	if config.Language == "ru" {
		currentLang = lang.RU
	}
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

// findConnection returns the index of the connection with the given server address, or -1
func findConnection(server string) int {
	for i, conn := range config.Connections {
		if conn.Server == server {
			return i
		}
	}
	return -1
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
