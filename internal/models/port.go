package models

import "time"

// PortInfo contains detailed information about a listening port and its associated process.
// This is the primary data structure for displaying ports in the TUI.
type PortInfo struct {
	// PortNumber is the TCP/UDP port number being listened on
	PortNumber int `json:"port_number"`
	// ProcessName is the name of the process using this port
	ProcessName string `json:"process_name"`
	// PID is the process ID of the process using this port
	PID int `json:"pid"`
	// User is the username of the process owner
	User string `json:"user"`
	// Command is the full command line that launched the process
	Command string `json:"command"`
	// IsDocker is true if this port belongs to a Docker container
	IsDocker bool `json:"is_docker"`
	// ContainerID is the Docker container ID (empty if not a Docker container)
	ContainerID string `json:"container_id"`
	// ContainerName is the Docker container name (empty if not a Docker container)
	ContainerName string `json:"container_name"`
	// ImageName is the Docker image name (empty if not a Docker container)
	ImageName string `json:"image_name"`
	// IsSystem is true if this is a system process that should be treated carefully
	IsSystem bool `json:"is_system"`
	// KillCount is how many times this port has been killed (tracked in history)
	KillCount int `json:"kill_count"`
	// LastKilled is when this port was last killed (zero if never)
	LastKilled time.Time `json:"last_killed"`
}

// HistoryEntry represents a single entry in the kill history log.
// It records each time a process is terminated.
type HistoryEntry struct {
	// ID is the unique identifier for this history entry
	ID int64 `json:"id"`
	// PortNumber is the port that was killed
	PortNumber int `json:"port_number"`
	// ProcessName is the name of the process that was killed
	ProcessName string `json:"process_name"`
	// PID is the process ID that was killed
	PID int `json:"pid"`
	// Command is the command line of the killed process
	Command string `json:"command"`
	// KilledAt is when the process was terminated
	KilledAt time.Time `json:"killed_at"`
}

// DockerInfo contains Docker-specific metadata for a container port.
// This is extracted when detecting that a port belongs to a Docker container.
type DockerInfo struct {
	// ContainerID is the unique Docker container ID
	ContainerID string `json:"container_id"`
	// ContainerName is the human-readable container name
	ContainerName string `json:"container_name"`
	// ImageName is the name of the Docker image the container is running
	ImageName string `json:"image_name"`
}

// IsCommonPort returns true if this port is commonly used for development.
// Common ports include: 80 (HTTP), 443 (HTTPS), 3000, 5000, 8000, 8080 (common dev servers).
func (p *PortInfo) IsCommonPort() bool {
	commonPorts := map[int]bool{
		80: true, 443: true, 3000: true,
		5000: true, 8000: true, 8080: true,
	}
	return commonPorts[p.PortNumber]
}

// IsRecommended returns true if this port has been killed 3 or more times.
// Frequently killed ports might be candidates for the user's attention.
func (p *PortInfo) IsRecommended() bool {
	return p.KillCount >= 3
}

// ShouldDisplayWarning returns true if this port should display a warning before killing.
// System processes and low-PID processes (<100) are considered potentially dangerous.
func (p *PortInfo) ShouldDisplayWarning() bool {
	return p.IsSystem || p.PID < 100
}

// IsSystemPort returns true if this is a well-known system port (0-1023).
// These ports typically require elevated permissions and are used by system services.
func (p *PortInfo) IsSystemPort() bool {
	return p.PortNumber >= 0 && p.PortNumber <= 1023
}
