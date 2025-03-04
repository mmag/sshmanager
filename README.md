# SSHman - SSH Connection Manager

A terminal-based SSH connection manager written in Go using [tview](https://github.com/rivo/tview).

## Features

- Manage SSH connections with friendly names
- Support for custom ports
- Terminal UI with keyboard navigation
- Config file storage in JSON format
- Connection validation
- Auto-scrolling connection list

## Installation

```bash
go install github.com/yourusername/sshman@latest
```

## Usage

Launch the application:

```bash
sshmanager
```

### Keyboard Shortcuts

- `↑`/`↓` - Navigate through lists
- `Enter` - Connect to selected server
- `Ctrl+E` - Edit selected connection
- `Ctrl+N` - Add new connection
- `Del` - Delete selected connection
- `Ctrl+C` - Exit application

### Configuration

Config file is stored at `~/sshman/sshman.json` in the following format:

```json
[
    {
        "server": "user@hostnameOrIP",
        "comment": "Description",
        "port": "22"
    },
    "language": "en"
]
```

## Building from Source

```bash
git clone https://github.com/mmag/sshmanager.git
cd sshmanager
go build
```

## Requirements

- Go 1.16 or higher
- SSH client installed in system

## License

MIT