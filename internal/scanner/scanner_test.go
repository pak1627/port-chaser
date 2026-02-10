package scanner

import (
	"testing"

	"github.com/manson/port-chaser/internal/models"
)

func TestDefaultPortSorter_SortByCommonPort(t *testing.T) {
	sorter := NewDefaultPortSorter()

	ports := []models.PortInfo{
		{PortNumber: 9999, ProcessName: "custom", PID: 1001},
		{PortNumber: 3000, ProcessName: "node", PID: 1002},
		{PortNumber: 8080, ProcessName: "python", PID: 1003},
		{PortNumber: 443, ProcessName: "nginx", PID: 1004},
		{PortNumber: 5000, ProcessName: "rails", PID: 1005},
		{PortNumber: 80, ProcessName: "httpd", PID: 1006},
	}

	result := sorter.SortByCommonPort(ports)

	expectedOrder := []int{80, 443, 3000, 5000, 8080, 9999}

	if len(result) != len(expectedOrder) {
		t.Fatalf("Result length mismatch: got=%d, want=%d", len(result), len(expectedOrder))
	}

	for i, expectedPort := range expectedOrder {
		if result[i].PortNumber != expectedPort {
			t.Errorf("Position %d: port number mismatch got=%d, want=%d", i, result[i].PortNumber, expectedPort)
		}
	}
}

func TestDefaultPortSorter_SortByPortNumber(t *testing.T) {
	sorter := NewDefaultPortSorter()

	ports := []models.PortInfo{
		{PortNumber: 8080, ProcessName: "python", PID: 1003},
		{PortNumber: 80, ProcessName: "httpd", PID: 1006},
		{PortNumber: 3000, ProcessName: "node", PID: 1002},
	}

	result := sorter.SortByPortNumber(ports)

	expectedOrder := []int{80, 3000, 8080}

	if len(result) != len(expectedOrder) {
		t.Fatalf("Result length mismatch: got=%d, want=%d", len(result), len(expectedOrder))
	}

	for i, expectedPort := range expectedOrder {
		if result[i].PortNumber != expectedPort {
			t.Errorf("Position %d: port number mismatch got=%d, want=%d", i, result[i].PortNumber, expectedPort)
		}
	}
}

func TestDefaultPortSorter_EmptyPorts(t *testing.T) {
	sorter := NewDefaultPortSorter()

	emptyPorts := []models.PortInfo{}

	result := sorter.SortByCommonPort(emptyPorts)

	if len(result) != 0 {
		t.Errorf("Empty list should return empty: got=%d", len(result))
	}

	result = sorter.SortByPortNumber(emptyPorts)

	if len(result) != 0 {
		t.Errorf("Empty list should return empty: got=%d", len(result))
	}
}

func TestScanResult_NewScanResult(t *testing.T) {
	ports := []models.PortInfo{
		{PortNumber: 3000, ProcessName: "node", PID: 1002, IsDocker: true},
		{PortNumber: 8080, ProcessName: "python", PID: 1003, IsDocker: false},
	}

	result := NewScanResult(ports, nil)

	if result.Count != 2 {
		t.Errorf("Port count mismatch: got=%d, want=2", result.Count)
	}

	if !result.HasDocker {
		t.Error("Should include Docker port")
	}

	if len(result.Ports) != 2 {
		t.Errorf("Port list length mismatch: got=%d, want=2", len(result.Ports))
	}
}

func TestScanResult_HasDocker(t *testing.T) {
	tests := []struct {
		name      string
		ports     []models.PortInfo
		hasDocker bool
	}{
		{
			name: "includes Docker port",
			ports: []models.PortInfo{
				{PortNumber: 3000, IsDocker: true},
			},
			hasDocker: true,
		},
		{
			name: "no Docker port",
			ports: []models.PortInfo{
				{PortNumber: 8080, IsDocker: false},
				{PortNumber: 3000, IsDocker: false},
			},
			hasDocker: false,
		},
		{
			name:      "empty list",
			ports:     []models.PortInfo{},
			hasDocker: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewScanResult(tt.ports, nil)
			if result.HasDocker != tt.hasDocker {
				t.Errorf("HasDocker = %v, want %v", result.HasDocker, tt.hasDocker)
			}
		})
	}
}

func TestScanResult_FilteredPorts(t *testing.T) {
	ports := []models.PortInfo{
		{PortNumber: 80, ProcessName: "httpd", PID: 100},
		{PortNumber: 3000, ProcessName: "node", PID: 200},
		{PortNumber: 8080, ProcessName: "python", PID: 300},
		{PortNumber: 443, ProcessName: "nginx", PID: 400},
	}

	result := NewScanResult(ports, nil)

	commonPorts := result.CommonPorts()
	if len(commonPorts) != 4 {
		t.Errorf("Common port count mismatch: got=%d, want=4", len(commonPorts))
	}

	filtered := result.FilteredPorts(func(p models.PortInfo) bool {
		return p.PortNumber == 3000
	})
	if len(filtered) != 1 {
		t.Errorf("Filter result mismatch: got=%d, want=1", len(filtered))
	}
	if filtered[0].PortNumber != 3000 {
		t.Errorf("Filtered port number mismatch: got=%d, want=3000", filtered[0].PortNumber)
	}
}

func TestScanResult_DockerPorts(t *testing.T) {
	ports := []models.PortInfo{
		{PortNumber: 3000, IsDocker: true, ContainerName: "node-app"},
		{PortNumber: 8080, IsDocker: false, ProcessName: "python"},
		{PortNumber: 5432, IsDocker: true, ContainerName: "postgres"},
	}

	result := NewScanResult(ports, nil)
	dockerPorts := result.DockerPorts()

	if len(dockerPorts) != 2 {
		t.Errorf("Docker port count mismatch: got=%d, want=2", len(dockerPorts))
	}

	for _, port := range dockerPorts {
		if !port.IsDocker {
			t.Error("Should only return Docker ports")
		}
	}
}

func TestScanResult_RecommendedPorts(t *testing.T) {
	ports := []models.PortInfo{
		{PortNumber: 3000, KillCount: 5},
		{PortNumber: 8080, KillCount: 1},
		{PortNumber: 5000, KillCount: 10},
		{PortNumber: 9000, KillCount: 0},
	}

	result := NewScanResult(ports, nil)
	recommended := result.RecommendedPorts()

	if len(recommended) != 2 {
		t.Errorf("Recommended port count mismatch: got=%d, want=2", len(recommended))
	}

	for _, port := range recommended {
		if port.KillCount < 3 {
			t.Errorf("Recommended port should have KillCount >= 3: got=%d", port.KillCount)
		}
	}
}

func TestScanResult_SystemPorts(t *testing.T) {
	ports := []models.PortInfo{
		{PortNumber: 80},
		{PortNumber: 443},
		{PortNumber: 3000},
		{PortNumber: 8080},
		{PortNumber: 1024},
	}

	result := NewScanResult(ports, nil)
	systemPorts := result.SystemPorts()

	if len(systemPorts) != 2 {
		t.Errorf("System port count mismatch: got=%d, want=2", len(systemPorts))
	}
}

func Benchmark_SortByCommonPort(b *testing.B) {
	sorter := NewDefaultPortSorter()

	ports := make([]models.PortInfo, 1000)
	for i := 0; i < 1000; i++ {
		ports[i] = models.PortInfo{
			PortNumber:  1000 + i,
			ProcessName: "process",
			PID:         1000 + i,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sorter.SortByCommonPort(ports)
	}
}
