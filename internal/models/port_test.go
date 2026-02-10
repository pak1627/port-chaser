// Package models에 대한 테스트입니다.
package models

import (
	"testing"
	"time"
)

// TestPortInfo_IsCommonPort는 일반 포트 확인 기능을 테스트합니다.
func TestPortInfo_IsCommonPort(t *testing.T) {
	tests := []struct {
		name string
		port int
		want bool
	}{
		{"HTTP 포트", 80, true},
		{"HTTPS 포트", 443, true},
		{"개발 포트 3000", 3000, true},
		{"개발 포트 5000", 5000, true},
		{"개발 포트 8000", 8000, true},
		{"개발 포트 8080", 8080, true},
		{"일반 포트 8081", 8081, false},
		{"일반 포트 9000", 9000, false},
		{"시스템 포트 22", 22, false},
		{"높은 포트 번호", 5432, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PortInfo{PortNumber: tt.port}
			if got := p.IsCommonPort(); got != tt.want {
				t.Errorf("PortInfo.IsCommonPort() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestPortInfo_IsRecommended는 추천 기능을 테스트합니다.
func TestPortInfo_IsRecommended(t *testing.T) {
	tests := []struct {
		name      string
		killCount int
		want      bool
	}{
		{"3회 종료 - 추천", 3, true},
		{"5회 종료 - 추천", 5, true},
		{"2회 종료 - 비추천", 2, false},
		{"0회 종료 - 비추천", 0, false},
		{"1회 종료 - 비추천", 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PortInfo{KillCount: tt.killCount}
			if got := p.IsRecommended(); got != tt.want {
				t.Errorf("PortInfo.IsRecommended() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestPortInfo_ShouldDisplayWarning은 경고 표시 기능을 테스트합니다.
func TestPortInfo_ShouldDisplayWarning(t *testing.T) {
	tests := []struct {
		name     string
		pid      int
		isSystem bool
		want     bool
	}{
		{"시스템 프로세스", 100, true, true},
		{"낮은 PID", 50, false, true},
		{"일반 프로세스", 1234, false, false},
		{"높은 PID 시스템 프로세스", 500, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PortInfo{PID: tt.pid, IsSystem: tt.isSystem}
			if got := p.ShouldDisplayWarning(); got != tt.want {
				t.Errorf("PortInfo.ShouldDisplayWarning() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestPortInfo_FullFields는 PortInfo의 모든 필드가 올바르게 설정되는지 테스트합니다.
func TestPortInfo_FullFields(t *testing.T) {
	now := time.Now()
	p := PortInfo{
		PortNumber:    3000,
		ProcessName:   "node",
		PID:           1234,
		User:          "user",
		Command:       "npm start",
		IsDocker:      false,
		ContainerID:   "",
		ContainerName: "",
		ImageName:     "",
		IsSystem:      false,
		KillCount:     0,
		LastKilled:    now,
	}

	if p.PortNumber != 3000 {
		t.Errorf("PortNumber = %v, want %v", p.PortNumber, 3000)
	}
	if p.ProcessName != "node" {
		t.Errorf("ProcessName = %v, want %v", p.ProcessName, "node")
	}
	if p.PID != 1234 {
		t.Errorf("PID = %v, want %v", p.PID, 1234)
	}
	if !p.LastKilled.Equal(now) {
		t.Errorf("LastKilled = %v, want %v", p.LastKilled, now)
	}
}

// TestPortInfo_DockerFields는 Docker 관련 필드를 테스트합니다.
func TestPortInfo_DockerFields(t *testing.T) {
	p := PortInfo{
		PortNumber:    5432,
		ProcessName:   "postgres",
		PID:           5678,
		IsDocker:      true,
		ContainerID:   "abc123",
		ContainerName: "db-container",
		ImageName:     "postgres:15",
	}

	if !p.IsDocker {
		t.Error("IsDocker = false, want true")
	}
	if p.ContainerID != "abc123" {
		t.Errorf("ContainerID = %v, want %v", p.ContainerID, "abc123")
	}
	if p.ContainerName != "db-container" {
		t.Errorf("ContainerName = %v, want %v", p.ContainerName, "db-container")
	}
	if p.ImageName != "postgres:15" {
		t.Errorf("ImageName = %v, want %v", p.ImageName, "postgres:15")
	}
}

// TestHistoryEntry_Fields는 HistoryEntry의 모든 필드를 테스트합니다.
func TestHistoryEntry_Fields(t *testing.T) {
	now := time.Now()
	h := HistoryEntry{
		ID:          1,
		PortNumber:  3000,
		ProcessName: "node",
		PID:         1234,
		Command:     "npm start",
		KilledAt:    now,
	}

	if h.ID != 1 {
		t.Errorf("ID = %v, want %v", h.ID, 1)
	}
	if h.PortNumber != 3000 {
		t.Errorf("PortNumber = %v, want %v", h.PortNumber, 3000)
	}
	if h.ProcessName != "node" {
		t.Errorf("ProcessName = %v, want %v", h.ProcessName, "node")
	}
	if h.PID != 1234 {
		t.Errorf("PID = %v, want %v", h.PID, 1234)
	}
	if !h.KilledAt.Equal(now) {
		t.Errorf("KilledAt = %v, want %v", h.KilledAt, now)
	}
}

// TestDockerInfo_Fields는 DockerInfo의 모든 필드를 테스트합니다.
func TestDockerInfo_Fields(t *testing.T) {
	d := DockerInfo{
		ContainerID:   "abc123",
		ContainerName: "test-container",
		ImageName:     "nginx:latest",
	}

	if d.ContainerID != "abc123" {
		t.Errorf("ContainerID = %v, want %v", d.ContainerID, "abc123")
	}
	if d.ContainerName != "test-container" {
		t.Errorf("ContainerName = %v, want %v", d.ContainerName, "test-container")
	}
	if d.ImageName != "nginx:latest" {
		t.Errorf("ImageName = %v, want %v", d.ImageName, "nginx:latest")
	}
}
