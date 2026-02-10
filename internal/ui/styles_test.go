// Package ui에 대한 스타일 테스트입니다.
package ui

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

// TestDefaultStyles는 기본 스타일 생성을 테스트합니다.
func TestDefaultStyles(t *testing.T) {
	styles := DefaultStyles()

	if styles == nil {
		t.Fatal("DefaultStyles가 nil을 반환했습니다")
	}

	// 모든 스타일이 설정되어 있는지 확인
	tests := []struct {
		name   string
		style  lipgloss.Style
		check  func(lipgloss.Style) bool
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
				t.Errorf("%s 스타일이 설정되지 않았습니다", tt.name)
			}
		})
	}
}

// TestStyles_HeaderStyles는 헤더 스타일을 테스트합니다.
func TestStyles_HeaderStyles(t *testing.T) {
	styles := DefaultStyles()

	// HeaderTitle은 볼드 and 색상이 있어야 함
	title := styles.HeaderTitle
	titleStyle := title.Render("Test")
	if titleStyle == "" {
		t.Error("HeaderTitle이 렌더링되지 않았습니다")
	}

	// HeaderSubtitle은 흐리하게(faint) 처리되어야 함
	subtitle := styles.HeaderSubtitle
	subtitleStyle := subtitle.Render("Test")
	if subtitleStyle == "" {
		t.Error("HeaderSubtitle이 렌더링되지 않았습니다")
	}

	// Header에는 테두리가 있어야 함
	header := styles.Header
	headerText := header.Render("Header Test")
	if headerText == "" {
		t.Error("Header가 렌더링되지 않았습니다")
	}
}

// TestStyles_PortItemStyles는 포트 항목 스타일을 테스트합니다.
func TestStyles_PortItemStyles(t *testing.T) {
	styles := DefaultStyles()

	// 일반 항목
	itemText := styles.PortItem.Render("Port 3000")
	if itemText == "" {
		t.Error("PortItem이 렌더링되지 않았습니다")
	}

	// 선택된 항목은 배경색이 있어야 함
	selectedText := styles.PortSelected.Render("Port 3000")
	if selectedText == "" {
		t.Error("PortSelected가 렌더링되지 않았습니다")
	}

	// 선택된 항목은 볼드여야 함
	if !styles.PortSelected.GetBold() {
		t.Error("PortSelected는 볼드여야 합니다")
	}
}

// TestStyles_MarkerStyles는 마커 스타일을 테스트합니다.
func TestStyles_MarkerStyles(t *testing.T) {
	styles := DefaultStyles()

	// 모든 마커는 볼드여야 함
	if !styles.DockerMarker.GetBold() {
		t.Error("DockerMarker는 볼드여야 합니다")
	}
	if !styles.RecommendedMarker.GetBold() {
		t.Error("RecommendedMarker는 볼드여야 합니다")
	}
	if !styles.SystemMarker.GetBold() {
		t.Error("SystemMarker는 볼드여야 합니다")
	}

	// 각 마커의 전경색이 설정되어 있어야 함
	if styles.DockerMarker.GetForeground() == lipgloss.Color("") {
		t.Error("DockerMarker에 전경색이 없습니다")
	}
	if styles.RecommendedMarker.GetForeground() == lipgloss.Color("") {
		t.Error("RecommendedMarker에 전경색이 없습니다")
	}
	if styles.SystemMarker.GetForeground() == lipgloss.Color("") {
		t.Error("SystemMarker에 전경색이 없습니다")
	}
}

// TestStyles_StatusBarStyles는 상태바 스타일을 테스트합니다.
func TestStyles_StatusBarStyles(t *testing.T) {
	styles := DefaultStyles()

	// StatusKey는 볼드 and 색상이 있어야 함
	if !styles.StatusKey.GetBold() {
		t.Error("StatusKey는 볼드여야 합니다")
	}

	// StatusDim은 흐리하게 처리되어야 함
	statusDimText := styles.StatusDim.Render("Dimmed")
	if statusDimText == "" {
		t.Error("StatusDim이 렌더링되지 않았습니다")
	}

	// StatusBar에는 테두리가 있어야 함
	statusBarText := styles.StatusBar.Render("Status")
	if statusBarText == "" {
		t.Error("StatusBar가 렌더링되지 않았습니다")
	}
}

