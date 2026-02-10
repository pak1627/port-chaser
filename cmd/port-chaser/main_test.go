package main

import (
	"context"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/manson/port-chaser/internal/models"
	"github.com/manson/port-chaser/internal/process"
)

func TestE2E_HelpFlag(t *testing.T) {
	if testing.Short() {
		t.Skip("E2E test skipped (-short)")
	}

	cmd := exec.Command("go", "run", "cmd/port-chaser/main.go", "--help")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("help execution failed: %v", err)
	}

	outputStr := string(output)
	expectedKeywords := []string{
		"Port Chaser",
		"Terminal UI",
		"Usage",
		"Options",
		"Key bindings",
	}

	for _, keyword := range expectedKeywords {
		if !strings.Contains(outputStr, keyword) {
			t.Errorf("help output missing '%s'", keyword)
		}
	}

	t.Logf("help output:\n%s", outputStr)
}

func TestE2E_VersionFlag(t *testing.T) {
	cmd := exec.Command("go", "run", "cmd/port-chaser/main.go", "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("version execution failed: %v", err)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Port Chaser") {
		t.Error("version output missing app name")
	}
	if !strings.Contains(outputStr, "v0.1.0") {
		t.Error("version output missing version number")
	}

	t.Logf("version output: %s", outputStr)
}

func TestE2E_InvalidFlags(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "run", "cmd/port-chaser/main.go", "--invalid")

	err := cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		return
	}

	if err != nil {
		t.Logf("invalid flag handling: %v", err)
	}
}

func TestE2E_ModelInitialization(t *testing.T) {
	model := initializeModel()

	if model.Ports == nil {
		t.Error("Ports not initialized")
	}
	if model.FilteredPorts == nil {
		t.Error("FilteredPorts not initialized")
	}
	if model.History == nil {
		t.Error("History not initialized")
	}
	if model.Scanner == nil {
		t.Error("Scanner not set")
	}
	if model.Killer == nil {
		t.Error("Killer not set")
	}

	if !model.Loading {
		t.Error("initial Loading state should be true")
	}
	if model.SelectedIndex != -1 {
		t.Error("initial SelectedIndex should be -1")
	}
	if model.ViewMode != 0 {
		t.Errorf("initial ViewMode should be Main: got=%d", model.ViewMode)
	}
}

func TestE2E_MockScanner(t *testing.T) {
	scanner := &testMockScanner{ports: getTestMockPorts()}

	ports, err := scanner.Scan()
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if len(ports) != 5 {
		t.Errorf("expected 5 ports: got=%d", len(ports))
	}

	port3000 := findPortByNumber(ports, 3000)
	if port3000 == nil {
		t.Error("port 3000 not found")
	} else {
		if !port3000.IsDocker {
			t.Error("port 3000 should be Docker")
		}
		if port3000.KillCount != 5 {
			t.Errorf("port 3000 KillCount should be 5: got=%d", port3000.KillCount)
		}
	}

	port80 := findPortByNumber(ports, 80)
	if port80 == nil {
		t.Error("port 80 not found")
	} else {
		if !port80.IsSystem {
			t.Error("port 80 should be system process")
		}
		if port80.PID != 100 {
			t.Errorf("port 80 PID should be 100: got=%d", port80.PID)
		}
	}
}

func TestE2E_MockPortsData(t *testing.T) {
	ports := getTestMockPorts()

	for i, port := range ports {
		if port.ProcessName == "" {
			t.Errorf("port %d: ProcessName is empty", i)
		}
		if port.User == "" {
			t.Errorf("port %d: User is empty", i)
		}
		if port.PID <= 0 {
			t.Errorf("port %d: invalid PID: %d", i, port.PID)
		}
	}

	foundCommon := 0
	for _, port := range ports {
		if port.IsCommonPort() {
			foundCommon++
		}
	}
	if foundCommon < 2 {
		t.Errorf("expected at least 2 common ports: got=%d", foundCommon)
	}

	dockerCount := 0
	for _, port := range ports {
		if port.IsDocker {
			dockerCount++
			if port.ContainerID == "" {
				t.Error("Docker port ContainerID is empty")
			}
			if port.ContainerName == "" {
				t.Error("Docker port ContainerName is empty")
			}
			if port.ImageName == "" {
				t.Error("Docker port ImageName is empty")
			}
		}
	}
	if dockerCount < 2 {
		t.Errorf("expected at least 2 Docker ports: got=%d", dockerCount)
	}
}

func TestE2E_CommonPortDetection(t *testing.T) {
	tests := []struct {
		name     string
		port     int
		expected bool
	}{
		{"port 80", 80, true},
		{"port 443", 443, true},
		{"port 3000", 3000, true},
		{"port 5000", 5000, true},
		{"port 8000", 8000, true},
		{"port 8080", 8080, true},
		{"port 9999", 9999, false},
		{"port 12345", 12345, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			port := models.PortInfo{PortNumber: tt.port}
			got := port.IsCommonPort()
			if got != tt.expected {
				t.Errorf("IsCommonPort(%d) = %v, want %v", tt.port, got, tt.expected)
			}
		})
	}
}

