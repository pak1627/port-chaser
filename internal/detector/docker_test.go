// Package detector에 대한 테스트입니다.
package detector

import (
	"testing"

	"github.com/manson/port-chaser/internal/models"
)

// TestMockDetector_Detect는 모의 감지기 테스트입니다.
func TestMockDetector_Detect(t *testing.T) {
	detector := NewMockDetector()

	// 테스트용 Docker 정보 설정
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

	// 테스트
	ports := []models.PortInfo{
		{PortNumber: 3000, PID: 1001},
		{PortNumber: 8080, PID: 1002},
		{PortNumber: 5432, PID: 1003},
	}

	result, err := detector.Detect(ports)
	if err != nil {
		t.Fatalf("Detect() error = %v", err)
	}

	// 3000과 5432는 Docker여야 함
	dockerCount := 0
	for _, port := range result {
		if port.IsDocker {
			dockerCount++
			if port.ContainerID == "" {
				t.Error("Docker 포트의 ContainerID가 비어있습니다")
			}
		}
	}

	if dockerCount != 2 {
		t.Errorf("Docker 포트 수 = %d, want 2", dockerCount)
	}

	// 8080은 Docker가 아니어야 함
	var port8080 *models.PortInfo
	for i := range result {
		if result[i].PortNumber == 8080 {
			port8080 = &result[i]
			break
		}
	}

	if port8080 == nil {
		t.Fatal("포트 8080을 찾을 수 없습니다")
	}

	if port8080.IsDocker {
		t.Error("포트 8080은 Docker가 아니어야 합니다")
	}
}

// TestMockDetector_DetectEmpty는 빈 포트 목록 테스트입니다.
func TestMockDetector_DetectEmpty(t *testing.T) {
	detector := NewMockDetector()

	result, err := detector.Detect([]models.PortInfo{})
	if err != nil {
		t.Fatalf("Detect() error = %v", err)
	}

	if len(result) != 0 {
		t.Errorf("결과 길이 = %d, want 0", len(result))
	}
}

// TestMockDetector_IsAvailable은 가용성 확인 테스트입니다.
func TestMockDetector_IsAvailable(t *testing.T) {
	detector := NewMockDetector()

	// 기본적으로 사용 가능
	if !detector.IsAvailable() {
		t.Error("MockDetector는 기본적으로 사용 가능해야 합니다")
	}

	// 사용 불가능으로 설정
	detector.SetAvailable(false)

	if detector.IsAvailable() {
		t.Error("SetAvailable(false) 후에 사용 불가능해야 합니다")
	}
}

// TestEnrichPortInfo는 포트 정보 강화 테스트입니다.
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
		t.Error("강화된 포트는 Docker여야 합니다")
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
