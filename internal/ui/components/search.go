// Package components는 검색 컴포넌트를 제공합니다.
package components

import (
	"strings"

	"github.com/manson/port-chaser/internal/ui"
)

// Search는 검색 입력 컴포넌트입니다.
type Search struct {
	styles   *ui.Styles
	query    string
	focused  bool
	cursor   int
}

// NewSearch는 새로운 Search를 생성합니다.
func NewSearch(styles *ui.Styles) *Search {
	return &Search{
		styles:  styles,
		query:   "",
		focused: false,
		cursor:  0,
	}
}

// Render는 검색 입력을 렌더링합니다.
func (s *Search) Render(width int) string {
	// 프롬프트
	prompt := s.styles.SearchPrompt.Render("🔍 /")

	// 입력 필드
	input := s.renderInput(width - len(prompt))

	// 결과 수
	resultText := s.styles.StatusDim.Render(fmt.Sprintf("[%d개 결과]", s.resultCount))

	return prompt + " " + input + " " + resultText
}

// renderInput은 입력 필드를 렌더링합니다.
func (s *Search) renderInput(width int) string {
	if s.query == "" {
		// 플레이스홀더
		placeholder := s.styles.StatusDim.Render("검색어 입력...")
		return placeholder
	}

	// 쿼리 텍스트
	queryText := s.query

	// 커서 위치
	cursorStr := ""
	if s.focused && s.cursor >= 0 && s.cursor <= len(queryText) {
		before := queryText[:s.cursor]
		after := queryText[s.cursor:]
		cursorStr = before + "▏" + after
	} else {
		cursorStr = queryText
	}

	return s.styles.SearchInput.Render(cursorStr)
}

// SetQuery는 검색어를 설정합니다.
func (s *Search) SetQuery(query string) {
	s.query = query
	s.cursor = len(query)
}

// GetQuery는 검색어를 반환합니다.
func (s *Search) GetQuery() string {
	return s.query
}

// AppendChar는 문자를 추가합니다.
func (s *Search) AppendChar(char string) {
	s.query = s.query[:s.cursor] + char + s.query[s.cursor:]
	s.cursor++
}

// DeleteChar는 문자를 삭제합니다.
func (s *Search) DeleteChar() {
	if s.cursor > 0 && len(s.query) > 0 {
		s.query = s.query[:s.cursor-1] + s.query[s.cursor:]
		s.cursor--
	}
}

// MoveCursor는 커서를 이동합니다.
func (s *Search) MoveCursor(delta int) {
	newCursor := s.cursor + delta
	if newCursor >= 0 && newCursor <= len(s.query) {
		s.cursor = newCursor
	}
}

// SetFocused는 포커스 상태를 설정합니다.
func (s *Search) SetFocused(focused bool) {
	s.focused = focused
}

// Clear는 검색어를 지웁니다.
func (s *Search) Clear() {
	s.query = ""
	s.cursor = 0
}

// resultCount는 결과 수를 계산합니다.
func (s *Search) resultCount() int {
	// 실제 구현에서는 전달받은 필터링 결과 수를 사용
	return 0
}

// Placeholder는 다른 패키지에서 사용할 수 있는 구조체입니다.
type Placeholder struct {
	styles   *ui.Styles
	query    string
	resultCount int
}

// NewPlaceholder는 새로운 Placeholder를 생성합니다.
func NewPlaceholder(styles *ui.Styles) *Placeholder {
	return &Placeholder{
		styles:   styles,
		query:    "",
		resultCount: 0,
	}
}

// SetQuery는 검색어를 설정합니다.
func (p *Placeholder) SetQuery(query string) {
	p.query = query
}

// SetResultCount는 결과 수를 설정합니다.
func (p *Placeholder) SetResultCount(count int) {
	p.resultCount = count
}

// Render는 검색 바를 렌더링합니다.
func (p *Placeholder) Render(width int) string {
	prompt := p.styles.SearchPrompt.Render("/")
	input := p.styles.SearchInput.Render(p.query)
	result := p.styles.StatusDim.Render(fmt.Sprintf("[%d개]", p.resultCount))

	return prompt + " " + input + " " + result
}