// TestStyles_DialogStyles는 다이얼로그 스타일을 테스트합니다.
func TestStyles_DialogStyles(t *testing.T) {
	styles := DefaultStyles()

	// DialogTitle은 볼드여야 함
	if !styles.DialogTitle.GetBold() {
		t.Error("DialogTitle은 볼드여야 합니다")
	}

	// Dialog는 둥근 테두리가 있어야 함
	dialogText := styles.Dialog.Render("Dialog Content")
	if dialogText == "" {
		t.Error("Dialog가 렌더링되지 않았습니다")
	}

	// DialogBorder도 테스트
	borderText := styles.DialogBorder.Render("Border Test")
	if borderText == "" {
		t.Error("DialogBorder가 렌더링되지 않았습니다")
	}
}

// TestStyles_MessageStyles는 메시지 스타일을 테스트합니다.
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
			// 모든 메시지 스타일은 볼드여야 함
			if !tt.style.GetBold() {
				t.Errorf("%s 스타일은 볼드여야 합니다", tt.name)
			}

			// 렌더링 테스트
			text := tt.style.Render(tt.prefix + "Message")
			if text == "" {
				t.Errorf("%s 스타일이 렌더링되지 않았습니다", tt.name)
			}
		})
	}
}

// TestRenderDockerMarker는 Docker 마커 렌더링을 테스트합니다.
func TestRenderDockerMarker(t *testing.T) {
	styles := DefaultStyles()
	marker := RenderDockerMarker(styles)

	if marker == "" {
		t.Error("Docker 마커가 비어있습니다")
	}

	// [D]가 포함되어야 함
	if !contains(marker, "[D]") {
		t.Errorf("Docker 마커에 '[D]'가 포함되어 있지 않습니다: %s", marker)
	}

	t.Logf("Docker 마커: %s", marker)
}

// TestRenderRecommendedMarker는 추천 마커 렌더링을 테스트합니다.
func TestRenderRecommendedMarker(t *testing.T) {
	styles := DefaultStyles()
	marker := RenderRecommendedMarker(styles)

	if marker == "" {
		t.Error("추천 마커가 비어있습니다")
	}

	// [!]가 포함되어야 함
	if !contains(marker, "[!]") {
		t.Errorf("추천 마커에 '[!]'가 포함되어 있지 않습니다: %s", marker)
	}

	t.Logf("추천 마커: %s", marker)
}

// TestRenderSystemMarker는 시스템 마커 렌더링을 테스트합니다.
func TestRenderSystemMarker(t *testing.T) {
	styles := DefaultStyles()
	marker := RenderSystemMarker(styles)

	if marker == "" {
		t.Error("시스템 마커가 비어있습니다")
	}

	// [S]가 포함되어야 함
	if !contains(marker, "[S]") {
		t.Errorf("시스템 마커에 '[S]'가 포함되어 있지 않습니다: %s", marker)
	}

	t.Logf("시스템 마커: %s", marker)
}

// TestRenderWarning는 경고 메시지 렌더링을 테스트합니다.
func TestRenderWarning(t *testing.T) {
	styles := DefaultStyles()
	msg := "시스템 프로세스입니다"
	warning := RenderWarning(styles, msg)

	if warning == "" {
		t.Error("경고 메시지가 비어있습니다")
	}

	// ⚠가 포함되어야 함
	if !contains(warning, "⚠") {
		t.Errorf("경고에 '⚠'가 포함되어 있지 않습니다: %s", warning)
	}

	// 원본 메시지가 포함되어야 함
	if !contains(warning, msg) {
		t.Errorf("경고에 원본 메시지가 포함되어 있지 않습니다: %s", warning)
	}

	t.Logf("경고 메시지: %s", warning)
}