func TestE2E_RecommendedPortDetection(t *testing.T) {
	tests := []struct {
		name      string
		killCount int
		expected  bool
	}{
		{"recommended (5 times)", 5, true},
		{"recommended (3 times)", 3, true},
		{"not recommended (2 times)", 2, false},
		{"not recommended (0 times)", 0, false},
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

func TestE2E_SystemProcessDetection(t *testing.T) {
	tests := []struct {
		name     string
		pid      int
		isSystem bool
		expected bool
	}{
		{"system process (PID 1)", 1, true, true},
		{"system process (PID 50)", 50, false, true},
		{"normal process", 1234, false, false},
		{"normal process (high PID)", 99999, true, false},
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

func TestE2E_KillerAdapter(t *testing.T) {
	adapter := &killerAdapter{killer: process.NewProcessKiller()}

	port := models.PortInfo{
		PID:         99999,
		PortNumber:  3000,
		ProcessName: "test",
		IsSystem:    false,
	}

	err := adapter.Kill(port)
	if err != nil {
		t.Logf("kill attempt result (expected): %v", err)
	}
}

func TestE2E_FullWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("E2E test skipped (-short)")
	}

	t.Run("step 1: help output", func(t *testing.T) {
		cmd := exec.Command("go", "run", "cmd/port-chaser/main.go", "-h")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("help execution failed: %v", err)
		}
		if len(output) < 100 {
			t.Error("help output too short")
		}
	})

	t.Run("step 2: version check", func(t *testing.T) {
		cmd := exec.Command("go", "run", "cmd/port-chaser/main.go", "-v")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("version execution failed: %v", err)
		}
		if !strings.Contains(string(output), "v0.1.0") {
			t.Error("version info incorrect")
		}
	})

	t.Run("step 3: model init", func(t *testing.T) {
		model := initializeModel()
		if model.Scanner == nil {
			t.Error("Scanner not initialized")
		}
		if model.Killer == nil {
			t.Error("Killer not initialized")
		}
	})

	t.Run("step 4: port scan", func(t *testing.T) {
		scanner := &testMockScanner{ports: getTestMockPorts()}
		ports, err := scanner.Scan()
		if err != nil {
			t.Fatalf("scan failed: %v", err)
		}
		if len(ports) != 5 {
			t.Errorf("expected 5 ports: got=%d", len(ports))
		}
	})
}

func TestE2E_ConcurrentExecution(t *testing.T) {
	if testing.Short() {
		t.Skip("E2E test skipped (-short)")
	}

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

	for i := 0; i < len(commands); i++ {
		if err := <-done; err != nil {
			t.Errorf("command %d execution failed: %v", i, err)
		}
	}
}

func findPortByNumber(ports []models.PortInfo, portNumber int) *models.PortInfo {
	for _, port := range ports {
		if port.PortNumber == portNumber {
			return &port
		}
	}
	return nil
}

func BenchmarkE2E_ModelInitialization(b *testing.B) {
	for i := 0; i < b.N; i++ {
		initializeModel()
	}
}

func BenchmarkE2E_Scan(b *testing.B) {
	scanner := &testMockScanner{ports: getTestMockPorts()}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scanner.Scan()
	}
}

type testMockScanner struct {
	ports []models.PortInfo
}

func (m *testMockScanner) Scan() ([]models.PortInfo, error) {
	return m.ports, nil
}

func getTestMockPorts() []models.PortInfo {
	now := time.Now()
	return []models.PortInfo{
		{
			PortNumber:    3000,
			ProcessName:   "node",
			PID:           12345,
			User:          "developer",
			Command:       "npm start",
			IsDocker:      true,
			ContainerID:   "abc123",
			ContainerName: "my-app",
			ImageName:     "node:16-alpine",
			IsSystem:      false,
			KillCount:     5,
			LastKilled:    now.Add(-24 * time.Hour),
		},
		{
			PortNumber:  8080,
			ProcessName: "python",
			PID:         23456,
			User:        "developer",
			Command:     "python app.py",
			IsDocker:    false,
			IsSystem:    false,
			KillCount:   0,
			LastKilled:  time.Time{},
		},
		{
			PortNumber:    5432,
			ProcessName:   "postgres",
			PID:           34567,
			User:          "postgres",
			Command:       "postgres -D /usr/local/var/postgres",
			IsDocker:      true,
			ContainerID:   "def456",
			ContainerName: "db",
			ImageName:     "postgres:14",
			IsSystem:      false,
			KillCount:     1,
			LastKilled:    now.Add(-48 * time.Hour),
		},
		{
			PortNumber:  80,
			ProcessName: "httpd",
			PID:         100,
			User:        "root",
			Command:     "/usr/sbin/httpd -D FOREGROUND",
			IsDocker:    false,
			IsSystem:    true,
			KillCount:   0,
			LastKilled:  time.Time{},
		},
		{
			PortNumber:  9000,
			ProcessName: "custom-app",
			PID:         45678,
			User:        "developer",
			Command:     "./custom-app",
			IsDocker:    false,
			IsSystem:    false,
			KillCount:   3,
			LastKilled:  now.Add(-2 * time.Hour),
		},
	}
}
