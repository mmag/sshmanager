# SSHman — TUI SSH Connection Manager

Terminal UI tool for managing SSH connections, written in Go using [tview](https://github.com/rivo/tview).

## Features

- Manage SSH connections with friendly names
- Group connections into tabs (e.g. Home, Work), created, renamed and deleted from the menu
- Support for custom ports and usernames
- Host availability indicators, refreshed by ping every 15 seconds (configurable)
- Full SSH login check for the selected host
- Terminal UI with keyboard navigation
- English and Russian interface languages
- Config file storage in JSON format
- Port validation (1–65535)
- Auto-scrolling connection list

## Installation

### Homebrew (macOS and Linux)

```bash
brew install mmag/tap/sshman
```

### Build from source

```bash
git clone https://github.com/mmag/sshmanager.git
cd sshmanager
go build -o sshman
# optionally: mv sshman ~/.local/bin
```

## Usage

Launch the application:

```bash
sshman
```

### Keyboard Shortcuts

- `↑`/`↓` - Navigate through lists
- `Tab` - Switch between sections
- `←`/`→`, `1`–`9` - Switch tabs (or click a tab)
- `Enter` - Connect to selected server
- `Ctrl+E` - Edit selected connection
- `Ctrl+N` - Add new connection
- `Del` - Delete selected connection
- `Ctrl+Enter` / `Ctrl+T` - Check SSH login for the selected server
- `Ctrl+R` - Refresh window and recheck host statuses
- `Ctrl+C` - Exit application

### Configuration

Config is stored at `~/sshman/sshman.json` in the following format:

```json
{
  "tabs": [
    {
      "name": "Home",
      "connections": [
        {
          "server": "hostnameOrIP",
          "comment": "Description",
          "port": "22",
          "username": "user"
        }
      ]
    },
    {
      "name": "Work",
      "connections": []
    }
  ],
  "language": "en",
  "check_interval": 15
}
```

The first tab can't be deleted. When another tab is deleted, its connections are either
deleted too or moved to the first tab. Configs from older versions with a top-level
`connections` list are loaded into the first tab.

`check_interval` sets how often hosts are pinged, in seconds; `0` limits checks to startup and `Ctrl+R`.
Ping doesn't touch sshd, so frequent checks don't trigger fail2ban and similar protection.
Checks are paused while an SSH session is open.

The SSH login check runs `ssh` in batch mode with your `~/.ssh/config`, so it also works
for host aliases and ProxyJump hosts, and reports whether key authentication succeeds.
Many terminals send `Ctrl+Enter` as a plain `Enter`; use `Ctrl+T` there.

## Requirements

- Go 1.24 or higher
- System SSH client available in PATH (`ssh`)

## License

MIT
