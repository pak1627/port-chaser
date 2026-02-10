# Port Chaser

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Test Coverage](https://img.shields.io/badge/Coverage-83.3%25-brightgreen)](https://github.com/manson/port-chaser)

**Port Chaser** is a Terminal UI (TUI) based port management tool for developers. It displays active ports on your local system in real-time, automatically detects Docker containers, and allows you to quickly terminate processes using keyboard shortcuts.

## Key Features

- **Real-time Port Scanning**: Display system port usage within 2 seconds
- **Intuitive TUI**: Vim-style keyboard navigation support
- **Docker Integration**: Automatic detection and visual indication of container ports
- **Smart Recommendations**: Auto-display frequently terminated processes
- **Search & Filtering**: Quick port finding with real-time search
- **Safe Termination**: Gradual SIGTERM → SIGKILL process termination
- **History Tracking**: SQLite-based termination history management

## Getting Started

### Installation

#### Method 1: Go install (Simplest)

Requires Go 1.21 or higher.

```bash
go install github.com/manson/port-chaser@latest
```

After installation, the binary will be created in `~/go/bin` or `$GOPATH/bin`. Make sure this path is included in your PATH.

```bash
# Add to PATH (if needed)
export PATH=$PATH:$(go env GOPATH)/bin
```

#### Method 2: Homebrew (Recommended for macOS)

```bash
brew tap manson/port-chaser
brew install port-chaser
```

#### Method 3: Build from Source

```bash
git clone https://github.com/manson/port-chaser.git
cd port-chaser
go build -o port-chaser ./cmd/port-chaser
sudo mv port-chaser /usr/local/bin/
```

### System Requirements

- Go 1.21 or higher
- macOS, Linux, or Windows
- Docker (optional, for container detection)

## Usage

### Basic Execution

```bash
port-chaser
```

### Command Line Options

```bash
# Version information
port-chaser --version

# Help
port-chaser --help
```

## TUI Interface

### Main Screen

```
┌─────────────────────────────────────────────────────────────┐
│  Port Chaser                                             12 │
├─────────────────────────────────────────────────────────────┤
│  PORT   PROCESS         PID    USER    DOCKER    STATUS     │
│  3000   node            12345  dev     [D]my   [!] 5x      │
│  8080   python          23456  dev     -        -          │
│  5432   postgres        34567  pg      [D]db   1x         │
│  9000   custom-app      45678  dev     -        [!] 3x      │
├─────────────────────────────────────────────────────────────┤
│ ↑/k:navigate | Enter:kill | /:search | d:Docker filter | q:quit │
└─────────────────────────────────────────────────────────────┘
```

### Keyboard Shortcuts

#### Navigation

| Key | Description |
|-----|-------------|
| `↑` / `k` | Move up |
| `↓` / `j` | Move down |
| `gg` | Go to top |
| `G` | Go to bottom |

#### Actions

| Key | Description |
|-----|-------------|
| `Enter` | Terminate process |
| `/` | Search mode |
| `d` | Toggle Docker-only filter |
| `h` | View history |
| `?` | Help |
| `r` | Refresh |
| `q` / `Ctrl+C` | Quit |

#### Search Mode

| Key | Description |
|-----|-------------|
| `ESC` / `n` | Cancel search |
| `Enter` | Complete search input |

### Marker Legend

| Marker | Meaning |
|--------|---------|
| `[D]` | Running in Docker container |
| `[!]` | Process terminated 3+ times in last 30 days |
| `⚠` | System critical process (terminate with caution) |

## Feature Details

### Docker Container Detection

Port Chaser automatically detects ports running in Docker containers. Detected container information is displayed as follows:

```
3000  node  12345  dev  [D]my-app(node:16)  [!] 5x
```

- `[D]` marker: Running in Docker container
- Displays container name and image information

### Smart Recommendation System

Processes terminated 3 or more times in the last 30 days are marked with `[!]` and displayed at the top for priority access. This helps quickly identify development servers that are frequently terminated.

### Safe Process Termination

Process termination proceeds in two stages:

1. **SIGTERM**: Send normal termination signal (wait 3 seconds)
2. **SIGKILL**: Force terminate if not terminated within 3 seconds

System critical processes (PID < 100) are protected to prevent accidental termination.

### History Tracking

All terminated processes are recorded in a SQLite database. Press `h` to view history and get suggested restart commands for previously terminated processes.

## Architecture

```
port-chaser/
├── cmd/port-chaser/      # Main entry point
├── internal/
│   ├── app/              # TUI application model
│   ├── detector/         # Docker detector
│   ├── models/           # Data models
│   ├── platform/         # Platform-specific process management
│   ├── process/          # Process termination functionality
│   ├── scanner/          # Port scanner
│   ├── storage/          # SQLite history storage
│   └── ui/               # Bubbletea UI components
└── go.mod
```

### Tech Stack

- **Bubbletea**: TUI framework
- **Lipgloss**: Styling
- **SQLite**: History storage (WAL mode)
- **Docker SDK**: Container detection (can be replaced with CLI)

## Development

### Running Tests

```bash
# All tests
go test ./...

# Check coverage
go test -cover ./...

# Test specific package
go test ./internal/scanner
```

### Building

```bash
# Local build
go build -o port-chaser ./cmd/port-chaser

# Cross compile
GOOS=linux GOARCH=amd64 go build -o port-chaser-linux ./cmd/port-chaser
GOOS=darwin GOARCH=amd64 go build -o port-chaser-mac ./cmd/port-chaser
GOOS=windows GOARCH=amd64 go build -o port-chaser.exe ./cmd/port-chaser
```

### Code Style

The project follows Go standard coding style:

- English code comments
- English user documentation
- Target 85%+ test coverage

## Contributing

Contributions are welcome! Please follow these steps:

1. Fork this repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'feat: add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Author

manson - [GitHub](https://github.com/manson)

## Acknowledgments

- [Bubbletea](https://github.com/charmbracelet/bubbletea) - Excellent TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Styling library
- All contributors

---

**Project Homepage**: https://github.com/manson/port-chaser
**Bug Reports**: https://github.com/manson/port-chaser/issues
