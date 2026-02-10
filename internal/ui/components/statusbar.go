// Package components는 상태바 컴포넌트를 제공합니다.
package components

import (
	"strings"

	"github.com/manson/port-chaser/internal/ui"
)

// StatusBar는 TUI 하단 상태바 컴포넌트입니다.
type StatusBar struct {
	styles    *ui.Styles
	bindings  *ui.KeyBindings
	message   string
	messageFn func() string
}

// NewStatusBar는 새로운 StatusBar를 생성합니다.
func NewStatusBar(styles *ui.Styles, bindings *ui.KeyBindings) *StatusBar {
	return &StatusBar{
		styles:   styles,
		bindings: bindings,
	}
}

// Render는 상태바를 렌더링합니다.
func (sb *StatusBar) Render(width int) string {
	// 메시지가 있으면 메시지 표시
	if sb.message != "" {
		return sb.renderMessage(width)
	}

	// 메시지 함수가 있으면 실행
	if sb.messageFn != nil {
		sb.message = sb.messageFn()
		sb.messageFn = nil // 일회성
		return sb.renderMessage(width)
	}

	// 기본 키 바인딩 가이드
	return sb.renderKeyGuide(width)
}

// renderMessage는 메시지를 렌더링합니다.
func (sb *StatusBar) renderMessage(width int) string {
	return sb.styles.StatusBar.Render(sb.message)
}

// renderKeyGuide는 키 바인딩 가이드를 렌더링합니다.
func (sb *StatusBar) renderKeyGuide(width int) string {
	help := []string{
		sb.formatKey("↑", "j") + "/" + sb.formatKey("↓", "k") + ": 선택",
		sb.formatKey("Enter") + ": 종료",
		sb.formatKey("/") + ": 검색",
		sb.formatKey("d") + ": Docker",
		sb.formatKey("q") + ": 종료",
	}

	guide := strings.Join(help, " │ ")

	return sb.styles.StatusBar.Render(guide)
}

// formatKey는 키 표시를 포맷팅합니다.
func (sb *StatusBar) formatKey(keys ...string) string {
	var formatted []string
	for _, key := range keys {
		formatted = append(formatted, sb.styles.StatusKey.Render(key))
	}
	return strings.Join(formatted, "/")
}

// SetMessage는 상태 메시지를 설정합니다.
func (sb *StatusBar) SetMessage(message string) {
	sb.message = message
}

// SetMessageFn은 상태 메시지 함수를 설정합니다 (일회성).
func (sb *StatusBar) SetMessageFn(fn func() string) {
	sb.messageFn = fn
}

// ClearMessage는 상태 메시지를 지웁니다.
func (sb *StatusBar) ClearMessage() {
	sb.message = ""
	sb.messageFn = nil
}
