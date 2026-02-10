package app

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/manson/port-chaser/internal/models"
)

// ViewMode represents the different UI views/states of the application.
// Each mode determines which screen is displayed and how keyboard input is handled.
type ViewMode int

const (
	// ViewModeMain is the primary port list view where users can browse and select ports
	ViewModeMain ViewMode = iota
	// ViewModeConfirmKill is the confirmation dialog shown before terminating a process
	ViewModeConfirmKill
	// ViewModeHistory displays the history of previously killed processes
	ViewModeHistory
	// ViewModeHelp shows the help/usage information
	ViewModeHelp
)

// String returns the string representation of the ViewMode for logging/debugging
func (vm ViewMode) String() string {
	switch vm {
	case ViewModeMain:
		return "main"
	case ViewModeConfirmKill:
		return "confirm_kill"
	case ViewModeHistory:
		return "history"
	case ViewModeHelp:
		return "help"
	default:
		return "unknown"
	}
}

// Model is the main application state for the Bubbletea TUI.
// It holds all mutable state including port data, UI state, and configuration.
// This follows the Bubbletea Model pattern: Init() -> Update() -> View() cycle.
type Model struct {
	// Ports is the complete list of discovered ports from the last scan
	Ports []models.PortInfo
	// FilteredPorts is the subset of ports currently displayed (respecting filters like Docker-only)
	FilteredPorts []models.PortInfo
	// SelectedIndex is the index of the currently selected port in FilteredPorts
	SelectedIndex int
	// ViewMode determines which screen is currently displayed
	ViewMode ViewMode
	// ShowDockerOnly when true filters to show only Docker container ports
	ShowDockerOnly bool
	// History contains records of previously killed processes
	History []models.HistoryEntry
	// Loading indicates a port scan is currently in progress
	Loading bool
	// Err holds the last error encountered during scanning or killing
	Err error
	// StatusMessage is a temporary notification message shown to the user
	StatusMessage string
	// StatusMessageTimeout is when the status message should be cleared
	StatusMessageTimeout time.Time
	// Quit when true signals the application should exit
	Quit bool
	// LastScanTime tracks when the last port scan was completed
	LastScanTime time.Time
	// KillConfirmationPort holds the port info pending user confirmation to kill
	KillConfirmationPort *models.PortInfo
	// Width is the current terminal width in characters
	Width int
	// Height is the current terminal height in characters
	Height int
	// Scanner is the port scanning interface (dependency injection for testing)
	Scanner       Scanner
	// Killer is the process termination interface (dependency injection for testing)
	Killer        Killer
	// Storage is the persistence layer for kill history (optional, nil means no persistence)
	Storage       Storage
	// PreviousPorts maps port numbers to their info from the last scan (for change detection)
	PreviousPorts map[int]models.PortInfo
	// NewPorts contains ports that appeared since the last scan (for highlighting)
	NewPorts     map[int]bool
	// RemovedPorts contains ports that disappeared since the last scan (for highlighting)
	RemovedPorts map[int]bool
}

// Scanner defines the interface for port scanning operations.
// This allows mocking in tests and swapping scanner implementations.
type Scanner interface {
	// Scan performs a complete port scan and returns all active ports
	Scan() ([]models.PortInfo, error)
}

// Killer defines the interface for process termination operations.
// This allows mocking in tests and platform-specific implementations.
type Killer interface {
	// Kill attempts to terminate the process associated with the given port
	Kill(port models.PortInfo) error
}

// Storage defines the interface for persisting and retrieving kill history.
// This allows the app to work without storage (nil Storage) or with various backends.
type Storage interface {
	// RecordKill saves a history entry when a process is killed
	RecordKill(entry models.HistoryEntry) error
	// GetHistory retrieves recent kill history, limited to the specified count
	GetHistory(limit int) ([]models.HistoryEntry, error)
	// GetKillCount returns how many times a specific port has been killed within days
	GetKillCount(port int, days int) (int, error)
	// Close closes the storage connection and releases resources
	Close() error
}

// Init is called by Bubbletea when the application starts.
// It returns the initial commands to run: load history, start port scanning, start tick timer, and enter alt screen.
func (m Model) Init() tea.Cmd {
	batch := []tea.Cmd{
		m.scanPortsCmd(),
		tea.Tick(time.Second*3, func(t time.Time) tea.Msg {
			return TickMsg{Time: t}
		}),
		tea.EnterAltScreen,
	}

	// Add history loading command if storage is available
	if m.Storage != nil {
		batch = append(batch, m.loadHistoryCmd())
	}

	return tea.Batch(batch...)
}

