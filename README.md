# Port Chaser

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

**Port Chaser** is a Terminal UI (TUI) based port management tool for developers. It displays active ports on your local system in real-time, automatically detects Docker containers, and allows you to quickly terminate processes using keyboard shortcuts.

## Features

- Real-time port scanning (within 2 seconds)
- Vim-style keyboard navigation
- Automatic Docker container detection
- Smart recommendations for frequently terminated processes
- SQLite-based termination history tracking

## Installation

### Go install

```bash
go install github.com/manson/port-chaser@latest
```

### Homebrew (macOS)

```bash
brew tap manson/port-chaser
brew install port-chaser
```

### Build from source

```bash
git clone https://github.com/manson/port-chaser.git
cd port-chaser
go build -o port-chaser ./cmd/port-chaser
sudo mv port-chaser /usr/local/bin/
```

## Usage

```bash
port-chaser
```

### Keyboard Shortcuts

| Key | Description |
|-----|-------------|
| `↑`/`k`, `↓`/`j` | Move up/down |
| `gg`, `G` | Go to top/bottom |
| `Enter` | Kill process |
| `d` | Toggle Docker filter |
| `h` | View history |
| `?` | Help |
| `r` | Refresh |
| `q` | Quit |

## Requirements

- Go 1.21+
- macOS, Linux, or Windows
- Docker (optional, for container detection)

## License

MIT License - see the [LICENSE](LICENSE) file for details.
