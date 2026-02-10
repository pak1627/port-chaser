// Package components에 대한 테스트
package components

import (
	"strings"
	"testing"

	"github.com/manson/port-chaser/internal/models"
	"github.com/manson/port-chaser/internal/ui"
)

// TestPortList_RenderEmpty는 빈 목록 렌더링을 테스트합니다.
func TestPortList_RenderEmpty(t *testing.T) {
	styles := ui.DefaultStyles()
	pl := NewPortList(styles)

	result := pl.Render([]models.PortInfo{}, -1, 80)

	if !strings.Contains(result, "활성 포트가 없습니다") {
		t.Error("빈 목록 메시지가 포함되어야 합니다")
	}
}

// TestPortList_Render는 포트 목록 렌더링을 테스트합니다.
func TestPortList_Render(t *testing.T) {
	styles := ui.DefaultStyles()
	pl := NewPortList(styles)

	ports := []models.PortInfo{
		{PortNumber: 3000, ProcessName: "node", PID: 1001, User: "user", Command: "npm start"},
		{PortNumber: 8080, ProcessName: "python", PID: 1002, User: "dev", Command: "python app.py"},
	}

	result := pl.Render(ports, 0, 80)

	// 헤더 확인
	if !strings.Contains(result, "포트") {
		t.Error("헤더에 '포트'가 포함되어야 합니다")
	}

	// 포트 번호 확인
	if !strings.Contains(result, "3000") {
		t.Error("포트 3000이 표시되어야 합니다")
	}

	if !strings.Contains(result, "8080") {
		t.Error("포트 8080이 표시되어야 합니다")
	}

	// 프로세스 이름 확인
	if !strings.Contains(result, "node") {
		t.Error("프로세스 'node'가 표시되어야 합니다")
	}

	if !strings.Contains(result, "python") {
		t.Error("프로세스 'python'이 표시되어야 합니다")
	}
}

// TestPortList_RenderDockerMarker는 Docker 마커 렌더링을 테스트합니다.
func TestPortList_RenderDockerMarker(t *testing.T) {
	styles := ui.DefaultStyles()
	pl := NewPortList(styles)

	ports := []models.PortInfo{
		{PortNumber: 3000, ProcessName: "node", PID: 1001, IsDocker: true, ContainerName: "app"},
	}

	result := pl.Render(ports, 0, 80)

	if !strings.Contains(result, "[D]") {
		t.Error("Docker 마커 [D]가 표시되어야 합니다")
	}
}

// TestPortList_RenderRecommendedMarker는 추천 마커 렌더링을 테스트합니다.
func TestPortList_RenderRecommendedMarker(t *testing.T) {
	styles := ui.DefaultStyles()
	pl := NewPortList(styles)

	ports := []models.PortInfo{
		{PortNumber: 3000, ProcessName: "node", PID: 1001, KillCount: 5},
	}

	result := pl.Render(ports, 0, 80)

	if !strings.Contains(result, "[!]") {
		t.Error("추천 마커 [!]가 표시되어야 합니다")
	}
}

// TestPortList_RenderSystemMarker는 시스템 마커 렌더링을 테스트합니다.
func TestPortList_RenderSystemMarker(t *testing.T) {
	styles := ui.DefaultStyles()
	pl := NewPortList(styles)

	ports := []models.PortInfo{
		{PortNumber: 80, ProcessName: "httpd", PID: 50, IsSystem: true},
	}

	result := pl.Render(ports, 0, 80)

	if !strings.Contains(result, "[S]") {
		t.Error("시스템 마커 [S]가 표시되어야 합니다")
	}
}

// TestHeader_Render는 헤더 렌더링을 테스트합니다.
func TestHeader_Render(t *testing.T) {
	styles := ui.DefaultStyles()
	header := NewHeader(styles, "Port Chaser")

	result := header.Render(10, false, 80)

	if !strings.Contains(result, "Port Chaser") {
		t.Error("제목이 표시되어야 합니다")
	}

	if !strings.Contains(result, "10개") {
		t.Error("포트 수가 표시되어야 합니다")
	}
}

// TestHeader_RenderDockerOnly는 Docker 필터 상태를 테스트합니다.
func TestHeader_RenderDockerOnly(t *testing.T) {
	styles := ui.DefaultStyles()
	header := NewHeader(styles, "Port Chaser")

	result := header.Render(5, true, 80)

	if !strings.Contains(result, "[Docker만]") {
		t.Error("Docker 필터 상태가 표시되어야 합니다")
	}
}

// TestStatusBar_Render는 상태바 렌더링을 테스트합니다.
func TestStatusBar_Render(t *testing.T) {
	styles := ui.DefaultStyles()
	bindings := ui.DefaultKeyBindings()
	statusBar := NewStatusBar(styles, bindings)

	result := statusBar.Render(80)

	// 키 가이드 확인
	if !strings.Contains(result, "q") {
		t.Error("종료 키 가이드가 있어야 합니다")
	}
}

// TestStatusBar_SetMessage는 상태 메시지 설정을 테스트합니다.
func TestStatusBar_SetMessage(t *testing.T) {
	styles := ui.DefaultStyles()
	bindings := ui.DefaultKeyBindings()
	statusBar := NewStatusBar(styles, bindings)

	statusBar.SetMessage("테스트 메시지")
	result := statusBar.Render(80)

	if !strings.Contains(result, "테스트 메시지") {
		t.Error("상태 메시지가 표시되어야 합니다")
	}
}

