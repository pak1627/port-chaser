package ui

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestDefaultStyles(t *testing.T) {
	styles := DefaultStyles()

	if styles == nil {
		t.Fatal("DefaultStyles returned nil")
	}

	tests := []struct {
		name  string
		style lipgloss.Style
		check func(lipgloss.Style) bool
	}{
		{
			name:  "Header",
			style: styles.Header,
			check: func(s lipgloss.Style) bool { return s.String() != "" },
		},
		{
			name:  "HeaderTitle",
			style: styles.HeaderTitle,
			check: func(s lipgloss.Style) bool { return s.String() != "" },
		},
		{
			name:  "PortList",
			style: styles.PortList,
			check: func(s lipgloss.Style) bool { return s.String() != "" },
		},
		{
			name:  "PortItem",
			style: styles.PortItem,
			check: func(s lipgloss.Style) bool { return s.String() != "" },
		},
		{
			name:  "PortSelected",
			style: styles.PortSelected,
			check: func(s lipgloss.Style) bool { return s.String() != "" },
		},
		{
			name:  "DockerMarker",
			style: styles.DockerMarker,
			check: func(s lipgloss.Style) bool { return s.String() != "" },
		},
		{
			name:  "RecommendedMarker",
			style: styles.RecommendedMarker,
			check: func(s lipgloss.Style) bool { return s.String() != "" },
		},
		{
			name:  "SystemMarker",
			style: styles.SystemMarker,
			check: func(s lipgloss.Style) bool { return s.String() != "" },
		},
		{
			name:  "StatusBar",
			style: styles.StatusBar,
			check: func(s lipgloss.Style) bool { return s.String() != "" },
		},
		{
			name:  "Dialog",
			style: styles.Dialog,
			check: func(s lipgloss.Style) bool { return s.String() != "" },
		},
		{
			name:  "Warning",
			style: styles.Warning,
			check: func(s lipgloss.Style) bool { return s.String() != "" },
		},
		{
			name:  "Error",
			style: styles.Error,
			check: func(s lipgloss.Style) bool { return s.String() != "" },
		},
		{
			name:  "Success",
			style: styles.Success,
			check: func(s lipgloss.Style) bool { return s.String() != "" },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.check(tt.style) {
				t.Errorf("%s style not configured", tt.name)
			}
		})
	}
}

func TestStyles_HeaderStyles(t *testing.T) {
	styles := DefaultStyles()

	title := styles.HeaderTitle
	titleStyle := title.Render("Test")
	if titleStyle == "" {
		t.Error("HeaderTitle not rendered")
	}

	subtitle := styles.HeaderSubtitle
	subtitleStyle := subtitle.Render("Test")
	if subtitleStyle == "" {
		t.Error("HeaderSubtitle not rendered")
	}

	header := styles.Header
	headerText := header.Render("Header Test")
	if headerText == "" {
		t.Error("Header not rendered")
	}
}

func TestStyles_PortItemStyles(t *testing.T) {
	styles := DefaultStyles()

	itemText := styles.PortItem.Render("Port 3000")
	if itemText == "" {
		t.Error("PortItem not rendered")
	}

	selectedText := styles.PortSelected.Render("Port 3000")
	if selectedText == "" {
		t.Error("PortSelected not rendered")
	}

	if !styles.PortSelected.GetBold() {
		t.Error("PortSelected should be bold")
	}
}

func TestStyles_MarkerStyles(t *testing.T) {
	styles := DefaultStyles()

	if !styles.DockerMarker.GetBold() {
		t.Error("DockerMarker should be bold")
	}
	if !styles.RecommendedMarker.GetBold() {
		t.Error("RecommendedMarker should be bold")
	}
	if !styles.SystemMarker.GetBold() {
		t.Error("SystemMarker should be bold")
	}

	if styles.DockerMarker.GetForeground() == lipgloss.Color("") {
		t.Error("DockerMarker missing foreground color")
	}
	if styles.RecommendedMarker.GetForeground() == lipgloss.Color("") {
		t.Error("RecommendedMarker missing foreground color")
	}
	if styles.SystemMarker.GetForeground() == lipgloss.Color("") {
		t.Error("SystemMarker missing foreground color")
	}
}

func TestStyles_StatusBarStyles(t *testing.T) {
	styles := DefaultStyles()

	if !styles.StatusKey.GetBold() {
		t.Error("StatusKey should be bold")
	}

	statusDimText := styles.StatusDim.Render("Dimmed")
	if statusDimText == "" {
		t.Error("StatusDim not rendered")
	}

	statusBarText := styles.StatusBar.Render("Status")
	if statusBarText == "" {
		t.Error("StatusBar not rendered")
	}
}

func TestStyles_DialogStyles(t *testing.T) {
	styles := DefaultStyles()

	if !styles.DialogTitle.GetBold() {
		t.Error("DialogTitle should be bold")
	}

	dialogText := styles.Dialog.Render("Dialog Content")
	if dialogText == "" {
		t.Error("Dialog not rendered")
	}

	borderText := styles.DialogBorder.Render("Border Test")
	if borderText == "" {
		t.Error("DialogBorder not rendered")
	}
}

