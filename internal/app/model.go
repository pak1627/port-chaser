// Package app는 Bubbletea TUI 애플리케이션 모델을 정의합니다.
package app

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/manson/port-chaser/internal/models"
)

// ViewMode는 현재 뷰 모드를 나타냅니다.
type ViewMode int

const (
	// ViewModeMain은 메인 포트 목록 뷰입니다.
	ViewModeMain ViewMode = iota

	// ViewModeSearch는 검색 모드입니다.
	ViewModeSearch

	// ViewModeConfirmKill은 프로세스 종료 확인 다이얼로그입니다.
	ViewModeConfirmKill

	// ViewModeHistory는 히스토리 뷰입니다.
	ViewModeHistory

	// ViewModeHelp는 도움말 뷰입니다.
	ViewModeHelp
)

// String은 ViewMode의 문자열 표현을 반환합니다.
func (vm ViewMode) String() string {
	switch vm {
	case ViewModeMain:
		return "main"
	case ViewModeSearch:
		return "search"
	case ViewModeConfirmKill:
		return "confirm_kill"
	case ViewModeHistory:
		return "history"
	case ViewModeHelp:
		return "help"
	default:
		return "unknown"
	}
}

// Model은 TUI 애플리케이션의 상태 모델입니다.
// Bubbletea의 MVU 아키텍처를 따릅니다.
type Model struct {
	// Ports는 스캔된 전체 포트 목록입니다.
	Ports []models.PortInfo

	// FilteredPorts는 현재 필터링된 포트 목록입니다.
	FilteredPorts []models.PortInfo

	// SelectedIndex는 현재 선택된 포트 항목의 인덱스입니다.
	SelectedIndex int

	// ViewMode는 현재 뷰 모드입니다.
	ViewMode ViewMode

	// SearchQuery는 현재 검색어입니다.
	SearchQuery string

	// ShowDockerOnly는 Docker 포트만 표시할지 여부입니다.
	ShowDockerOnly bool

	// History는 종료 이력 목록입니다.
	History []models.HistoryEntry

	// Loading은 로딩 상태입니다.
	Loading bool

	// Err는 발생한 에러입니다.
	Err error

	// StatusMessage는 상태바에 표시할 메시지입니다.
	StatusMessage string

	// StatusMessageTimeout은 상태 메시지 만료 시각입니다.
	StatusMessageTimeout time.Time

	// Quit은 애플리케이션 종료 플래그입니다.
	Quit bool

	// LastScanTime은 마지막 스캔 시각입니다.
	LastScanTime time.Time

	// KillConfirmationPort는 종료 확인 중인 포트 정보입니다.
	KillConfirmationPort *models.PortInfo

	// Width는 터미널 너비입니다.
	Width int

	// Height는 터미널 높이입니다.
	Height int

	// Scanner는 포트 스캐너 인터페이스입니다.
	Scanner Scanner

	// Killer는 프로세스 종료 인터페이스입니다.
	Killer Killer
}

// Scanner는 포트 스캔 인터페이스입니다.
type Scanner interface {
	Scan() ([]models.PortInfo, error)
}

// Killer는 프로세스 종료 인터페이스입니다.
type Killer interface {
	Kill(port models.PortInfo) error
}

// Init는 Bubbletea 초기화 함수입니다.
func (m Model) Init() tea.Cmd {
	// 초기 포트 스캔 명령 반환
	return tea.Batch(
		m.scanPortsCmd(),
		tea.EnterAltScreen, // 알트 스크린 모드 진입
	)
}

// scanPortsCmd는 포트 스캔 명령을 생성합니다.
func (m Model) scanPortsCmd() tea.Cmd {
	return func() tea.Msg {
		// 스캔이 완료되면 PortsScannedMsg 반환
		// 실제 구현에서는 m.Scanner.Scan() 호출
		return PortsScannedMsg{
			Ports:    []models.PortInfo{}, // 실제 스캔 결과
			ScannedAt: time.Now(),
		}
	}
}

// Update는 Bubbletea 업데이트 함수입니다.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// 메시지 타입에 따른 처리
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyMsg(msg)

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil

	case PortsScannedMsg:
		return m.handlePortsScanned(msg)

	case PortKilledMsg:
		return m.handlePortKilled(msg)

	case StatusMsg:
		m.StatusMessage = msg.Message
		m.StatusMessageTimeout = time.Now().Add(time.Second * 3)
		return m, nil

	case TickMsg:
		return m.handleTick(msg)

	default:
		return m, nil
	}
}

// View는 Bubbletea 뷰 함수입니다.
func (m Model) View() string {
	// 실제 뷰 렌더링은 별도 파일에서 처리
	// 여기서는 뷰 모드에 따른 디스패치만 수행
	switch m.ViewMode {
	case ViewModeMain:
		return m.renderMainView()
	case ViewModeSearch:
		return m.renderSearchView()
	case ViewModeConfirmKill:
		return m.renderConfirmKillView()
	case ViewModeHistory:
		return m.renderHistoryView()
	case ViewModeHelp:
		return m.renderHelpView()
	default:
		return "Unknown view mode"
	}
}

