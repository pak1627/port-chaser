package components

import (
	"strings"
	"testing"

	"github.com/manson/port-chaser/internal/models"
	"github.com/manson/port-chaser/internal/ui"
)

func TestPortList_RenderEmpty(t *testing.T) {
	styles := ui.DefaultStyles()
	pl := NewPortList(styles)

	result := pl.Render([]models.PortInfo{}, -1, 80)

	if !strings.Contains(result, "No active ports") {
		t.Error("empty list message should be present")
	}
}

func TestPortList_Render(t *testing.T) {
	styles := ui.DefaultStyles()
	pl := NewPortList(styles)

	ports := []models.PortInfo{
		{PortNumber: 3000, ProcessName: "node", PID: 1001, User: "user", Command: "npm start"},
		{PortNumber: 8080, ProcessName: "python", PID: 1002, User: "dev", Command: "python app.py"},
	}

	result := pl.Render(ports, 0, 80)

	if !strings.Contains(result, "Port") {
		t.Error("header should contain 'Port'")
	}

	if !strings.Contains(result, "3000") {
		t.Error("port 3000 should be displayed")
	}

	if !strings.Contains(result, "8080") {
		t.Error("port 8080 should be displayed")
	}

	if !strings.Contains(result, "node") {
		t.Error("process 'node' should be displayed")
	}

	if !strings.Contains(result, "python") {
		t.Error("process 'python' should be displayed")
	}
}

func TestPortList_RenderDockerMarker(t *testing.T) {
	styles := ui.DefaultStyles()
	pl := NewPortList(styles)

	ports := []models.PortInfo{
		{PortNumber: 3000, ProcessName: "node", PID: 1001, IsDocker: true, ContainerName: "app"},
	}

	result := pl.Render(ports, 0, 80)

	if !strings.Contains(result, "[D]") {
		t.Error("Docker marker [D] should be displayed")
	}
}

func TestPortList_RenderRecommendedMarker(t *testing.T) {
	styles := ui.DefaultStyles()
	pl := NewPortList(styles)

	ports := []models.PortInfo{
		{PortNumber: 3000, ProcessName: "node", PID: 1001, KillCount: 5},
	}

	result := pl.Render(ports, 0, 80)

	if !strings.Contains(result, "[!]") {
		t.Error("recommended marker [!] should be displayed")
	}
}

func TestPortList_RenderSystemMarker(t *testing.T) {
	styles := ui.DefaultStyles()
	pl := NewPortList(styles)

	ports := []models.PortInfo{
		{PortNumber: 80, ProcessName: "httpd", PID: 50, IsSystem: true},
	}

	result := pl.Render(ports, 0, 80)

	if !strings.Contains(result, "[S]") {
		t.Error("system marker [S] should be displayed")
	}
}

func TestHeader_Render(t *testing.T) {
	styles := ui.DefaultStyles()
	header := NewHeader(styles, "Port Chaser")

	result := header.Render(10, false, 80)

	if !strings.Contains(result, "Port Chaser") {
		t.Error("title should be displayed")
	}

	if !strings.Contains(result, "10 ports") {
		t.Error("port count should be displayed")
	}
}

func TestHeader_RenderDockerOnly(t *testing.T) {
	styles := ui.DefaultStyles()
	header := NewHeader(styles, "Port Chaser")

	result := header.Render(5, true, 80)

	if !strings.Contains(result, "[Docker only]") {
		t.Error("Docker filter status should be displayed")
	}
}

func TestStatusBar_Render(t *testing.T) {
	styles := ui.DefaultStyles()
	bindings := ui.DefaultKeyBindings()
	statusBar := NewStatusBar(styles, bindings)

	result := statusBar.Render(80)

	if !strings.Contains(result, "q") {
		t.Error("quit key guide should be present")
	}
}

func TestStatusBar_SetMessage(t *testing.T) {
	styles := ui.DefaultStyles()
	bindings := ui.DefaultKeyBindings()
	statusBar := NewStatusBar(styles, bindings)

	statusBar.SetMessage("test message")
	result := statusBar.Render(80)

	if !strings.Contains(result, "test message") {
		t.Error("status message should be displayed")
	}
}

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

	if !strings.Contains(result, "Confirm Kill Process") {
		t.Error("dialog title should be displayed")
	}

	if !strings.Contains(result, "3000") {
		t.Error("port number should be displayed")
	}

	if !strings.Contains(result, "node") {
		t.Error("process name should be displayed")
	}

	if !strings.Contains(result, "[y]") {
		t.Error("confirm prompt should be present")
	}
}

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

	if !strings.Contains(result, "System critical process") {
		t.Error("system process warning should be displayed")
	}
}

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
		t.Error("Docker info should be displayed")
	}

	if !strings.Contains(result, "my-app") {
		t.Error("container name should be displayed")
	}
}

func TestDialog_RenderConfirmKill_Nil(t *testing.T) {
	styles := ui.DefaultStyles()
	dialog := NewDialog(styles)

	result := dialog.RenderConfirmKill(nil)

	if result != "" {
		t.Error("nil port should return empty string")
	}
}

func Benchmark_PortList_Render(b *testing.B) {
	styles := ui.DefaultStyles()
	pl := NewPortList(styles)

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