// scanPortsCmd returns a command that performs a port scan and returns the result as a message.
// This runs asynchronously and will send a PortsScannedMsg when complete.
func (m Model) scanPortsCmd() tea.Cmd {
	return func() tea.Msg {
		ports, err := m.Scanner.Scan()
		if err != nil {
			return PortsScannedMsg{
				Ports:     nil,
				ScannedAt: time.Now(),
				Error:     err,
			}
		}
		return PortsScannedMsg{
			Ports:     ports,
			ScannedAt: time.Now(),
		}
	}
}

// loadHistoryCmd returns a command that loads history from storage.
// This runs asynchronously and will send a HistoryLoadedMsg when complete.
func (m Model) loadHistoryCmd() tea.Cmd {
	return func() tea.Msg {
		history, err := m.Storage.GetHistory(100)
		return HistoryLoadedMsg{
			History: history,
			Error:   err,
		}
	}
}

// Update is the core of the Bubbletea Elm Architecture.
// It receives messages and returns the updated model along with commands to execute.
// All state transitions happen here in response to messages (key presses, ticks, scan results, etc).
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Delegate keyboard input handling based on current view mode
		return m.handleKeyMsg(msg)

	case tea.WindowSizeMsg:
		// Update terminal dimensions when window is resized
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil

	case PortsScannedMsg:
		// Handle completed port scan results
		return m.handlePortsScanned(msg)

	case PortKilledMsg:
		// Handle process kill completion (success or failure)
		return m.handlePortKilled(msg)

	case StatusMsg:
		// Display a temporary status message
		m.StatusMessage = msg.Message
		m.StatusMessageTimeout = time.Now().Add(time.Second * 3)
		return m, nil

	case TickMsg:
		// Handle periodic tasks like clearing status messages and auto-refreshing ports
		return m.handleTick(msg)

	case ClearHighlightsMsg:
		// Clear new/removed port highlights after a delay
		m.NewPorts = make(map[int]bool)
		m.RemovedPorts = make(map[int]bool)
		return m, nil

	case HistoryLoadedMsg:
		// Handle loaded history from storage
		if msg.Error == nil {
			m.History = msg.History
		}
		// If loading fails, continue with empty history (don't crash)
		return m, nil

	default:
		return m, nil
	}
}

// View renders the current UI state as a string.
// Bubbletea calls this after each Update to display the application.
// The rendering is delegated based on the current ViewMode.
func (m Model) View() string {
	switch m.ViewMode {
	case ViewModeMain:
		return m.renderMainView()
	case ViewModeConfirmKill:
		return m.renderConfirmKillView()
	case ViewModeHistory:
		return m.renderHistoryView()
	case ViewModeHelp:
		return m.renderHelpView()
	default:
		return "Unknown view mode"
	}
}

// PortsScannedMsg is sent when a port scan completes.
// It contains the scan results, timestamp, and any error that occurred.
type PortsScannedMsg struct {
	Ports     []models.PortInfo
	ScannedAt time.Time
	Error     error
}

// PortKilledMsg is sent when a process kill operation completes.
// It contains the target port info, success status, and result message.
type PortKilledMsg struct {
	Port    models.PortInfo
	Success bool
	Message string
}

// StatusMsg is a temporary notification message to display to the user.
// Examples: "Killed node", "Kill failed: permission denied"
type StatusMsg struct {
	Message string
}

// ClearHighlightsMsg is sent to remove new/removed port highlighting after a delay.
type ClearHighlightsMsg struct{}

// TickMsg is sent periodically to trigger auto-refresh and cleanup tasks.
type TickMsg struct {
	Time time.Time
}

// HistoryLoadedMsg is sent when history is loaded from storage.
// It contains the loaded history entries and any error that occurred.
type HistoryLoadedMsg struct {
	History []models.HistoryEntry
	Error   error
}

// handleKeyMsg routes keyboard input to the appropriate handler based on current view mode.
// Each view mode has its own key bindings and behavior.
func (m Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.ViewMode {
	case ViewModeMain:
		return m.handleMainKeyMsg(msg)
	case ViewModeConfirmKill:
		return m.handleConfirmKeyMsg(msg)
	case ViewModeHistory:
		return m.handleHistoryKeyMsg(msg)
	case ViewModeHelp:
		return m.handleHelpKeyMsg(msg)
	default:
		return m, nil
	}
}