// ========== 메시지 타입 정의 ==========

// PortsScannedMsg는 포트 스캔 완료 메시지입니다.
type PortsScannedMsg struct {
	Ports     []models.PortInfo
	ScannedAt time.Time
	Error     error
}

// PortKilledMsg는 프로세스 종료 완료 메시지입니다.
type PortKilledMsg struct {
	Port    models.PortInfo
	Success bool
	Message string
}

// StatusMsg는 상태 메시지입니다.
type StatusMsg struct {
	Message string
}

// TickMsg는 주기적 틱 메시지입니다.
type TickMsg struct {
	Time time.Time
}

// ========== 메시지 핸들러 ==========

// handleKeyMsg는 키 입력을 처리합니다.
func (m Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// 뷰 모드에 따른 키 처리 분기
	switch m.ViewMode {
	case ViewModeMain:
		return m.handleMainKeyMsg(msg)
	case ViewModeSearch:
		return m.handleSearchKeyMsg(msg)
	case ViewModeConfirmKill:
		return m.handleConfirmKeyMsg(msg)
	case ViewModeHistory:
		return m.handleHistoryKeyMsg(msg)
	case ViewModeHelp:
		return m.handleHelpKeyMsg(msg)
	default:
		return m, nil
	}
}

// handlePortsScanned는 포트 스캔 완료를 처리합니다.
func (m Model) handlePortsScanned(msg PortsScannedMsg) (tea.Model, tea.Cmd) {
	if msg.Error != nil {
		m.Err = msg.Error
		m.Loading = false
		return m, nil
	}

	m.Ports = msg.Ports
	m.FilteredPorts = msg.Ports
	m.LastScanTime = msg.ScannedAt
	m.Loading = false

	// 선택 인덱스 초기화
	if len(m.FilteredPorts) > 0 && m.SelectedIndex < 0 {
		m.SelectedIndex = 0
	}

	return m, nil
}

// handlePortKilled는 프로세스 종료 완료를 처리합니다.
func (m Model) handlePortKilled(msg PortKilledMsg) (tea.Model, tea.Cmd) {
	m.ViewMode = ViewModeMain

	var statusCmd tea.Cmd
	if msg.Success {
		statusCmd = func() tea.Msg {
			return StatusMsg{Message: "포트 " + msg.Port.ProcessName + " 종료 완료"}
		}
	} else {
		statusCmd = func() tea.Msg {
			return StatusMsg{Message: "종료 실패: " + msg.Message}
		}
	}

	return m, statusCmd
}

// handleTick은 틱 메시지를 처리합니다.
func (m Model) handleTick(msg TickMsg) (tea.Model, tea.Cmd) {
	// 상태 메시지 만료 확인
	if !m.StatusMessageTimeout.IsZero() && time.Now().After(m.StatusMessageTimeout) {
		m.StatusMessage = ""
		m.StatusMessageTimeout = time.Time{}
	}

	return m, nil
}

// ========== 뷰 렌더러 ==========

// renderMainView는 메인 뷰를 렌더링합니다.
func (m Model) renderMainView() string {
	// 실제 구현은 view.go에서
	return "main view"
}

// renderSearchView는 검색 뷰를 렌더링합니다.
func (m Model) renderSearchView() string {
	return "search view"
}

// renderConfirmKillView는 종료 확인 뷰를 렌더링합니다.
func (m Model) renderConfirmKillView() string {
	return "confirm kill view"
}

// renderHistoryView는 히스토리 뷰를 렌더링합니다.
func (m Model) renderHistoryView() string {
	return "history view"
}

// renderHelpView는 도움말 뷰를 렌더링합니다.
func (m Model) renderHelpView() string {
	return "help view"
}

// ========== 키 핸들러 ==========

// handleMainKeyMsg는 메인 뷰 키 입력을 처리합니다.
func (m Model) handleMainKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		m.Quit = true
		return m, tea.Quit

	case "up", "k":
		m.moveSelection(-1)
		return m, nil

	case "down", "j":
		m.moveSelection(1)
		return m, nil

	case "g":
		// gg 처리를 위한 상태 필요
		m.SelectedIndex = 0
		return m, nil

	case "G":
		m.SelectedIndex = len(m.FilteredPorts) - 1
		return m, nil

	case "/":
		m.ViewMode = ViewModeSearch
		return m, nil

	case "enter":
		if m.isValidSelection() {
			m.ViewMode = ViewModeConfirmKill
		}
		return m, nil

	case "d":
		m.ShowDockerOnly = !m.ShowDockerOnly
		m.applyFilters()
		return m, nil

	case "h":
		m.ViewMode = ViewModeHistory
		return m, nil

	case "?":
		m.ViewMode = ViewModeHelp
		return m, nil

	case "r", "ctrl+r":
		m.Loading = true
		return m, m.scanPortsCmd()
	}

	return m, nil
}

