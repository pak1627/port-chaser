package ui

import (
	"testing"
)

func TestDefaultKeyBindings(t *testing.T) {
	kb := DefaultKeyBindings()

	if kb == nil {
		t.Fatal("DefaultKeyBindings returned nil")
	}

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
				t.Errorf("%s binding missing Help Key", tt.name)
			}
			if help.Desc == "" {
				t.Errorf("%s binding missing Help Desc", tt.name)
			}
		})
	}
}

func TestKeyBindings_HelpText(t *testing.T) {
	kb := DefaultKeyBindings()
	text := kb.HelpText()

	if text == "" {
		t.Error("HelpText is empty")
	}

	expectedKeys := []string{"↑/k", "↓/j", "enter", "q"}
	for _, expectedKey := range expectedKeys {
		if !contains(text, expectedKey) {
			t.Errorf("HelpText missing '%s': %s", expectedKey, text)
		}
	}

	t.Logf("HelpText: %s", text)
}

func TestKeyBindings_ShortHelp(t *testing.T) {
	kb := DefaultKeyBindings()
	shortHelp := kb.ShortHelp()

	if len(shortHelp) == 0 {
		t.Error("ShortHelp is empty")
	}

	expectedLength := 6
	if len(shortHelp) != expectedLength {
		t.Errorf("ShortHelp length should be %d: got=%d", expectedLength, len(shortHelp))
	}
}

func TestKeyBindings_FullHelp(t *testing.T) {
	kb := DefaultKeyBindings()
	fullHelp := kb.FullHelp()

	if len(fullHelp) == 0 {
		t.Error("FullHelp is empty")
	}

	shortHelp := kb.ShortHelp()
	if len(fullHelp) <= len(shortHelp) {
		t.Errorf("FullHelp should be longer than ShortHelp: FullHelp=%d, ShortHelp=%d",
			len(fullHelp), len(shortHelp))
	}
}

func TestKeyBindings_VimStyle(t *testing.T) {
	kb := DefaultKeyBindings()

	if !containsKey(kb.NavigateUp, "k") {
		t.Error("NavigateUp missing 'k' key")
	}
	if !containsKey(kb.NavigateUp, "up") {
		t.Error("NavigateUp missing 'up' key")
	}

	if !containsKey(kb.NavigateDown, "j") {
		t.Error("NavigateDown missing 'j' key")
	}
	if !containsKey(kb.NavigateDown, "down") {
		t.Error("NavigateDown missing 'down' key")
	}

	if !containsKey(kb.NavigateTop, "g") {
		t.Error("NavigateTop missing 'g' key")
	}

	if !containsKey(kb.NavigateBottom, "G") {
		t.Error("NavigateBottom missing 'G' key")
	}
}

func TestKeyBindings_SpecialKeys(t *testing.T) {
	kb := DefaultKeyBindings()

	if !containsKey(kb.Quit, "ctrl+q") {
		t.Error("Quit missing 'ctrl+q' key")
	}

	if !containsKey(kb.Refresh, "ctrl+r") {
		t.Error("Refresh missing 'ctrl+r' key")
	}

	if !containsKey(kb.Cancel, "esc") {
		t.Error("Cancel missing 'esc' key")
	}
}

func TestKeyBindings_ConfirmCancel(t *testing.T) {
	kb := DefaultKeyBindings()

	if !containsKey(kb.Confirm, "y") {
		t.Error("Confirm missing 'y' key")
	}
	if !containsKey(kb.Confirm, "Y") {
		t.Error("Confirm missing 'Y' key")
	}

	if len(kb.Cancel.keys) < 3 {
		t.Errorf("Cancel should have at least 3 keys: got=%d", len(kb.Cancel.keys))
	}
}

func TestJoinStrings(t *testing.T) {
	tests := []struct {
		name     string
		strs     []string
		sep      string
		expected string
	}{
		{
			name:     "empty slice",
			strs:     []string{},
			sep:      ",",
			expected: "",
		},
		{
			name:     "single element",
			strs:     []string{"a"},
			sep:      ",",
			expected: "a",
		},
		{
			name:     "multiple elements",
			strs:     []string{"a", "b", "c"},
			sep:      ",",
			expected: "a,b,c",
		},
		{
			name:     "different separator",
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

func TestKeyBindings_HelpMessages(t *testing.T) {
	kb := DefaultKeyBindings()

	tests := []struct {
		name         string
		binding      KeyBinding
		expectedHelp string
	}{
		{"Quit", kb.Quit, "quit"},
		{"NavigateUp", kb.NavigateUp, "up"},
		{"NavigateDown", kb.NavigateDown, "down"},
		{"KillProcess", kb.KillProcess, "kill process"},
		{"Search", kb.Search, "search"},
		{"ShowHelp", kb.ShowHelp, "help"},
		{"ShowHistory", kb.ShowHistory, "history"},
		{"Refresh", kb.Refresh, "refresh"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			help := tt.binding.Help()
			if help.Desc != tt.expectedHelp {
				t.Errorf("Help message differs: got=%v, want=%v",
					help.Desc, tt.expectedHelp)
			}
		})
	}
}

func TestNewBinding(t *testing.T) {
	kb := NewBinding(
		WithKeys("a", "b"),
		WithHelp("k", "description"),
	)

	if len(kb.keys) != 2 {
		t.Errorf("keys length = %d, want 2", len(kb.keys))
	}

	if kb.help.key != "k" {
		t.Errorf("help.key = %s, want k", kb.help.key)
	}

	if kb.help.desc != "description" {
		t.Errorf("help.desc = %s, want description", kb.help.desc)
	}
}

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
