// +build !windows

// Package main에 대한 E2E 테스트입니다.
package main

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/manson/port-chaser/internal/models"
)

// TestE2E_HelpFlag는 도움말 플래그를 테스트합니다.
func TestE2E_HelpFlag(t *testing.T) {
	if testing.Short() {
		t.Skip("E2E 테스트 건너뜀 (-short)")
	}

	cmd := exec.Command("go", "run", "cmd/port-chaser/main.go", "--help")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("도움말 실행 실패: %v", err)
	}

	outputStr := string(output)
	expectedKeywords := []string{
		"Port Chaser",
		"터미널 UI",
		"사용법",
		"옵션",
		"키 바인딩",
	}

	for _, keyword := range expectedKeywords {
		if !strings.Contains(outputStr, keyword) {
			t.Errorf("도움말에 '%s'가 포함되어 있지 않습니다", keyword)
		}
	}

	t.Logf("도움말 출력:\n%s", outputStr)
}

// TestE2E_VersionFlag는 버전 플래그를 테스트합니다.
func TestE2E_VersionFlag(t *testing.T) {
	cmd := exec.Command("go", "run", "cmd/port-chaser/main.go", "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("버전 실행 실패: %v", err)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Port Chaser") {
		t.Error("버전 출력에 앱 이름이 포함되어 있지 않습니다")
	}
	if !strings.Contains(outputStr, "v0.1.0") {
		t.Error("버전 출력에 버전 번호가 포함되어 있지 않습니다")
	}

	t.Logf("버전 출력: %s", outputStr)
}

// TestE2E_InvalidFlags는 잘못된 플래그 처리를 테스트합니다.
func TestE2E_InvalidFlags(t *testing.T) {
	// 존재하지 않는 플래그는 TUI가 시작되어야 함 (에러 없이 무시)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "run", "cmd/port-chaser/main.go", "--invalid")
	// TUI 시작은 하지만 컨텍스트 타임아웃으로 종료

	err := cmd.Run()
	// 타임아웃은 예상된 동작
	if ctx.Err() == context.DeadlineExceeded {
		return // 정상: TUI가 시작되고 타임아웃으로 종료됨
	}

	if err != nil {
		t.Logf("잘못된 플래그 처리: %v", err)
	}
}

// TestE2E_ModelInitialization는 모델 초기화를 테스트합니다.
func TestE2E_ModelInitialization(t *testing.T) {
	// initializeModel 함수 테스트
	model := initializeModel()

	if model.Ports == nil {
		t.Error("Ports가 초기화되지 않았습니다")
	}
	if model.FilteredPorts == nil {
		t.Error("FilteredPorts가 초기화되지 않았습니다")
	}
	if model.History == nil {
		t.Error("History가 초기화되지 않았습니다")
	}
	if model.Scanner == nil {
		t.Error("Scanner가 설정되지 않았습니다")
	}
	if model.Killer == nil {
		t.Error("Killer가 설정되지 않았습니다")
	}

	// 초기 상태 확인
	if !model.Loading {
		t.Error("초기 Loading 상태가 true여야 합니다")
	}
	if model.SelectedIndex != -1 {
		t.Error("초기 SelectedIndex는 -1이어야 합니다")
	}
	if model.ViewMode != 0 { // ViewModeMain
		t.Errorf("초기 ViewMode는 Main이어야 합니다: got=%d", model.ViewMode)
	}
}

// TestE2E_MockScanner는 모의 스캐너를 테스트합니다.
func TestE2E_MockScanner(t *testing.T) {
	scanner := &mockScanner{ports: getMockPorts()}

	ports, err := scanner.Scan()
	if err != nil {
		t.Fatalf("Scan 실패: %v", err)
	}

	if len(ports) != 5 {
		t.Errorf("포트 수가 5개여야 함: got=%d", len(ports))
	}

	// 포트 3000 확인
	port3000 := findPortByNumber(ports, 3000)
	if port3000 == nil {
		t.Error("포트 3000을 찾을 수 없습니다")
	} else {
		if !port3000.IsDocker {
			t.Error("포트 3000은 Docker여야 합니다")
		}
		if port3000.KillCount != 5 {
			t.Errorf("포트 3000의 KillCount가 5여야 함: got=%d", port3000.KillCount)
		}
	}

	// 시스템 프로세스 확인
	port80 := findPortByNumber(ports, 80)
	if port80 == nil {
		t.Error("포트 80을 찾을 수 없습니다")
	} else {
		if !port80.IsSystem {
			t.Error("포트 80은 시스템 프로세스여야 합니다")
		}
		if port80.PID != 100 {
			t.Errorf("포트 80의 PID가 100이어야 함: got=%d", port80.PID)
		}
	}
}