// handlePortsScanned processes port scan results and detects port changes.
// It compares the current scan with the previous scan to highlight new/removed ports.
func (m Model) handlePortsScanned(msg PortsScannedMsg) (tea.Model, tea.Cmd) {
	if msg.Error != nil {
		m.Err = msg.Error
		m.Loading = false
		return m, nil
	}

	// Convert ports slice to map for efficient comparison
	currentPorts := make(map[int]models.PortInfo)
	for _, port := range msg.Ports {
		currentPorts[port.PortNumber] = port
	}

	// Detect new ports (present now but not in previous scan)
	newPorts := make(map[int]bool)
	removedPorts := make(map[int]bool)

	for portNum := range currentPorts {
		if _, exists := m.PreviousPorts[portNum]; !exists {
			newPorts[portNum] = true
		}
	}

	// Detect removed ports (present before but not in current scan)
	for portNum := range m.PreviousPorts {
		if _, exists := currentPorts[portNum]; !exists {
			removedPorts[portNum] = true
		}
	}

	// Update model state with scan results
	m.Ports = msg.Ports
	m.FilteredPorts = msg.Ports
	m.LastScanTime = msg.ScannedAt
	m.PreviousPorts = currentPorts
	m.NewPorts = newPorts
	m.RemovedPorts = removedPorts
	m.Loading = false

	// Schedule clearing of highlights after 3 seconds
	return m, tea.Tick(time.Second*3, func(t time.Time) tea.Msg {
		return ClearHighlightsMsg{}
	})
}

// handlePortKilled handles the result of a process kill operation.
// It shows a status message, records history if storage is available, and triggers a port rescan.
func (m Model) handlePortKilled(msg PortKilledMsg) (tea.Model, tea.Cmd) {
	m.ViewMode = ViewModeMain
	m.Loading = true

	// Record to history if kill succeeded and storage is available
	if msg.Success && m.Storage != nil {
		entry := models.HistoryEntry{
			PortNumber:  msg.Port.PortNumber,
			ProcessName:  msg.Port.ProcessName,
			PID:         msg.Port.PID,
			Command:     msg.Port.Command,
			KilledAt:    time.Now(),
		}

		// Record the kill (ignore errors to not disrupt UI)
		if err := m.Storage.RecordKill(entry); err != nil {
			// Log error but don't disrupt the flow
			// In production, you might want to show a warning
		} else {
			// Add to in-memory history for immediate display
			m.History = append([]models.HistoryEntry{entry}, m.History...)
			// Limit history to 100 entries in memory
			if len(m.History) > 100 {
				m.History = m.History[:100]
			}
		}
	}

	// Generate appropriate status message based on kill result
	var statusCmd tea.Cmd
	if msg.Success {
		statusCmd = func() tea.Msg {
			return StatusMsg{Message: "Killed " + msg.Port.ProcessName}
		}
	} else {
		statusCmd = func() tea.Msg {
			return StatusMsg{Message: "Kill failed: " + msg.Message}
		}
	}

	return m, tea.Batch(statusCmd, m.scanPortsCmd())
}

// handleTick processes periodic events like clearing expired status messages and auto-refreshing.
// The tick fires every 3 seconds to handle background tasks.
func (m Model) handleTick(msg TickMsg) (tea.Model, tea.Cmd) {
	// Clear status message if timeout has expired
	if !m.StatusMessageTimeout.IsZero() && time.Now().After(m.StatusMessageTimeout) {
		m.StatusMessage = ""
		m.StatusMessageTimeout = time.Time{}
	}

	// Auto-refresh ports every 3 seconds
	if time.Since(m.LastScanTime) > 3*time.Second {
		return m, m.scanPortsCmd()
	}

	return m, nil
}

// renderMainView renders the primary port list interface.
// It displays the title, status message, loading state, or the list of active ports.
func (m Model) renderMainView() string {
	var sb strings.Builder

	sb.WriteString("Port Chaser - Port Manager\n\n")

	if m.StatusMessage != "" {
		sb.WriteString(m.StatusMessage + "\n\n")
	}

	if m.Loading {
		sb.WriteString("Scanning ports...\n")
		return sb.String()
	}

	if len(m.FilteredPorts) == 0 {
		sb.WriteString("No active ports found.\n")
	} else {
		sb.WriteString("Active Ports:\n\n")

		// Render each port with selection cursor and highlights
		for i, port := range m.FilteredPorts {
			prefix := "  "
			if i == m.SelectedIndex {
				prefix = "> "
			}

			// Add highlight indicators for new/removed ports
			highlight := ""
			if m.NewPorts[port.PortNumber] {
				highlight = "\033[32m[NEW]\033[0m "
			} else if m.RemovedPorts[port.PortNumber] {
				highlight = "\033[31m[GONE]\033[0m "
			}

			sb.WriteString(fmt.Sprintf("%s%s%d - %s (PID: %d)\n",
				prefix, highlight, port.PortNumber, port.ProcessName, port.PID))

			// Show additional info for Docker containers
			if port.IsDocker {
				sb.WriteString(fmt.Sprintf("    Docker: %s (%s)\n",
					port.ContainerName, port.ImageName))
			}
			// Warn about system processes
			if port.IsSystem {
				sb.WriteString("    [System Process]\n")
			}
			// Show kill count for frequently killed ports
			if port.KillCount > 0 {
				sb.WriteString(fmt.Sprintf("    Kill Count: %d\n", port.KillCount))
			}
		}
	}

	sb.WriteString("\nKeys: ↑/k=up, ↓/j=down, Enter=kill, d=Docker only, q=quit\n")

	return sb.String()
}