// TestDialog_RenderConfirmKill는 종료 확인 다이얼로그를 테스트합니다.
func TestDialog_RenderConfirmKill(t *testing.T) {
	styles := ui.DefaultStyles()
	dialog := NewDialog(styles)

	port := &models.PortInfo{
		PortNumber:  3000,
		ProcessName: "node",
		PID:         1001,
		Command:     "npm start",
	}

	result := dialog.RenderConfirmKill(port)

	if !strings.Contains(result, "프로세스 종료 확인") {
		t.Error("다이얼로그 제목이 표시되어야 합니다")
	}

	if !strings.Contains(result, "3000") {
		t.Error("포트 번호가 표시되어야 합니다")
	}

	if !strings.Contains(result, "node") {
		t.Error("프로세스 이름이 표시되어야 합니다")
	}

	if !strings.Contains(result, "[y]") {
		t.Error("확인 프롬프트가 있어야 합니다")
	}
}

// TestDialog_RenderConfirmKill_SystemProcess는 시스템 프로세스 경고를 테스트합니다.
func TestDialog_RenderConfirmKill_SystemProcess(t *testing.T) {
	styles := ui.DefaultStyles()
	dialog := NewDialog(styles)

	port := &models.PortInfo{
		PortNumber:  80,
		ProcessName: "httpd",
		PID:         1,
		IsSystem:    true,
	}

	result := dialog.RenderConfirmKill(port)

	if !strings.Contains(result, "시스템 중요 프로세스") {
		t.Error("시스템 프로세스 경고가 표시되어야 합니다")
	}
}

// TestDialog_RenderConfirmKill_Docker는 Docker 포트 다이얼로그를 테스트합니다.
func TestDialog_RenderConfirmKill_Docker(t *testing.T) {
	styles := ui.DefaultStyles()
	dialog := NewDialog(styles)

	port := &models.PortInfo{
		PortNumber:    3000,
		ProcessName:   "node",
		PID:           1001,
		IsDocker:      true,
		ContainerName: "my-app",
		ImageName:     "node:16",
	}

	result := dialog.RenderConfirmKill(port)

	if !strings.Contains(result, "Docker:") {
		t.Error("Docker 정보가 표시되어야 합니다")
	}

	if !strings.Contains(result, "my-app") {
		t.Error("컨테이너 이름이 표시되어야 합니다")
	}
}

// TestDialog_RenderConfirmKill_Nil은 nil 포트 처리를 테스트합니다.
func TestDialog_RenderConfirmKill_Nil(t *testing.T) {
	styles := ui.DefaultStyles()
	dialog := NewDialog(styles)

	result := dialog.RenderConfirmKill(nil)

	if result != "" {
		t.Error("nil 포트는 빈 문자열을 반환해야 합니다")
	}
}

// TestSearch_SetQuery는 검색어 설정을 테스트합니다.
func TestSearch_SetQuery(t *testing.T) {
	styles := ui.DefaultStyles()
	search := NewSearch(styles)

	search.SetQuery("test")

	if search.GetQuery() != "test" {
		t.Errorf("검색어 = %v, want 'test'", search.GetQuery())
	}
}

// TestSearch_Clear는 검색어 지우기를 테스트합니다.
func TestSearch_Clear(t *testing.T) {
	styles := ui.DefaultStyles()
	search := NewSearch(styles)

	search.SetQuery("test")
	search.Clear()

	if search.GetQuery() != "" {
		t.Error("지운 후 검색어가 비어있어야 합니다")
	}
}

// TestSearch_AppendChar는 문자 추가를 테스트합니다.
func TestSearch_AppendChar(t *testing.T) {
	styles := ui.DefaultStyles()
	search := NewSearch(styles)

	search.SetQuery("tes")
	search.AppendChar("t")

	if search.GetQuery() != "test" {
		t.Errorf("문자 추가 후 = %v, want 'test'", search.GetQuery())
	}
}

// TestSearch_DeleteChar는 문자 삭제를 테스트합니다.
func TestSearch_DeleteChar(t *testing.T) {
	styles := ui.DefaultStyles()
	search := NewSearch(styles)

	search.SetQuery("test")
	search.DeleteChar()

	if search.GetQuery() != "tes" {
		t.Errorf("문자 삭제 후 = %v, want 'tes'", search.GetQuery())
	}
}

// TestSearch_MoveCursor는 커서 이동을 테스트합니다.
func TestSearch_MoveCursor(t *testing.T) {
	styles := ui.DefaultStyles()
	search := NewSearch(styles)

	search.SetQuery("test")
	search.MoveCursor(-1)
	if search.cursor != 3 {
		t.Errorf("커서 위치 = %d, want 3", search.cursor)
	}

	search.MoveCursor(2)
	if search.cursor != 4 {
		t.Errorf("커서 위치 = %d, want 4", search.cursor)
	}

	// 경계 테스트
	search.MoveCursor(10) // 범위 초과
	if search.cursor != 4 {
		t.Errorf("커서 위치 = %d, want 4 (최대)", search.cursor)
	}
}

// Benchmark_PortList_Render는 포트 목록 렌더링 성능을 벤치마킹합니다.
func Benchmark_PortList_Render(b *testing.B) {
	styles := ui.DefaultStyles()
	pl := NewPortList(styles)

	// 100개 포트 생성
	ports := make([]models.PortInfo, 100)
	for i := 0; i < 100; i++ {
		ports[i] = models.PortInfo{
			PortNumber:  3000 + i,
			ProcessName: "process",
			PID:         1000 + i,
			User:        "user",
			Command:     "command",
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pl.Render(ports, 0, 80)
	}
}
