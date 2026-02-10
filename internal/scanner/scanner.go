package scanner

import (
	"sort"

	"github.com/manson/port-chaser/internal/models"
)

// Scanner defines the interface for port scanning operations.
// Implementations can scan all ports or query specific ports by number.
type Scanner interface {
	// Scan performs a complete port scan and returns all active ports
	Scan() ([]models.PortInfo, error)
	// ScanByPort returns detailed info for a specific port number
	ScanByPort(portNumber int) (*models.PortInfo, error)
}

// PortSorter defines methods for sorting port lists by different criteria.
// This allows flexible ordering of scan results based on user preferences.
type PortSorter interface {
	// SortByCommonPort sorts common ports (80, 443, etc.) first, then by port number
	SortByCommonPort(ports []models.PortInfo) []models.PortInfo
	// SortByPortNumber sorts ports strictly by port number in ascending order
	SortByPortNumber(ports []models.PortInfo) []models.PortInfo
}

// DefaultPortSorter implements PortSorter with standard sorting behavior.
// It prioritizes common development ports for easier navigation.
type DefaultPortSorter struct{}

// NewDefaultPortSorter creates a new DefaultPortSorter instance.
func NewDefaultPortSorter() *DefaultPortSorter {
	return &DefaultPortSorter{}
}

// SortByCommonPort sorts ports with common ports appearing first.
// Common ports (80, 443, 3000, 5000, 8000, 8080) are sorted to the top
// in ascending order, followed by remaining ports sorted by port number.
// This helps users find frequently-used development ports quickly.
func (s *DefaultPortSorter) SortByCommonPort(ports []models.PortInfo) []models.PortInfo {
	if len(ports) == 0 {
		return ports
	}

	// Create a copy to avoid modifying the original slice
	result := make([]models.PortInfo, len(ports))
	copy(result, ports)

	// Define which ports are considered "common" for development
	commonPortsSet := map[int]bool{
		80:   true,  // HTTP
		443:  true,  // HTTPS
		3000: true,  // Common dev server port
		5000: true,  // Common dev server port
		8000: true,  // Common dev server port
		8080: true,  // HTTP alternate
	}

	// Separate ports into common and other groups
	var commonPorts []models.PortInfo
	var otherPorts []models.PortInfo

	for _, port := range result {
		if commonPortsSet[port.PortNumber] {
			commonPorts = append(commonPorts, port)
		} else {
			otherPorts = append(otherPorts, port)
		}
	}

	// Sort each group by port number
	sort.Slice(commonPorts, func(i, j int) bool {
		return commonPorts[i].PortNumber < commonPorts[j].PortNumber
	})

	sort.Slice(otherPorts, func(i, j int) bool {
		return otherPorts[i].PortNumber < otherPorts[j].PortNumber
	})

	// Combine with common ports first
	result = append(commonPorts, otherPorts...)
	return result
}

// SortByPortNumber sorts ports strictly by port number in ascending order.
// This provides a simple, predictable ordering of all ports.
func (s *DefaultPortSorter) SortByPortNumber(ports []models.PortInfo) []models.PortInfo {
	if len(ports) == 0 {
		return ports
	}

	// Create a copy to avoid modifying the original slice
	result := make([]models.PortInfo, len(ports))
	copy(result, ports)

	sort.Slice(result, func(i, j int) bool {
		return result[i].PortNumber < result[j].PortNumber
	})

	return result
}

// ScanResult encapsulates the results of a port scan operation.
// It provides convenient methods for filtering and analyzing scan results.
type ScanResult struct {
	Ports     []models.PortInfo // Discovered ports
	Count     int               // Total number of ports found
	Error     error             // Any error that occurred during scanning
	Duration  int64             // Time taken for the scan (in milliseconds)
	HasDocker bool              // True if any Docker containers were found
}

// NewScanResult creates a new ScanResult from a ports slice and error.
// It automatically sets the count and detects presence of Docker containers.
func NewScanResult(ports []models.PortInfo, err error) *ScanResult {
	result := &ScanResult{
		Ports: ports,
		Count: len(ports),
		Error: err,
	}

	// Check if any ports belong to Docker containers
	for _, port := range ports {
		if port.IsDocker {
			result.HasDocker = true
			break
		}
	}

	return result
}

// FilteredPorts returns a subset of ports that match the given predicate.
// The predicate function receives each port and returns true to include it.
func (r *ScanResult) FilteredPorts(predicate func(models.PortInfo) bool) []models.PortInfo {
	var filtered []models.PortInfo
	for _, port := range r.Ports {
		if predicate(port) {
			filtered = append(filtered, port)
		}
	}
	return filtered
}

// CommonPorts returns only ports that are considered common (80, 443, etc.).
func (r *ScanResult) CommonPorts() []models.PortInfo {
	return r.FilteredPorts(func(p models.PortInfo) bool {
		return p.IsCommonPort()
	})
}

// DockerPorts returns only ports belonging to Docker containers.
func (r *ScanResult) DockerPorts() []models.PortInfo {
	return r.FilteredPorts(func(p models.PortInfo) bool {
		return p.IsDocker
	})
}

// RecommendedPorts returns ports that are frequently killed (KillCount >= 3).
// These might be processes the user often wants to terminate.
func (r *ScanResult) RecommendedPorts() []models.PortInfo {
	return r.FilteredPorts(func(p models.PortInfo) bool {
		return p.IsRecommended()
	})
}

// SystemPorts returns only system ports (port numbers 0-1023).
// These typically require elevated permissions and should be treated carefully.
func (r *ScanResult) SystemPorts() []models.PortInfo {
	return r.FilteredPorts(func(p models.PortInfo) bool {
		return p.IsSystemPort()
	})
}