// renderConfirmKillView renders the confirmation dialog before killing a process.
// It displays detailed information about the selected process and asks for confirmation.
func (m Model) renderConfirmKillView() string {
	if !m.isValidSelection() {
		m.ViewMode = ViewModeMain
		return m.renderMainView()
	}

	port := m.FilteredPorts[m.SelectedIndex]

	var sb strings.Builder
	sb.WriteString("⚠️  Confirm Kill Process\n\n")
	fmt.Fprintf(&sb, "Are you sure you want to kill this process?\n\n")
	sb.WriteString(fmt.Sprintf("  Port: %d\n", port.PortNumber))
	sb.WriteString(fmt.Sprintf("  Process: %s\n", port.ProcessName))
	sb.WriteString(fmt.Sprintf("  PID: %d\n", port.PID))
	sb.WriteString(fmt.Sprintf("  Command: %s\n", port.Command))

	if port.IsSystem {
		sb.WriteString("\n  [System Process - Be Careful]\n")
	}

	sb.WriteString("\nPress 'y' to kill, 'n' or Esc to cancel")

	return sb.String()
}

// renderHistoryView displays the history of previously killed processes.
// It shows a list of all processes that have been terminated, with timestamps.
func (m Model) renderHistoryView() string {
	var sb strings.Builder

	sb.WriteString("Kill History\n\n")

	if len(m.History) == 0 {
		sb.WriteString("No history available.\n")
	} else {
		for i, entry := range m.History {
			// Format: entry number, port, process name, PID, timestamp
			timestamp := entry.KilledAt.Format("2006-01-02 15:04:05")
			sb.WriteString(fmt.Sprintf("%d. Port %d - %s (PID: %d)\n",
				i+1, entry.PortNumber, entry.ProcessName, entry.PID))
			sb.WriteString(fmt.Sprintf("   Command: %s\n", truncateString(entry.Command, 60)))
			sb.WriteString(fmt.Sprintf("   Killed: %s\n\n", timestamp))
		}
	}

	sb.WriteString("Press q, esc, or h to return")

	return sb.String()
}

// truncateString truncates a string to a maximum length, appending "..." if truncated.
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// renderHelpView displays keyboard shortcuts and usage information.
// It shows all available commands and markers for understanding the UI.
func (m Model) renderHelpView() string {
	var sb strings.Builder

	sb.WriteString("Keyboard Shortcuts & Help\n\n")

	sb.WriteString("Navigation:\n")
	sb.WriteString("  ↑/k        Move selection up\n")
	sb.WriteString("  ↓/j        Move selection down\n")
	sb.WriteString("  g          Jump to first item\n")
	sb.WriteString("  G          Jump to last item\n\n")

	sb.WriteString("Actions:\n")
	sb.WriteString("  Enter      Kill selected process\n")
	sb.WriteString("  d          Toggle Docker-only filter\n")
	sb.WriteString("  r/Ctrl+R   Refresh port list\n\n")

	sb.WriteString("Views:\n")
	sb.WriteString("  h          Show kill history\n")
	sb.WriteString("  ?          Show this help screen\n")
	sb.WriteString("  q/Esc      Quit or return to main view\n\n")

	sb.WriteString("Markers:\n")
	sb.WriteString("  [D]        Docker container port\n")
	sb.WriteString("  [!]        Frequently killed (recommended)\n")
	sb.WriteString("  [S]        System process (be careful)\n")
	sb.WriteString("  [NEW]      New port since last scan\n")
	sb.WriteString("  [GONE]     Port removed since last scan\n\n")

	sb.WriteString("Press q, esc, or ? to return")

	return sb.String()
}