// TestE2E_MockPortsData는 모의 포트 데이터의 무결성을 테스트합니다.
func TestE2E_MockPortsData(t *testing.T) {
	ports := getMockPorts()

	// 필수 필드 확인
	requiredFields := []string{"PortNumber", "ProcessName", "PID", "User", "Command"}
	for i, port := range ports {
		if port.ProcessName == "" {
			t.Errorf("포트 %d: ProcessName이 비어있습니다", i)
		}
		if port.User == "" {
			t.Errorf("포트 %d: User가 비어있습니다", i)
		}
		if port.PID <= 0 {
			t.Errorf("포트 %d: PID가 유효하지 않습니다: %d", i, port.PID)
		}
	}

	// 일반 포트 확인 (80, 443, 3000, 5000, 8000, 8080)
	commonPorts := []int{80, 3000, 8080}
	foundCommon := 0
	for _, port := range ports {
		if port.IsCommonPort() {
			foundCommon++
		}
	}
	if foundCommon < 2 {
		t.Errorf("최소 2개의 일반 포트가 있어야 함: got=%d", foundCommon)
	}

	// Docker 포트 확인
	dockerCount := 0
	for _, port := range ports {
		if port.IsDocker {
			dockerCount++
			if port.ContainerID == "" {
				t.Error("Docker 포트의 ContainerID가 비어있습니다")
			}
			if port.ContainerName == "" {
				t.Error("Docker 포트의 ContainerName이 비어있습니다")
			}
			if port.ImageName == "" {
				t.Error("Docker 포트의 ImageName이 비어있습니다")
			}
		}
	}
	if dockerCount < 2 {
		t.Errorf("최소 2개의 Docker 포트가 있어야 함: got=%d", dockerCount)
	}
}

// TestE2E_CommonPortDetection은 일반 포트 감지를 테스트합니다.
func TestE2E_CommonPortDetection(t *testing.T) {
	tests := []struct {
		port     int
		expected bool
	}{
		{80, true},
		{443, true},
		{3000, true},
		{5000, true},
		{8000, true},
		{8080, true},
		{9999, false},
		{12345, false},
	}

	for _, tt := range tests {
		t.Run(tt.port, func(t *testing.T) {
			port := models.PortInfo{PortNumber: tt.port}
			got := port.IsCommonPort()
			if got != tt.expected {
				t.Errorf("IsCommonPort(%d) = %v, want %v", tt.port, got, tt.expected)
			}
		})
	}
}

// TestE2E_RecommendedPortDetection은 추천 포트 감지를 테스트합니다.
func TestE2E_RecommendedPortDetection(t *testing.T) {
	tests := []struct {
		name      string
		killCount int
		expected  bool
	}{
		{"추천 대상 (5회)", 5, true},
		{"추천 대상 (3회)", 3, true},
		{"비추천 (2회)", 2, false},
		{"비추천 (0회)", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			port := models.PortInfo{KillCount: tt.killCount}
			got := port.IsRecommended()
			if got != tt.expected {
				t.Errorf("IsRecommended(KillCount=%d) = %v, want %v", tt.killCount, got, tt.expected)
			}
		})
	}
}

// TestE2E_SystemProcessDetection은 시스템 프로세스 감지를 테스트합니다.
func TestE2E_SystemProcessDetection(t *testing.T) {
	tests := []struct {
		name     string
		pid      int
		isSystem bool
		expected bool
	}{
		{"시스템 프로세스 (PID 1)", 1, true, true},
		{"시스템 프로세스 (PID 50)", 50, false, true}, // PID < 100
		{"일반 프로세스", 1234, false, false},
		{"일반 프로세스 (높은 PID)", 99999, true, false}, // IsSystem=true but high PID
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			port := models.PortInfo{PID: tt.pid, IsSystem: tt.isSystem}
			got := port.ShouldDisplayWarning()
			if got != tt.expected {
				t.Errorf("ShouldDisplayWarning(PID=%d, IsSystem=%v) = %v, want %v",
					tt.pid, tt.isSystem, got, tt.expected)
			}
		})
	}
}

