// Package components는 헤더 컴포넌트를 제공합니다.
package components

import (
	"fmt"
	"strings"

	"github.com/manson/port-chaser/internal/ui"
)

// Header는 TUI 헤더 컴포넌트입니다.
type Header struct {
	styles *ui.Styles
	title  string
}

// NewHeader는 새로운 Header를 생성합니다.
func NewHeader(styles *ui.Styles, title string) *Header {
	return &Header{
		styles: styles,
		title:  title,
	}
}

// Render는 헤더를 렌더링합니다.
func (h *Header) Render(portCount int, showDockerOnly bool, width int) string {
	// 제목 라인
	titleLine := h.renderTitleLine(portCount, showDockerOnly)

	// 경계선
	divider := strings.Repeat("─", width-2) // 좌우 패딩 고려

	return h.styles.Header.Render(titleLine + "\n" + divider)
}

// renderTitleLine은 제목 라인을 렌더링합니다.
func (h *Header) renderTitleLine(portCount int, showDockerOnly bool) string {
	// 제목
	title := h.styles.HeaderTitle.Render(h.title)

	// 포트 수
	portText := fmt.Sprintf("%d개 포트", portCount)
	portCountDisplay := h.styles.HeaderSubtitle.Render(portText)

	// Docker 필터 상태
	var filterDisplay string
	if showDockerOnly {
		filterDisplay = h.styles.DockerMarker.Render("[Docker만]")
	} else {
		filterDisplay = h.styles.Muted.Render("[전체]")
	}

	// 도움말 힌트
	helpHint := h.styles.StatusKey.Render("[?]")

	// 조립
	left := title + " " + portCountDisplay + " " + filterDisplay
	right := helpHint

	// 중간 공백 계산
	availableWidth := 80 // 기준 너비
	leftWidth := len(left)
	rightWidth := len(right)
	padding := availableWidth - leftWidth - rightWidth

	if padding < 1 {
		padding = 1
	}

	middle := strings.Repeat(" ", padding)

	return left + middle + right
}

// SetTitle은 헤더 제목을 설정합니다.
func (h *Header) SetTitle(title string) {
	h.title = title
}
