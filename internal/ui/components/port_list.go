// Package components는 TUI 컴포넌트를 제공합니다.
package components

import (
	"strings"

	"github.com/manson/port-chaser/internal/models"
	"github.com/manson/port-chaser/internal/ui"
)

// PortList는 포트 목록 컴포넌트입니다.
type PortList struct {
	styles *ui.Styles
}

// NewPortList는 새로운 PortList를 생성합니다.
func NewPortList(styles *ui.Styles) *PortList {
	return &PortList{
		styles: styles,
	}
}

// Render는 포트 목록을 렌더링합니다.
func (pl *PortList) Render(ports []models.PortInfo, selectedIndex int, width int) string {
	if len(ports) == 0 {
		return pl.renderEmpty()
	}

	var sb strings.Builder

	// 포트 헤더
	sb.WriteString(pl.renderHeader())

	// 포트 항목
	for i, port := range ports {
	 isSelected := i == selectedIndex
		sb.WriteString(pl.renderPortItem(port, isSelected))
	}

	return sb.String()
}

// renderEmpty는 빈 목록 메시지를 렌더링합니다.
func (pl *PortList) renderEmpty() string {
	return pl.styles.Muted.Render("  활성 포트가 없습니다")
}

// renderHeader는 포트 목록 헤더를 렌더링합니다.
func (pl *PortList) renderHeader() string {
	// 포트 번호, 프로세스, PID, 사용자, 커맨드
	header := "  포트    프로세스       PID     사용자    커맨드"
	return pl.styles.Muted.Render(header)
}

// renderPortItem은 단일 포트 항목을 렌더링합니다.
func (pl *PortList) renderPortItem(port models.PortInfo, isSelected bool) string {
	// 마커 생성
	markers := pl.renderMarkers(port)

	// 포트 정보
	portNum := pl.formatPortNumber(port.PortNumber)
	processName := pl.formatProcessName(port.ProcessName, 16)
	pid := pl.formatPID(port.PID)
	user := pl.formatUser(port.User, 10)
	command := pl.formatCommand(port.Command, 30)

	// 라인 조립
	line := strings.Join([]string{
		markers,
		portNum,
		processName,
		pid,
		user,
		command,
	}, " ")

	// 선택 상태에 따라 스타일 적용
	if isSelected {
		return pl.styles.PortSelected.Render(line)
	}
	return pl.styles.PortItem.Render(line)
}

// renderMarkers는 포트 마커를 렌더링합니다.
func (pl *PortList) renderMarkers(port models.PortInfo) string {
	var markers []string

	if port.IsDocker {
		markers = append(markers, ui.RenderDockerMarker(pl.styles))
	}

	if port.IsRecommended() {
		markers = append(markers, ui.RenderRecommendedMarker(pl.styles))
	}

	if port.ShouldDisplayWarning() {
		markers = append(markers, ui.RenderSystemMarker(pl.styles))
	}

	if len(markers) == 0 {
		return "   " // 마커 없으면 공백
	}

	return strings.Join(markers, " ")
}

// formatPortNumber는 포트 번호를 포맷팅합니다.
func (pl *PortList) formatPortNumber(portNum int) string {
	s := intToString(portNum)
	return padRight(s, 6)
}

// formatProcessName은 프로세스 이름을 포맷팅합니다.
func (pl *PortList) formatProcessName(name string, maxWidth int) string {
	if len(name) > maxWidth {
		return name[:maxWidth-3] + "..."
	}
	return padRight(name, maxWidth)
}

// formatPID는 PID를 포맷팅합니다.
func (pl *PortList) formatPID(pid int) string {
	s := intToString(pid)
	return padRight(s, 7)
}

// formatUser는 사용자 이름을 포맷팅합니다.
func (pl *PortList) formatUser(user string, maxWidth int) string {
	if user == "" {
		user = "-"
	}
	if len(user) > maxWidth {
		return user[:maxWidth]
	}
	return padRight(user, maxWidth)
}

// formatCommand는 커맨드를 포맷팅합니다.
func (pl *PortList) formatCommand(command string, maxWidth int) string {
	if command == "" {
		return "-"
	}
	if len(command) > maxWidth {
		return command[:maxWidth-3] + "..."
	}
	return command
}

// ========== 유틸리티 함수 ==========

func padRight(s string, width int) string {
	for len(s) < width {
		s += " "
	}
	return s
}

func intToString(i int) string {
	if i == 0 {
		return "0"
	}
	var result string
	for i > 0 {
		result = string(rune('0'+i%10)) + result
		i /= 10
	}
	return result
}
