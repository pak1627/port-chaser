// Package ui에 대한 테스트입니다.
package ui

import (
	"testing"
)

// TestDefaultKeyBindings는 기본 키 바인딩 생성을 테스트합니다.
func TestDefaultKeyBindings(t *testing.T) {
	kb := DefaultKeyBindings()

	if kb == nil {
		t.Fatal("DefaultKeyBindings가 nil을 반환했습니다")
	}

	// 모든 바인딩이 설정되어 있는지 확인
	tests := []struct {
		name    string
		binding KeyBinding
	}{
		{"Quit", kb.Quit},
		{"NavigateUp", kb.NavigateUp},
		{"NavigateDown", kb.NavigateDown},
		{"NavigateTop", kb.NavigateTop},
		{"NavigateBottom", kb.NavigateBottom},
		{"KillProcess", kb.KillProcess},
		{"Search", kb.Search},
		{"Confirm", kb.Confirm},
		{"Cancel", kb.Cancel},
		{"ToggleDockerOnly", kb.ToggleDockerOnly},
		{"ShowHelp", kb.ShowHelp},
		{"ShowHistory", kb.ShowHistory},
		{"Refresh", kb.Refresh},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			help := tt.binding.Help()
			if help.Key == "" {
				t.Errorf("%s 바인딩에 Help Key가 없습니다", tt.name)
			}
			if help.Desc == "" {
				t.Errorf("%s 바인딩에 Help Desc가 없습니다", tt.name)
			}
		})
	}
}

// TestKeyBindings_HelpText는 HelpText 메서드를 테스트합니다.
func TestKeyBindings_HelpText(t *testing.T) {
	kb := DefaultKeyBindings()
	text := kb.HelpText()

	if text == "" {
		t.Error("HelpText가 비어있습니다")
	}

	// 주요 키가 포함되어 있는지 확인
	expectedKeys := []string{"↑/k", "↓/j", "enter", "q"}
	for _, expectedKey := range expectedKeys {
		if !contains(text, expectedKey) {
			t.Errorf("HelpText에 '%s'가 포함되어 있지 않습니다: %s", expectedKey, text)
		}
	}

	t.Logf("HelpText: %s", text)
}

// TestKeyBindings_ShortHelp는 ShortHelp 메서드를 테스트합니다.
func TestKeyBindings_ShortHelp(t *testing.T) {
	kb := DefaultKeyBindings()
	shortHelp := kb.ShortHelp()

	if len(shortHelp) == 0 {
		t.Error("ShortHelp가 비어있습니다")
	}

	// ShortHelp는 주요 키만 포함해야 함
	expectedLength := 6 // NavigateUp, NavigateDown, KillProcess, Search, ToggleDockerOnly, Quit
	if len(shortHelp) != expectedLength {
		t.Errorf("ShortHelp 길이가 %d이어야 함: got=%d", expectedLength, len(shortHelp))
	}
}

// TestKeyBindings_FullHelp는 FullHelp 메서드를 테스트합니다.
func TestKeyBindings_FullHelp(t *testing.T) {
	kb := DefaultKeyBindings()
	fullHelp := kb.FullHelp()

	if len(fullHelp) == 0 {
		t.Error("FullHelp가 비어있습니다")
	}

	// FullHelp는 ShortHelp보다 길어야 함
	shortHelp := kb.ShortHelp()
	if len(fullHelp) <= len(shortHelp) {
		t.Errorf("FullHelp는 ShortHelp보다 길어야 함: FullHelp=%d, ShortHelp=%d",
			len(fullHelp), len(shortHelp))
	}
}

// TestKeyBindings_VimStyle은 vim 스타일 키 바인딩을 테스트합니다.
func TestKeyBindings_VimStyle(t *testing.T) {
	kb := DefaultKeyBindings()

	// j/k로 위아래 이동
	if !containsKey(kb.NavigateUp, "k") {
		t.Error("NavigateUp에 'k' 키가 포함되어 있지 않습니다")
	}
	if !containsKey(kb.NavigateUp, "up") {
		t.Error("NavigateUp에 'up' 키가 포함되어 있지 않습니다")
	}

	if !containsKey(kb.NavigateDown, "j") {
		t.Error("NavigateDown에 'j' 키가 포함되어 있지 않습니다")
	}
	if !containsKey(kb.NavigateDown, "down") {
		t.Error("NavigateDown에 'down' 키가 포함되어 있지 않습니다")
	}

	// gg로 맨 위, G로 맨 아래
	if !containsKey(kb.NavigateTop, "g") {
		t.Error("NavigateTop에 'g' 키가 포함되어 있지 않습니다")
	}

	if !containsKey(kb.NavigateBottom, "G") {
		t.Error("NavigateBottom에 'G' 키가 포함되어 있지 않습니다")
	}
}