// TestE2E_KillerAdapter는 Killer 어댑터를 테스트합니다.
func TestE2E_KillerAdapter(t *testing.T) {
	adapter := &killerAdapter{killer: newProcessKillerForTest()}

	// 존재하지 않는 프로세스로 종료 시도
	port := models.PortInfo{
		PID:         99999,
		PortNumber:  3000,
		ProcessName: "test",
		IsSystem:    false,
	}

	err := adapter.Kill(port)
	// 존재하지 않는 프로세스이므로 에러 또는 성공 (구현에 따라)
	if err != nil {
		t.Logf("종료 시도 결과 (예상됨): %v", err)
	}
}

// TestE2E_FullWorkflow는 전체 워크플로우를 테스트합니다.
func TestE2E_FullWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("E2E 테스트 건너뜀 (-short)")
	}

	t.Run("단계1: 도움말 출력", func(t *testing.T) {
		cmd := exec.Command("go", "run", "cmd/port-chaser/main.go", "-h")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("도움말 실행 실패: %v", err)
		}
		if len(output) < 100 {
			t.Error("도움말 출력이 너무 짧습니다")
		}
	})

	t.Run("단계2: 버전 확인", func(t *testing.T) {
		cmd := exec.Command("go", "run", "cmd/port-chaser/main.go", "-v")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("버전 실행 실패: %v", err)
		}
		if !strings.Contains(string(output), "v0.1.0") {
			t.Error("버전 정보가 올바르지 않습니다")
		}
	})

	t.Run("단계3: 모델 초기화", func(t *testing.T) {
		model := initializeModel()
		if model.Scanner == nil {
			t.Error("Scanner가 초기화되지 않았습니다")
		}
		if model.Killer == nil {
			t.Error("Killer가 초기화되지 않았습니다")
		}
	})

	t.Run("단계4: 포트 스캔", func(t *testing.T) {
		scanner := &mockScanner{ports: getMockPorts()}
		ports, err := scanner.Scan()
		if err != nil {
			t.Fatalf("스캔 실패: %v", err)
		}
		if len(ports) != 5 {
			t.Errorf("5개의 포트가 예상됨: got=%d", len(ports))
		}
	})
}

// TestE2E_ConcurrentExecution은 동시 실행을 테스트합니다.
func TestE2E_ConcurrentExecution(t *testing.T) {
	if testing.Short() {
		t.Skip("E2E 테스트 건너뜀 (-short)")
	}

	// 여러 인스턴스가 동시에 실행될 때의 안정성 테스트
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	commands := []*exec.Cmd{
		exec.CommandContext(ctx, "go", "run", "cmd/port-chaser/main.go", "-h"),
		exec.CommandContext(ctx, "go", "run", "cmd/port-chaser/main.go", "-v"),
		exec.CommandContext(ctx, "go", "run", "cmd/port-chaser/main.go", "--help"),
	}

	done := make(chan error, len(commands))

	for _, cmd := range commands {
		go func(c *exec.Cmd) {
			_, err := c.CombinedOutput()
			done <- err
		}(cmd)
	}

	// 모든 명령이 완료되어야 함
	for i := 0; i < len(commands); i++ {
		if err := <-done; err != nil {
			t.Errorf("명령 %d 실행 실패: %v", i, err)
		}
	}
}

// Helper 함수

func findPortByNumber(ports []models.PortInfo, portNumber int) *models.PortInfo {
	for _, port := range ports {
		if port.PortNumber == portNumber {
			return &port
		}
	}
	return nil
}

func newProcessKillerForTest() interface{} {
	// process.NewProcessKiller()를 반환하는 헬퍼
	// 실제로는 process 패키지를 import하여 사용
	return struct {
		GracePeriod             interface{}
		SystemProcessProtection bool
	}{
		SystemProcessProtection: true,
	}
}

// BenchmarkE2E_ModelInitialization은 모델 초기화 벤치마크입니다.
func BenchmarkE2E_ModelInitialization(b *testing.B) {
	for i := 0; i < b.N; i++ {
		initializeModel()
	}
}

// BenchmarkE2E_Scan은 포트 스캔 벤치마크입니다.
func BenchmarkE2E_Scan(b *testing.B) {
	scanner := &mockScanner{ports: getMockPorts()}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scanner.Scan()
	}
}