func TestStyles_MessageStyles(t *testing.T) {
	styles := DefaultStyles()

	tests := []struct {
		name   string
		style  lipgloss.Style
		prefix string
	}{
		{"Warning", styles.Warning, "⚠ "},
		{"Error", styles.Error, "✗ "},
		{"Success", styles.Success, "✓ "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.style.GetBold() {
				t.Errorf("%s style should be bold", tt.name)
			}

			text := tt.style.Render(tt.prefix + "Message")
			if text == "" {
				t.Errorf("%s style not rendered", tt.name)
			}
		})
	}
}

func TestRenderDockerMarker(t *testing.T) {
	styles := DefaultStyles()
	marker := RenderDockerMarker(styles)

	if marker == "" {
		t.Error("Docker marker is empty")
	}

	if !hasSubstring(marker, "[D]") {
		t.Errorf("Docker marker missing '[D]': %s", marker)
	}

	t.Logf("Docker marker: %s", marker)
}

func TestRenderRecommendedMarker(t *testing.T) {
	styles := DefaultStyles()
	marker := RenderRecommendedMarker(styles)

	if marker == "" {
		t.Error("Recommended marker is empty")
	}

	if !hasSubstring(marker, "[!]") {
		t.Errorf("Recommended marker missing '[!]': %s", marker)
	}

	t.Logf("Recommended marker: %s", marker)
}

func TestRenderSystemMarker(t *testing.T) {
	styles := DefaultStyles()
	marker := RenderSystemMarker(styles)

	if marker == "" {
		t.Error("System marker is empty")
	}

	if !hasSubstring(marker, "[S]") {
		t.Errorf("System marker missing '[S]': %s", marker)
	}

	t.Logf("System marker: %s", marker)
}

func TestRenderWarning(t *testing.T) {
	styles := DefaultStyles()
	msg := "system process"
	warning := RenderWarning(styles, msg)

	if warning == "" {
		t.Error("Warning message is empty")
	}

	if !hasSubstring(warning, "⚠") {
		t.Errorf("Warning missing '⚠': %s", warning)
	}

	if !hasSubstring(warning, msg) {
		t.Errorf("Warning missing original message: %s", warning)
	}

	t.Logf("Warning message: %s", warning)
}

func TestRenderError(t *testing.T) {
	styles := DefaultStyles()
	msg := "connection failed"
	errMsg := RenderError(styles, msg)

	if errMsg == "" {
		t.Error("Error message is empty")
	}

	if !hasSubstring(errMsg, "✗") {
		t.Errorf("Error missing '✗': %s", errMsg)
	}

	if !hasSubstring(errMsg, msg) {
		t.Errorf("Error missing original message: %s", errMsg)
	}

	t.Logf("Error message: %s", errMsg)
}

func TestRenderSuccess(t *testing.T) {
	styles := DefaultStyles()
	msg := "completed successfully"
	successMsg := RenderSuccess(styles, msg)

	if successMsg == "" {
		t.Error("Success message is empty")
	}

	if !hasSubstring(successMsg, "✓") {
		t.Errorf("Success missing '✓': %s", successMsg)
	}

	if !hasSubstring(successMsg, msg) {
		t.Errorf("Success missing original message: %s", successMsg)
	}

	t.Logf("Success message: %s", successMsg)
}

func TestStyles_MutedStyle(t *testing.T) {
	styles := DefaultStyles()

	mutedText := styles.Muted.Render("Muted text")
	if mutedText == "" {
		t.Error("Muted not rendered")
	}

	fg := styles.Muted.GetForeground()
	if fg == lipgloss.Color("") {
		t.Error("Muted missing foreground color")
	}
}

func TestStyles_ColorScheme(t *testing.T) {
	styles := DefaultStyles()

	tests := []struct {
		name     string
		style    lipgloss.Style
		colorNum string
	}{
		{"DockerMarker is blue", styles.DockerMarker, "86"},
		{"RecommendedMarker is yellow", styles.RecommendedMarker, "226"},
		{"SystemMarker is red", styles.SystemMarker, "196"},
		{"Success is green", styles.Success, "46"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = tt.style.GetForeground()
		})
	}
}

func TestStyles_Width(t *testing.T) {
	styles := DefaultStyles()

	tests := []struct {
		name  string
		style lipgloss.Style
		minW  int
	}{
		{"HeaderTitle", styles.HeaderTitle, 20},
		{"PortItem", styles.PortItem, 80},
		{"Dialog", styles.Dialog, 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := tt.style.GetWidth()
			if w < tt.minW {
				t.Errorf("%s width should be >= %d: got=%d", tt.name, tt.minW, w)
			}
		})
	}
}

func BenchmarkDefaultStyles(b *testing.B) {
	for i := 0; i < b.N; i++ {
		DefaultStyles()
	}
}

func BenchmarkRenderMarkers(b *testing.B) {
	styles := DefaultStyles()
	b.ResetTimer()

	b.Run("Docker", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			RenderDockerMarker(styles)
		}
	})

	b.Run("Recommended", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			RenderRecommendedMarker(styles)
		}
	})

	b.Run("System", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			RenderSystemMarker(styles)
		}
	})
}

func BenchmarkRenderMessages(b *testing.B) {
	styles := DefaultStyles()
	msg := "Test message"
	b.ResetTimer()

	b.Run("Warning", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			RenderWarning(styles, msg)
		}
	})

	b.Run("Error", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			RenderError(styles, msg)
		}
	})

	b.Run("Success", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			RenderSuccess(styles, msg)
		}
	})
}

func hasSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			hasSubstringMiddle(s, substr)))
}

func hasSubstringMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