// TestKeyBindings_SpecialKeys는 특수 키 조합을 테스트합니다.
func TestKeyBindings_SpecialKeys(t *testing.T) {
	kb := DefaultKeyBindings()

	// Ctrl+Q로 종료
	if !containsKey(kb.Quit, "ctrl+q") {
		t.Error("Quit에 'ctrl+q' 키가 포함되어 있지 않습니다")
	}

	// Ctrl+R로 새로고침
	if !containsKey(kb.Refresh, "ctrl+r") {
		t.Error("Refresh에 'ctrl+r' 키가 포함되어 있지 않습니다")
	}

	// ESC로 취소
	if !containsKey(kb.Cancel, "esc") {
		t.Error("Cancel에 'esc' 키가 포함되어 있지 않습니다")
	}
}

// TestKeyBindings_ConfirmCancel는 확인/취소 키를 테스트합니다.
func TestKeyBindings_ConfirmCancel(t *testing.T) {
	kb := DefaultKeyBindings()

	// 확인: y 또는 Y
	if !containsKey(kb.Confirm, "y") {
		t.Error("Confirm에 'y' 키가 포함되어 있지 않습니다")
	}
	if !containsKey(kb.Confirm, "Y") {
		t.Error("Confirm에 'Y' 키가 포함되어 있지 않습니다")
	}

	// 취소: n, N, ESC
	if len(kb.Cancel.keys) < 3 {
		t.Errorf("Cancel에 최소 3개 키가 있어야 함: got=%d", len(kb.Cancel.keys))
	}
}

// TestJoinStrings는 joinStrings 헬퍼 함수를 테스트합니다.
func TestJoinStrings(t *testing.T) {
	tests := []struct {
		name     string
		strs     []string
		sep      string
		expected string
	}{
		{
			name:     "빈 슬라이스",
			strs:     []string{},
			sep:      ",",
			expected: "",
		},
		{
			name:     "단일 요소",
			strs:     []string{"a"},
			sep:      ",",
			expected: "a",
		},
		{
			name:     "여러 요소",
			strs:     []string{"a", "b", "c"},
			sep:      ",",
			expected: "a,b,c",
		},
		{
			name:     "다른 구분자",
			strs:     []string{"a", "b", "c"},
			sep:      " | ",
			expected: "a | b | c",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := joinStrings(tt.strs, tt.sep)
			if result != tt.expected {
				t.Errorf("joinStrings() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestKeyBindings_HelpMessages는 도움말 메시지를 테스트합니다.
func TestKeyBindings_HelpMessages(t *testing.T) {
	kb := DefaultKeyBindings()

	tests := []struct {
		name         string
		binding      KeyBinding
		expectedHelp string
	}{
		{"Quit", kb.Quit, "종료"},
		{"NavigateUp", kb.NavigateUp, "위로"},
		{"NavigateDown", kb.NavigateDown, "아래로"},
		{"KillProcess", kb.KillProcess, "프로세스 종료"},
		{"Search", kb.Search, "검색"},
		{"ShowHelp", kb.ShowHelp, "도움말"},
		{"ShowHistory", kb.ShowHistory, "히스토리"},
		{"Refresh", kb.Refresh, "새고침"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			help := tt.binding.Help()
			if help.Desc != tt.expectedHelp {
				t.Errorf("Help 메시지가 예상과 다릅니다: got=%v, want=%v",
					help.Desc, tt.expectedHelp)
			}
		})
	}
}

// TestNewBinding은 NewBinding 함수를 테스트합니다.
func TestNewBinding(t *testing.T) {
	kb := NewBinding(
		WithKeys("a", "b"),
		WithHelp("k", "description"),
	)

	if len(kb.keys) != 2 {
		t.Errorf("keys 길이 = %d, want 2", len(kb.keys))
	}

	if kb.help.key != "k" {
		t.Errorf("help.key = %s, want k", kb.help.key)
	}

	if kb.help.desc != "description" {
		t.Errorf("help.desc = %s, want description", kb.help.desc)
	}
}

// Helper 함수
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		len(s) > len(substr) && (s[:len(substr)] == substr ||
		s[len(s)-len(substr):] == substr ||
		containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func containsKey(kb KeyBinding, key string) bool {
	for _, k := range kb.keys {
		if k == key {
			return true
		}
	}
	return false
}
