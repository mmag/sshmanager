# SSHman — TUI SSH Connection Manager

Terminal UI tool for managing SSH connections, written in Go using [tview](https://github.com/rivo/tview).

## Features

- Manage SSH connections with friendly names
- Support for custom ports and usernames
- Online/offline status indicators for hosts
- Terminal UI with keyboard navigation
- English and Russian interface languages
- Config file storage in JSON format
- Port validation (1–65535)
- Auto-scrolling connection list

## Installation

Build locally and run the `sshman` binary:

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
- `Enter` - Connect to selected server
- `Ctrl+E` - Edit selected connection
- `Ctrl+N` - Add new connection
- `Del` - Delete selected connection
- `Ctrl+R` - Refresh window and recheck host statuses
- `Ctrl+C` - Exit application

### Configuration

Config is stored at `~/sshman/sshman.json` in the following format:

```json
{
  "connections": [
    {
      "server": "hostnameOrIP",
      "comment": "Description",
      "port": "22",
      "username": "user"
    }
  ],
  "language": "en"
}
```

## Building from Source

Same as Installation above, or:

```bash
go build -o sshman
```

## Requirements

- Go 1.24 or higher
- System SSH client available in PATH (`ssh`)

## License

MIT
