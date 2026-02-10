// Package detector는 Docker 컨테이너 감지 기능을 제공합니다.
package detector

import (
	"github.com/manson/port-chaser/internal/models"
)

// Detector는 Docker 컨테이너 감지 인터페이스입니다.
type Detector interface {
	// Detect는 포트 목록에서 Docker 컨테이너를 감지하고 정보를 강화합니다.
	Detect(ports []models.PortInfo) ([]models.PortInfo, error)

	// IsAvailable은 Docker 감지기가 사용 가능한지 확인합니다.
	IsAvailable() bool
}

// MockDetector는 테스트용 모의 Docker 감지기입니다.
type MockDetector struct {
	available bool
	dockerMap map[int]models.DockerInfo
}

// NewMockDetector는 새로운 MockDetector를 생성합니다.
func NewMockDetector() *MockDetector {
	return &MockDetector{
		available: true,
		dockerMap: make(map[int]models.DockerInfo),
	}
}

// SetAvailable은 가용성을 설정합니다.
func (m *MockDetector) SetAvailable(available bool) {
	m.available = available
}

// SetDockerInfo는 포트 번호에 대한 Docker 정보를 설정합니다.
func (m *MockDetector) SetDockerInfo(portNumber int, info models.DockerInfo) {
	m.dockerMap[portNumber] = info
}

// Detect는 포트 목록에서 Docker 컨테이너를 감지합니다.
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

// IsAvailable은 MockDetector가 사용 가능한지 확인합니다.
func (m *MockDetector) IsAvailable() bool {
	return m.available
}

// EnrichPortInfo는 포트 정보에 Docker 정보를 추가합니다.
func EnrichPortInfo(port models.PortInfo, dockerInfo models.DockerInfo) models.PortInfo {
	port.IsDocker = true
	port.ContainerID = dockerInfo.ContainerID
	port.ContainerName = dockerInfo.ContainerName
	port.ImageName = dockerInfo.ImageName
	return port
}