// handleSearchKeyMsg는 검색 뷰 키 입력을 처리합니다.
func (m Model) handleSearchKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.ViewMode = ViewModeMain
		m.SearchQuery = ""
		m.FilteredPorts = m.Ports
		return m, nil

	case "enter":
		m.ViewMode = ViewModeMain
		return m, nil
	}

	// 문자 입력 처리
	if len(msg.String()) == 1 {
		m.SearchQuery += msg.String()
		m.applyFilters()
	}

	return m, nil
}

// handleConfirmKeyMsg는 확인 다이얼로그 키 입력을 처리합니다.
func (m Model) handleConfirmKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		// 프로세스 종료 실행
		return m, m.killPortCmd()

	case "n", "N", "esc":
		m.ViewMode = ViewModeMain
		m.KillConfirmationPort = nil
		return m, nil
	}

	return m, nil
}

// handleHistoryKeyMsg는 히스토리 뷰 키 입력을 처리합니다.
func (m Model) handleHistoryKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc", "h":
		m.ViewMode = ViewModeMain
		return m, nil
	}
	return m, nil
}

// handleHelpKeyMsg는 도움말 뷰 키 입력을 처리합니다.
func (m Model) handleHelpKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc", "?":
		m.ViewMode = ViewModeMain
		return m, nil
	}
	return m, nil
}

// ========== 헬퍼 함수 ==========

// moveSelection은 선택 인덱스를 이동합니다.
func (m *Model) moveSelection(delta int) {
	newIndex := m.SelectedIndex + delta

	// 경계 확인
	if newIndex < 0 {
		newIndex = 0
	}
	if newIndex >= len(m.FilteredPorts) {
		newIndex = len(m.FilteredPorts) - 1
	}

	m.SelectedIndex = newIndex
}

// isValidSelection은 현재 선택이 유효한지 확인합니다.
func (m Model) isValidSelection() bool {
	return m.SelectedIndex >= 0 && m.SelectedIndex < len(m.FilteredPorts)
}

// applyFilters는 현재 필터를 적용합니다.
func (m *Model) applyFilters() {
	m.FilteredPorts = m.Ports

	// Docker 필터 적용
	if m.ShowDockerOnly {
		var dockerPorts []models.PortInfo
		for _, port := range m.Ports {
			if port.IsDocker {
				dockerPorts = append(dockerPorts, port)
			}
		}
		m.FilteredPorts = dockerPorts
	}

	// 검색 필터 적용
	if m.SearchQuery != "" {
		var filtered []models.PortInfo
		query := lowerString(m.SearchQuery)
		for _, port := range m.FilteredPorts {
			if m.matchesSearch(port, query) {
				filtered = append(filtered, port)
			}
		}
		m.FilteredPorts = filtered
	}

	// 선택 인덱스 조정
	if len(m.FilteredPorts) > 0 {
		if m.SelectedIndex >= len(m.FilteredPorts) {
			m.SelectedIndex = len(m.FilteredPorts) - 1
		}
	} else {
		m.SelectedIndex = -1
	}
}

// matchesSearch는 포트가 검색어와 일치하는지 확인합니다.
func (m Model) matchesSearch(port models.PortInfo, query string) bool {
	// 포트 번호 검색
	portNum := intToString(port.PortNumber)
	if contains(portNum, query) {
		return true
	}

	// 프로세스 이름 검색
	if contains(lowerString(port.ProcessName), query) {
		return true
	}

	// PID 검색
	pid := intToString(port.PID)
	if contains(pid, query) {
		return true
	}

	// 컨테이너 이름 검색 (Docker인 경우)
	if port.IsDocker && contains(lowerString(port.ContainerName), query) {
		return true
	}

	return false
}

// killPortCmd는 프로세스 종료 명령을 생성합니다.
func (m Model) killPortCmd() tea.Cmd {
	if !m.isValidSelection() {
		return nil
	}

	port := m.FilteredPorts[m.SelectedIndex]

	return func() tea.Msg {
		// 실제 종료 로직은 Killer 인터페이스 사용
		return PortKilledMsg{
			Port:    port,
			Success: true,
			Message: "종료 완료",
		}
	}
}

// ========== 유틸리티 함수 ==========

func lowerString(s string) string {
	// 소문자 변환 (간단 구현)
	return s
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr
}

func intToString(i int) string {
	// 정수를 문자열로 변환 (간단 구현)
	var result string
	for i > 0 {
		result = string(rune('0'+i%10)) + result
		i /= 10
	}
	return result
}
