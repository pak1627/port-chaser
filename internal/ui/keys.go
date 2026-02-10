// Package ui는 키 바인딩을 정의합니다.
package ui

import tea "github.com/charmbracelet/bubbletea"

// KeyBinding은 키 바인딩을 나타냅니다.
type KeyBinding struct {
	keys []string
	help struct {
		key string
		desc string
	}
}

// NewBinding은 새로운 키 바인딩을 생성합니다.
func NewBinding(opts ...BindingOpt) KeyBinding {
	kb := KeyBinding{}
	for _, opt := range opts {
		opt(&kb)
	}
	return kb
}

// BindingOpt는 키 바인딩 옵션입니다.
type BindingOpt func(*KeyBinding)

// WithKeys는 키를 설정합니다.
func WithKeys(keys ...string) BindingOpt {
	return func(kb *KeyBinding) {
		kb.keys = keys
	}
}

// WithHelp는 도움말을 설정합니다.
func WithHelp(key, desc string) BindingOpt {
	return func(kb *KeyBinding) {
		kb.help.key = key
		kb.help.desc = desc
	}
}

// Help는 도움말 정보를 반환합니다.
type Help struct {
	Key string
	Desc string
}

// Help는 도움말을 반환합니다.
func (kb KeyBinding) Help() Help {
	return Help{
		Key:  kb.help.key,
		Desc: kb.help.desc,
	}
}

// KeyBindings는 TUI 키 바인딩을 정의합니다.
type KeyBindings struct {
	// Quit는 애플리케이션 종료 키입니다.
	Quit KeyBinding

	// NavigateUp은 위로 이동 키입니다.
	NavigateUp KeyBinding

	// NavigateDown은 아래로 이동 키입니다.
	NavigateDown KeyBinding

	// NavigateTop은 맨 위로 이동 키입니다.
	NavigateTop KeyBinding

	// NavigateBottom은 맨 아래로 이동 키입니다.
	NavigateBottom KeyBinding

	// KillProcess는 프로세스 종료 키입니다.
	KillProcess KeyBinding

	// Search는 검색 모드 진입 키입니다.
	Search KeyBinding

	// Confirm은 확인 키입니다.
	Confirm KeyBinding

	// Cancel은 취소 키입니다.
	Cancel KeyBinding

	// ToggleDockerOnly는 Docker 전용 필터 토글 키입니다.
	ToggleDockerOnly KeyBinding

	// ShowHelp는 도움말 표시 키입니다.
	ShowHelp KeyBinding

	// ShowHistory는 히스토리 표시 키입니다.
	ShowHistory KeyBinding

	// Refresh는 새로고침 키입니다.
	Refresh KeyBinding
}

// DefaultKeyBindings는 기본 키 바인딩을 반환합니다.
func DefaultKeyBindings() *KeyBindings {
	kb := &KeyBindings{}

	// 종료: q 또는 Ctrl+Q
	kb.Quit = NewBinding(
		WithKeys("q", "ctrl+q"),
		WithHelp("q", "종료"),
	)

	// 네비게이션: 위/아래 방향키, j/k (vim 스타일)
	kb.NavigateUp = NewBinding(
		WithKeys("up", "k"),
		WithHelp("↑/k", "위로"),
	)

	kb.NavigateDown = NewBinding(
		WithKeys("down", "j"),
		WithHelp("↓/j", "아래로"),
	)

	// vim 스타일: gg로 맨 위, G로 맨 아래
	kb.NavigateTop = NewBinding(
		WithKeys("g"),
		WithHelp("gg", "맨 위"),
	)

	kb.NavigateBottom = NewBinding(
		WithKeys("G"),
		WithHelp("G", "맨 아래"),
	)

	// 프로세스 종료: Enter
	kb.KillProcess = NewBinding(
		WithKeys("enter"),
		WithHelp("enter", "프로세스 종료"),
	)

	// 검색: /
	kb.Search = NewBinding(
		WithKeys("/"),
		WithHelp("/", "검색"),
	)

	// 확인: y
	kb.Confirm = NewBinding(
		WithKeys("y", "Y"),
		WithHelp("y", "확인"),
	)

	// 취소: n, ESC
	kb.Cancel = NewBinding(
		WithKeys("n", "N", "esc"),
		WithHelp("n/esc", "취소"),
	)

	// Docker 필터: d
	kb.ToggleDockerOnly = NewBinding(
		WithKeys("d"),
		WithHelp("d", "Docker 필터"),
	)

	// 도움말: ?
	kb.ShowHelp = NewBinding(
		WithKeys("?"),
		WithHelp("?", "도움말"),
	)

	// 히스토리: h
	kb.ShowHistory = NewBinding(
		WithKeys("h"),
		WithHelp("h", "히스토리"),
	)

	// 새로고침: r
	kb.Refresh = NewBinding(
		WithKeys("r", "ctrl+r"),
		WithHelp("r", "새고침"),
	)

	return kb
}

// HelpText는 키 바인딩 도움말 텍스트를 반환합니다.
func (kb *KeyBindings) HelpText() string {
	help := []string{
		kb.NavigateUp.Help().Key + "/" + kb.NavigateDown.Help().Key + ": " + "이동",
		kb.KillProcess.Help().Key + ": " + "종료",
		kb.Search.Help().Key + ": " + "검색",
		kb.ToggleDockerOnly.Help().Key + ": " + "Docker 필터",
		kb.Quit.Help().Key + ": " + "종료",
	}
	return " | " + joinStrings(help, " | ")
}

// ShortHelp는 짧은 도움말을 반환합니다.
func (kb *KeyBindings) ShortHelp() []KeyBinding {
	return []KeyBinding{
		kb.NavigateUp,
		kb.NavigateDown,
		kb.KillProcess,
		kb.Search,
		kb.ToggleDockerOnly,
		kb.Quit,
	}
}

// FullHelp는 전체 도움말을 반환합니다.
func (kb *KeyBindings) FullHelp() []KeyBinding {
	return []KeyBinding{
		kb.NavigateUp,
		kb.NavigateDown,
		kb.NavigateTop,
		kb.NavigateBottom,
		kb.KillProcess,
		kb.Search,
		kb.ToggleDockerOnly,
		kb.ShowHistory,
		kb.ShowHelp,
		kb.Refresh,
		kb.Quit,
	}
}

// joinStrings는 문자열 슬라이스를 구분자로 연결합니다.
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}

// Matches는 키가 이 바인딩과 일치하는지 확인합니다.
func (kb KeyBinding) Matches(msg tea.Msg) bool {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return false
	}
	for _, k := range kb.keys {
		if keyMsg.String() == k {
			return true
		}
	}
	return false
}
