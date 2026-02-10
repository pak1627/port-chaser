// Package scanner에 대한 테스트
package scanner

import (
	"testing"

	"github.com/manson/port-chaser/internal/models"
)

// TestDefaultPortSorter_SortByCommonPort는 일반 포트 정렬을 테스트합니다.
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

	// 일반 포트가 상단에 위치해야 함
	expectedOrder := []int{80, 443, 3000, 5000, 8080, 9999}

	if len(result) != len(expectedOrder) {
		t.Fatalf("결과 길이가 다릅니다: got=%d, want=%d", len(result), len(expectedOrder))
	}

	for i, expectedPort := range expectedOrder {
		if result[i].PortNumber != expectedPort {
			t.Errorf("위치 %d: 포트 번호가 다릅니다 got=%d, want=%d", i, result[i].PortNumber, expectedPort)
		}
	}
}

// TestDefaultPortSorter_SortByPortNumber는 포트 번호 정렬을 테스트합니다.
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
		t.Fatalf("결과 길이가 다릅니다: got=%d, want=%d", len(result), len(expectedOrder))
	}

	for i, expectedPort := range expectedOrder {
		if result[i].PortNumber != expectedPort {
			t.Errorf("위치 %d: 포트 번호가 다릅니다 got=%d, want=%d", i, result[i].PortNumber, expectedPort)
		}
	}
}

// TestDefaultPortSorter_EmptyPorts는 빈 포트 목록 처리를 테스트합니다.
func TestDefaultPortSorter_EmptyPorts(t *testing.T) {
	sorter := NewDefaultPortSorter()

	emptyPorts := []models.PortInfo{}

	result := sorter.SortByCommonPort(emptyPorts)

	if len(result) != 0 {
		t.Errorf("빈 목록은 빈 상태로 반환되어야 합니다: got=%d", len(result))
	}

	result = sorter.SortByPortNumber(emptyPorts)

	if len(result) != 0 {
		t.Errorf("빈 목록은 빈 상태로 반환되어야 합니다: got=%d", len(result))
	}
}

// TestScanResult_NewScanResult는 ScanResult 생성을 테스트합니다.
func TestScanResult_NewScanResult(t *testing.T) {
	ports := []models.PortInfo{
		{PortNumber: 3000, ProcessName: "node", PID: 1002, IsDocker: true},
		{PortNumber: 8080, ProcessName: "python", PID: 1003, IsDocker: false},
	}

	result := NewScanResult(ports, nil)

	if result.Count != 2 {
		t.Errorf("포트 수가 다릅니다: got=%d, want=2", result.Count)
	}

	if !result.HasDocker {
		t.Error("Docker 포트가 포함되어 있어야 합니다")
	}

	if len(result.Ports) != 2 {
		t.Errorf("포트 목록 길이가 다릅니다: got=%d, want=2", len(result.Ports))
	}
}

// TestScanResult_HasDocker는 Docker 포트 감지를 테스트합니다.
func TestScanResult_HasDocker(t *testing.T) {
	tests := []struct {
		name      string
		ports     []models.PortInfo
		hasDocker bool
	}{
		{
			name:      "Docker 포트 포함",
			ports: []models.PortInfo{
				{PortNumber: 3000, IsDocker: true},
			},
			hasDocker: true,
		},
		{
			name: "Docker 포트 없음",
			ports: []models.PortInfo{
				{PortNumber: 8080, IsDocker: false},
				{PortNumber: 3000, IsDocker: false},
			},
			hasDocker: false,
		},
		{
			name:      "빈 목록",
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

// TestScanResult_FilteredPorts는 필터링 기능을 테스트합니다.
func TestScanResult_FilteredPorts(t *testing.T) {
	ports := []models.PortInfo{
		{PortNumber: 80, ProcessName: "httpd", PID: 100},
		{PortNumber: 3000, ProcessName: "node", PID: 200},
		{PortNumber: 8080, ProcessName: "python", PID: 300},
		{PortNumber: 443, ProcessName: "nginx", PID: 400},
	}

	result := NewScanResult(ports, nil)

	// 일반 포트 필터링
	commonPorts := result.CommonPorts()
	if len(commonPorts) != 4 {
		t.Errorf("일반 포트 수가 다릅니다: got=%d, want=4", len(commonPorts))
	}

	// 포트 번호 3000으로 필터링
	filtered := result.FilteredPorts(func(p models.PortInfo) bool {
		return p.PortNumber == 3000
	})
	if len(filtered) != 1 {
		t.Errorf("필터링 결과가 다릅니다: got=%d, want=1", len(filtered))
	}
	if filtered[0].PortNumber != 3000 {
		t.Errorf("필터링된 포트 번호가 다릅니다: got=%d, want=3000", filtered[0].PortNumber)
	}
}

// TestScanResult_DockerPorts는 Docker 포트 필터링을 테스트합니다.
func TestScanResult_DockerPorts(t *testing.T) {
	ports := []models.PortInfo{
		{PortNumber: 3000, IsDocker: true, ContainerName: "node-app"},
		{PortNumber: 8080, IsDocker: false, ProcessName: "python"},
		{PortNumber: 5432, IsDocker: true, ContainerName: "postgres"},
	}

	result := NewScanResult(ports, nil)
	dockerPorts := result.DockerPorts()

	if len(dockerPorts) != 2 {
		t.Errorf("Docker 포트 수가 다릅니다: got=%d, want=2", len(dockerPorts))
	}

	for _, port := range dockerPorts {
		if !port.IsDocker {
			t.Error("Docker 포트만 반환되어야 합니다")
		}
	}
}

// TestScanResult_RecommendedPorts는 추천 포트 필터링을 테스트합니다.
func TestScanResult_RecommendedPorts(t *testing.T) {
	ports := []models.PortInfo{
		{PortNumber: 3000, KillCount: 5},       // 추천
		{PortNumber: 8080, KillCount: 1},       // 비추천
		{PortNumber: 5000, KillCount: 10},      // 추천
		{PortNumber: 9000, KillCount: 0},       // 비추천
	}

	result := NewScanResult(ports, nil)
	recommended := result.RecommendedPorts()

	if len(recommended) != 2 {
		t.Errorf("추천 포트 수가 다릅니다: got=%d, want=2", len(recommended))
	}

	for _, port := range recommended {
		if port.KillCount < 3 {
			t.Errorf("추천 포트는 KillCount가 3 이상이어야 합니다: got=%d", port.KillCount)
		}
	}
}

// TestScanResult_SystemPorts는 시스템 포트 필터링을 테스트합니다.
func TestScanResult_SystemPorts(t *testing.T) {
	ports := []models.PortInfo{
		{PortNumber: 80},    // 시스템 포트
		{PortNumber: 443},   // 시스템 포트
		{PortNumber: 3000},  // 일반 포트
		{PortNumber: 8080},  // 일반 포트
		{PortNumber: 1024},  // 에지 케이스 (1024는 시스템 포트 아님)
	}

	result := NewScanResult(ports, nil)
	systemPorts := result.SystemPorts()

	if len(systemPorts) != 2 {
		t.Errorf("시스템 포트 수가 다릅니다: got=%d, want=2", len(systemPorts))
	}
}

// Benchmark_SortByCommonPort는 정렬 성능을 벤치마킹합니다.
func Benchmark_SortByCommonPort(b *testing.B) {
	sorter := NewDefaultPortSorter()

	// 1000개 포트 생성
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
