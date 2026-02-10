package main

import (
	"context"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/manson/port-chaser/internal/app"
	"github.com/manson/port-chaser/internal/models"
	"github.com/manson/port-chaser/internal/process"
	"github.com/manson/port-chaser/internal/scanner"
	"github.com/manson/port-chaser/internal/storage"
)

const (
	// appName is the human-readable name of the application
	appName = "Port Chaser"
	// version is the current semantic version
	version = "0.1.0"
)

// main is the application entry point.
// It handles command-line flags and starts the Bubbletea TUI program.
func main() {
	// Handle command-line flags for version and help
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-v", "--version", "version":
			fmt.Printf("%s v%s\n", appName, version)
			os.Exit(0)
		case "-h", "--help", "help":
			printHelp()
			os.Exit(0)
		}
	}

	// Initialize the application model with dependencies
	model := initializeModel()

	// Ensure storage is closed when the app exits
	if model.Storage != nil {
		defer func() {
			if err := model.Storage.Close(); err != nil {
				// Log error but don't affect exit code
				fmt.Fprintf(os.Stderr, "warning: failed to close storage: %v\n", err)
			}
		}()
	}

	// Create and run the Bubbletea program
	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),       // Use alternate screen buffer (fullscreen TUI)
		tea.WithMouseCellMotion(), // Enable mouse support for future features
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
}

// initializeModel creates the initial application state with all dependencies wired up.
// This is where dependency injection happens for testability.
func initializeModel() app.Model {
	killer := process.NewProcessKiller()
	commonPortScanner := scanner.NewCommonPortScanner()

	// Initialize storage (SQLite backend)
	// If storage initialization fails, the app will work without persistence
	var sto app.Storage
	sqliteStorage, err := storage.NewSQLite(storage.DefaultConfig())
	if err == nil {
		sto = sqliteStorage
	}
	// If err != nil, sto remains nil and the app works without persistence

	return app.Model{
		Ports:          []models.PortInfo{},
		FilteredPorts:  []models.PortInfo{},
		SelectedIndex:  -1,
		ViewMode:       app.ViewModeMain,
		ShowDockerOnly: false,
		History:        []models.HistoryEntry{},
		Loading:        true,
		Width:          80,
		Height:         24,
		Scanner:        commonPortScanner,
		Killer:         &killerAdapter{killer: killer},
		Storage:        sto,
		PreviousPorts:  make(map[int]models.PortInfo),
		NewPorts:       make(map[int]bool),
		RemovedPorts:   make(map[int]bool),
	}
}

// printHelp displays usage information and keyboard shortcuts.
func printHelp() {
	help := `Port Chaser - Terminal UI Port Management Tool

Usage:
  port-chaser [options]

Options:
  -v, --version     Show version
  -h, --help        Show help

TUI Key Bindings:
  Arrow/k/j         Navigate up/down
  gg, G             Jump to top/bottom
  Enter             Kill process
  /                 Search
  d                 Toggle Docker filter
  h                 Show history
  ?                 Show help
  r                 Refresh
  q, Ctrl+C         Quit

Project: https://github.com/manson/port-chaser
`
	fmt.Println(help)
}

// killerAdapter adapts the process.Killer interface to the app.Killer interface.
// The app.Killer interface is simpler (takes PortInfo), while process.Killer
// takes a PID and context separately. This adapter bridges the two.
type killerAdapter struct {
	killer *process.ProcessKiller
}

// Kill attempts to terminate the process associated with the given port.
// It uses a 5-second timeout and returns an error if termination fails.
func (a *killerAdapter) Kill(port models.PortInfo) error {
	// Create a context with 5-second timeout (5 * 1e9 nanoseconds)
	ctx, cancel := context.WithTimeout(context.Background(), 5*1000000000)
	defer cancel()

	result, err := a.killer.Kill(ctx, port.PID, &port)
	if err != nil {
		return err
	}
	if !result.Success {
		return fmt.Errorf("kill failed: %s", result.Message)
	}
	return nil
}
