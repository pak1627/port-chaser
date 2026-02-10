package detector

import (
	"testing"

	"github.com/manson/port-chaser/internal/models"
)

func TestMockDetector_Detect(t *testing.T) {
	detector := NewMockDetector()

	detector.SetDockerInfo(3000, models.DockerInfo{
		ContainerID:   "abc123",
		ContainerName: "node-app",
		ImageName:     "node:18",
	})

	detector.SetDockerInfo(5432, models.DockerInfo{
		ContainerID:   "def456",
		ContainerName: "postgres",
		ImageName:     "postgres:15",
	})

	ports := []models.PortInfo{
		{PortNumber: 3000, PID: 1001},
		{PortNumber: 8080, PID: 1002},
		{PortNumber: 5432, PID: 1003},
	}

	result, err := detector.Detect(ports)
	if err != nil {
		t.Fatalf("Detect() error = %v", err)
	}

	dockerCount := 0
	for _, port := range result {
		if port.IsDocker {
			dockerCount++
			if port.ContainerID == "" {
				t.Error("Docker port ContainerID is empty")
			}
		}
	}

	if dockerCount != 2 {
		t.Errorf("Docker port count = %d, want 2", dockerCount)
	}

	var port8080 *models.PortInfo
	for i := range result {
		if result[i].PortNumber == 8080 {
			port8080 = &result[i]
			break
		}
	}

	if port8080 == nil {
		t.Fatal("port 8080 not found")
	}

	if port8080.IsDocker {
		t.Error("port 8080 should not be Docker")
	}
}

func TestMockDetector_DetectEmpty(t *testing.T) {
	detector := NewMockDetector()

	result, err := detector.Detect([]models.PortInfo{})
	if err != nil {
		t.Fatalf("Detect() error = %v", err)
	}

	if len(result) != 0 {
		t.Errorf("result length = %d, want 0", len(result))
	}
}

func TestMockDetector_IsAvailable(t *testing.T) {
	detector := NewMockDetector()

	if !detector.IsAvailable() {
		t.Error("MockDetector should be available by default")
	}

	detector.SetAvailable(false)

	if detector.IsAvailable() {
		t.Error("should not be available after SetAvailable(false)")
	}
}

func TestEnrichPortInfo(t *testing.T) {
	dockerInfo := models.DockerInfo{
		ContainerID:   "abc123",
		ContainerName: "test-container",
		ImageName:     "nginx:latest",
	}

	port := models.PortInfo{
		PortNumber: 80,
		PID:        100,
	}

	enriched := EnrichPortInfo(port, dockerInfo)

	if !enriched.IsDocker {
		t.Error("enriched port should be Docker")
	}

	if enriched.ContainerID != "abc123" {
		t.Errorf("ContainerID = %s, want abc123", enriched.ContainerID)
	}

	if enriched.ContainerName != "test-container" {
		t.Errorf("ContainerName = %s, want test-container", enriched.ContainerName)
	}

	if enriched.ImageName != "nginx:latest" {
		t.Errorf("ImageName = %s, want nginx:latest", enriched.ImageName)
	}
}
