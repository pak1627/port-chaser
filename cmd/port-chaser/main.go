// Package main은 Port Chaser의 진입점입니다.
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/manson/port-chaser/internal/app"
	"github.com/manson/port-chaser/internal/models"
	"github.com/manson/port-chaser/internal/process"
	"github.com/manson/port-chaser/internal/scanner"
)

const (
	// appName은 애플리케이션 이름입니다.
	appName = "Port Chaser"

	// version은 애플리케이션 버전입니다.
	version = "0.1.0"
)

func main() {
	// 인자 처리
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-v", "--version", "version":
			fmt.Printf("%s v%s\n", appName, version)
			os.Exit(0)
		case "-h", "--help", "help":
			printHelp()
			os.Exit(0)
		}
	}

	// 모델 초기화
	model := initializeModel()

	// Bubbletea 프로그램 시작
	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),       // 대체 스크린 모드
		tea.WithMouseCellMotion(), // 마우스 셀 모션
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("에러: %v\n", err)
		os.Exit(1)
	}
}

// initializeModel은 TUI 모델을 초기화합니다.
func initializeModel() app.Model {
	// 프로세스 종료기 생성
	killer := process.NewProcessKiller()

	// 포트 스캐너 생성 (실제 구현에서는 gopsutil 사용)
	scanner := &mockScanner{
		ports: getMockPorts(),
	}

	return app.Model{
		Ports:          []models.PortInfo{},
		FilteredPorts:  []models.PortInfo{},
		SelectedIndex:  -1,
		ViewMode:       app.ViewModeMain,
		SearchQuery:    "",
		ShowDockerOnly: false,
		History:        []models.HistoryEntry{},
		Loading:        true,
		Width:          80,
		Height:         24,
		Scanner:        scanner,
		Killer:         &killerAdapter{killer: killer},
	}
}

// printHelp는 도움말을 출력합니다.
func printHelp() {
	help := `Port Chaser - 터미널 UI 기반 포트 관리 도구

사용법:
  port-chaser [옵션]

옵션:
  -v, --version     버전 정보 출력
  -h, --help        도움말 출력

TUI 키 바인딩:
  ↑/k, ↓/j          위/아래 이동
  gg, G             맨 위/맨 아래 이동
  Enter             프로세스 종료
  /                 검색
  d                 Docker 포트만 표시 토글
  h                 히스토리 보기
  ?                 도움말
  r                 새로고침
  q, Ctrl+C         종료

프로젝트 홈페이지: https://github.com/manson/port-chaser
`
	fmt.Println(help)
}

// ========== Mock 구현 (실제 구현 시 제거) ==========

// mockScanner는 테스트용 모의 스캐너입니다.
type mockScanner struct {
	ports []models.PortInfo
}

func (m *mockScanner) Scan() ([]models.PortInfo, error) {
	// 실제 스캔 시뮬레이션
	time.Sleep(100 * time.Millisecond)
	return m.ports, nil
}

// getMockPorts는 모의 포트 데이터를 반환합니다.
func getMockPorts() []models.PortInfo {
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
			PortNumber:    8080,
			ProcessName:   "python",
			PID:           23456,
			User:          "developer",
			Command:       "python app.py",
			IsDocker:      false,
			IsSystem:      false,
			KillCount:     0,
			LastKilled:    time.Time{},
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
			PortNumber:    80,
			ProcessName:   "httpd",
			PID:           100,
			User:          "root",
			Command:       "/usr/sbin/httpd -D FOREGROUND",
			IsDocker:      false,
			IsSystem:      true,
			KillCount:     0,
			LastKilled:    time.Time{},
		},
		{
			PortNumber:    9000,
			ProcessName:   "custom-app",
			PID:           45678,
			User:          "developer",
			Command:       "./custom-app",
			IsDocker:      false,
			IsSystem:      false,
			KillCount:     3,
			LastKilled:    now.Add(-2 * time.Hour),
		},
	}
}

// killerAdapter는 process.Killer를 app.Killer 인터페이스에 맞춥니다.
type killerAdapter struct {
	killer *process.ProcessKiller
}

func (a *killerAdapter) Kill(port models.PortInfo) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := a.killer.Kill(ctx, port.PID, &port)
	if err != nil {
		return err
	}
	if !result.Success {
		return fmt.Errorf("종료 실패: %s", result.Message)
	}
	return nil
}