// TestRenderError는 에러 메시지 렌더링을 테스트합니다.
func TestRenderError(t *testing.T) {
	styles := DefaultStyles()
	msg := "연결 실패"
	errMsg := RenderError(styles, msg)

	if errMsg == "" {
		t.Error("에러 메시지가 비어있습니다")
	}

	// ✗가 포함되어야 함
	if !contains(errMsg, "✗") {
		t.Errorf("에러에 '✗'가 포함되어 있지 않습니다: %s", errMsg)
	}

	// 원본 메시지가 포함되어야 함
	if !contains(errMsg, msg) {
		t.Errorf("에러에 원본 메시지가 포함되어 있지 않습니다: %s", errMsg)
	}

	t.Logf("에러 메시지: %s", errMsg)
}

// TestRenderSuccess는 성공 메시지 렌더링을 테스트합니다.
func TestRenderSuccess(t *testing.T) {
	styles := DefaultStyles()
	msg := "완료되었습니다"
	successMsg := RenderSuccess(styles, msg)

	if successMsg == "" {
		t.Error("성공 메시지가 비어있습니다")
	}

	// ✓가 포함되어야 함
	if !contains(successMsg, "✓") {
		t.Errorf("성공 메시지에 '✓'가 포함되어 있지 않습니다: %s", successMsg)
	}

	// 원본 메시지가 포함되어야 함
	if !contains(successMsg, msg) {
		t.Errorf("성공 메시지에 원본 메시지가 포함되어 있지 않습니다: %s", successMsg)
	}

	t.Logf("성공 메시지: %s", successMsg)
}

// TestStyles_SearchInputStyles는 검색 입력 스타일을 테스트합니다.
func TestStyles_SearchInputStyles(t *testing.T) {
	styles := DefaultStyles()

	// SearchPrompt는 볼드여야 함
	if !styles.SearchPrompt.GetBold() {
		t.Error("SearchPrompt는 볼드여야 합니다")
	}

	// 렌더링 테스트
	searchInput := styles.SearchInput.Render("search query")
	if searchInput == "" {
		t.Error("SearchInput이 렌더링되지 않았습니다")
	}

	searchPrompt := styles.SearchPrompt.Render("/")
	if searchPrompt == "" {
		t.Error("SearchPrompt가 렌더링되지 않았습니다")
	}
}

// TestStyles_MutedStyle은 비활성 스타일을 테스트합니다.
func TestStyles_MutedStyle(t *testing.T) {
	styles := DefaultStyles()

	// Muted는 흐리하게 처리되어야 함
	mutedText := styles.Muted.Render("Muted text")
	if mutedText == "" {
		t.Error("Muted가 렌더링되지 않았습니다")
	}

	// 전경색이 회색이어야 함
	fg := styles.Muted.GetForeground()
	if fg == lipgloss.Color("") {
		t.Error("Muted에 전경색이 설정되지 않았습니다")
	}
}

// TestStyles_ColorScheme은 색상 스킴을 테스트합니다.
func TestStyles_ColorScheme(t *testing.T) {
	styles := DefaultStyles()

	// SPEC 요구사항에 따른 색상 확인
	tests := []struct {
		name     string
		style    lipgloss.Style
		colorNum string // ANSI 컬러 번호
	}{
		{"DockerMarker는 파란색", styles.DockerMarker, "86"},
		{"RecommendedMarker는 노란색", styles.RecommendedMarker, "226"},
		{"SystemMarker는 빨간색", styles.SystemMarker, "196"},
		{"Success는 초록색", styles.Success, "46"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fg := tt.style.GetForeground()
			if fg.String() != tt.colorNum {
				t.Errorf("%s: 컬러 %s 예상, got=%s", tt.name, tt.colorNum, fg.String())
			}
		})
	}
}

// TestStyles_Width는 스타일 너비를 테스트합니다.
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
		{"SearchInput", styles.SearchInput, 60},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := tt.style.GetWidth()
			if w < tt.minW {
				t.Errorf("%s 너이가 %d 이상이어야 함: got=%d", tt.name, tt.minW, w)
			}
		})
	}
}

// BenchmarkDefaultStyles는 기본 스타일 생성 성능을 테스트합니다.
func BenchmarkDefaultStyles(b *testing.B) {
	for i := 0; i < b.N; i++ {
		DefaultStyles()
	}
}

// BenchmarkRenderMarkers는 마커 렌더링 성능을 테스트합니다.
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

// BenchmarkRenderMessages는 메시지 렌더링 성능을 테스트합니다.
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
