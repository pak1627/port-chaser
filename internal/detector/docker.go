package detector

import (
	"github.com/manson/port-chaser/internal/models"
)

// Detector defines the interface for detecting Docker containers associated with ports.
// Implementations can query the Docker daemon to find container information.
type Detector interface {
	// Detect enriches port information with Docker details when available
	Detect(ports []models.PortInfo) ([]models.PortInfo, error)
	// IsAvailable returns true if Docker is available and can be queried
	IsAvailable() bool
}

// MockDetector is a test implementation of Detector that allows setting predefined responses.
// This is useful for testing without requiring an actual Docker daemon to be running.
type MockDetector struct {
	available bool                      // whether the mock reports Docker as available
	dockerMap map[int]models.DockerInfo // mapping of port numbers to Docker info
}

// NewMockDetector creates a new MockDetector with default settings.
// By default it reports as available with an empty Docker mapping.
func NewMockDetector() *MockDetector {
	return &MockDetector{
		available: true,
		dockerMap: make(map[int]models.DockerInfo),
	}
}

// SetAvailable sets whether the mock detector reports Docker as available.
func (m *MockDetector) SetAvailable(available bool) {
	m.available = available
}

// SetDockerInfo associates Docker information with a specific port number.
// When Detect is called, ports with this number will be enriched with the provided info.
func (m *MockDetector) SetDockerInfo(portNumber int, info models.DockerInfo) {
	m.dockerMap[portNumber] = info
}

// Detect enriches port information based on the pre-configured dockerMap.
// Ports that have a corresponding entry in dockerMap will be marked as Docker ports
// with the associated container ID, name, and image information.
func (m *MockDetector) Detect(ports []models.PortInfo) ([]models.PortInfo, error) {
	if !m.available {
		return ports, nil
	}

	result := make([]models.PortInfo, len(ports))
	for i, port := range ports {
		if dockerInfo, ok := m.dockerMap[port.PortNumber]; ok {
			result[i] = EnrichPortInfo(port, dockerInfo)
		} else {
			result[i] = port
		}
	}

	return result, nil
}

// IsAvailable returns whether this mock detector reports as available.
func (m *MockDetector) IsAvailable() bool {
	return m.available
}

// EnrichPortInfo creates a new PortInfo with Docker details merged in.
// It sets the IsDocker flag and populates container-related fields.
func EnrichPortInfo(port models.PortInfo, dockerInfo models.DockerInfo) models.PortInfo {
	port.IsDocker = true
	port.ContainerID = dockerInfo.ContainerID
	port.ContainerName = dockerInfo.ContainerName
	port.ImageName = dockerInfo.ImageName
	return port
}