// handleMainKeyMsg handles keyboard input when in the main port list view.
// Supports navigation, selection, filtering, and process kill initiation.
func (m Model) handleMainKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		// Quit the application
		m.Quit = true
		return m, tea.Quit

	case "up", "k":
		// Move selection up (vim-style k also supported)
		m.moveSelection(-1)
		return m, nil

	case "down", "j":
		// Move selection down (vim-style j also supported)
		m.moveSelection(1)
		return m, nil

	case "g":
		// Jump to first item (vim-style)
		m.SelectedIndex = 0
		return m, nil

	case "G":
		// Jump to last item (vim-style)
		m.SelectedIndex = len(m.FilteredPorts) - 1
		return m, nil

	case "enter":
		// Open kill confirmation dialog for selected port
		if m.isValidSelection() {
			m.KillConfirmationPort = &m.FilteredPorts[m.SelectedIndex]
			m.ViewMode = ViewModeConfirmKill
		}
		return m, nil

	case "d":
		// Toggle Docker-only filter
		m.ShowDockerOnly = !m.ShowDockerOnly
		m.applyFilters()
		return m, nil

	case "h":
		// Open history view
		m.ViewMode = ViewModeHistory
		return m, nil

	case "?":
		// Open help view
		m.ViewMode = ViewModeHelp
		return m, nil

	case "r", "ctrl+r":
		// Manual refresh of port list
		m.Loading = true
		return m, m.scanPortsCmd()
	}

	return m, nil
}

// handleConfirmKeyMsg handles keyboard input in the kill confirmation dialog.
// 'y' confirms the kill, 'n' or 'Esc' cancels.
func (m Model) handleConfirmKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		// User confirmed - execute the kill command
		return m, m.killPortCmd()

	case "n", "N", "esc":
		// User cancelled - return to main view
		m.ViewMode = ViewModeMain
		m.KillConfirmationPort = nil
		return m, nil
	}

	return m, nil
}

// handleHistoryKeyMsg handles keyboard input in the history view.
// Any of q, esc, or h returns to the main view.
func (m Model) handleHistoryKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc", "h":
		m.ViewMode = ViewModeMain
		return m, nil
	}
	return m, nil
}

// handleHelpKeyMsg handles keyboard input in the help view.
// Any of q, esc, or ? returns to the main view.
func (m Model) handleHelpKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc", "?":
		m.ViewMode = ViewModeMain
		return m, nil
	}
	return m, nil
}

// moveSelection changes the selected port index by the given delta.
// Positive delta moves down, negative delta moves up.
// The index is clamped to valid bounds to prevent out-of-range selection.
func (m *Model) moveSelection(delta int) {
	newIndex := m.SelectedIndex + delta

	if newIndex < 0 {
		newIndex = 0
	}
	if newIndex >= len(m.FilteredPorts) {
		newIndex = len(m.FilteredPorts) - 1
	}

	m.SelectedIndex = newIndex
}

// isValidSelection checks if the current selection index points to a valid port.
// Returns false if the index is out of range or the filtered list is empty.
func (m Model) isValidSelection() bool {
	return m.SelectedIndex >= 0 && m.SelectedIndex < len(m.FilteredPorts)
}

// applyFilters filters the Ports list into FilteredPorts based on active filters.
// Currently supports Docker-only filtering. Adjusts selection index if needed.
func (m *Model) applyFilters() {
	m.FilteredPorts = m.Ports

	// Apply Docker-only filter if enabled
	if m.ShowDockerOnly {
		var dockerPorts []models.PortInfo
		for _, port := range m.Ports {
			if port.IsDocker {
				dockerPorts = append(dockerPorts, port)
			}
		}
		m.FilteredPorts = dockerPorts
	}

	// Adjust selection index if filter reduced the list
	if len(m.FilteredPorts) > 0 {
		if m.SelectedIndex >= len(m.FilteredPorts) {
			m.SelectedIndex = len(m.FilteredPorts) - 1
		}
	} else {
		m.SelectedIndex = -1
	}
}

// killPortCmd returns a command that kills the currently selected port's process.
// The command runs asynchronously and sends a PortKilledMsg when complete.
func (m Model) killPortCmd() tea.Cmd {
	if !m.isValidSelection() {
		return nil
	}

	port := m.FilteredPorts[m.SelectedIndex]

	return func() tea.Msg {
		err := m.Killer.Kill(port)

		if err != nil {
			return PortKilledMsg{
				Port:    port,
				Success: false,
				Message: err.Error(),
			}
		}

		return PortKilledMsg{
			Port:    port,
			Success: true,
			Message: "Process killed successfully",
		}
	}
}

