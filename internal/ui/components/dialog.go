// Package components는 다이얼로그 컴포넌트를 제공합니다.
package components

import (
	"fmt"
	"strings"

	"github.com/manson/port-chaser/internal/models"
	"github.com/manson/port-chaser/internal/ui"
)

// Dialog는 확인 다이얼로그 컴포넌트입니다.
type Dialog struct {
	styles *ui.Styles
}

// NewDialog는 새로운 Dialog를 생성합니다.
func NewDialog(styles *ui.Styles) *Dialog {
	return &Dialog{
		styles: styles,
	}
}

// RenderConfirmKill은 프로세스 종료 확인 다이얼로그를 렌더링합니다.
func (d *Dialog) RenderConfirmKill(port *models.PortInfo) string {
	if port == nil {
		return ""
	}

	var lines []string

	// 제목
	title := d.styles.DialogTitle.Render("프로세스 종료 확인")
	lines = append(lines, title)

	// 빈 줄
	lines = append(lines, "")

	// 경고 메시지 (시스템 프로세스인 경우)
	if port.ShouldDisplayWarning() {
		warning := ui.RenderWarning(d.styles, "시스템 중요 프로세스입니다!")
		lines = append(lines, warning)
		lines = append(lines, "")
	}

	// 포트 정보
	lines = append(lines, d.formatPortInfo(port))

	// 빈 줄
	lines = append(lines, "")

	// 확인 프롬프트
	prompt := d.styles.StatusKey.Render("[y]") + " 확인  " +
		d.styles.StatusDim.Render("[n/esc] 취소")
	lines = append(lines, prompt)

	// 다이얼로그 박스로 감싸
	content := strings.Join(lines, "\n")
	return d.styles.Dialog.Render(content)
}

// formatPortInfo는 포트 정보를 포맷팅합니다.
func (d *Dialog) formatPortInfo(port *models.PortInfo) string {
	var info []string

	// 마커와 포트 번호
	marker := d.getMarker(port)
	portLine := fmt.Sprintf("포트: %s %d", marker, port.PortNumber)
	info = append(info, portLine)

	// 프로세스 정보
	processLine := fmt.Sprintf("프로세스: %s (PID: %d)", port.ProcessName, port.PID)
	info = append(info, processLine)

	// Docker 정보 (있는 경우)
	if port.IsDocker {
		dockerLine := fmt.Sprintf("Docker: %s (%s)", port.ContainerName, port.ImageName)
		info = append(info, dockerLine)
	}

	// 커맨드
	if port.Command != "" {
		// 커맨드가 너무 길면 자르기
		cmd := port.Command
		if len(cmd) > 50 {
			cmd = cmd[:47] + "..."
		}
		cmdLine := "커맨드: " + cmd
		info = append(info, cmdLine)
	}

	return strings.Join(info, "\n")
}

// getMarker는 포트 마커를 반환합니다.
func (d *Dialog) getMarker(port *models.PortInfo) string {
	if port.IsDocker {
		return ui.RenderDockerMarker(d.styles)
	}
	if port.IsRecommended() {
		return ui.RenderRecommendedMarker(d.styles)
	}
	if port.ShouldDisplayWarning() {
		return ui.RenderSystemMarker(d.styles)
	}
	return ""
}
