package components

import (
	"strings"

	"github.com/manson/port-chaser/internal/models"
	"github.com/manson/port-chaser/internal/ui"
)

type PortList struct {
	styles *ui.Styles
}

func NewPortList(styles *ui.Styles) *PortList {
	return &PortList{
		styles: styles,
	}
}

func (pl *PortList) Render(ports []models.PortInfo, selectedIndex int, width int) string {
	if len(ports) == 0 {
		return pl.renderEmpty()
	}

	var sb strings.Builder

	sb.WriteString(pl.renderHeader())

	for i, port := range ports {
		isSelected := i == selectedIndex
		sb.WriteString(pl.renderPortItem(port, isSelected))
	}

	return sb.String()
}

func (pl *PortList) renderEmpty() string {
	return pl.styles.Muted.Render("  No active ports")
}

func (pl *PortList) renderHeader() string {
	header := "  Port   Process      PID     User    Command"
	return pl.styles.Muted.Render(header)
}

func (pl *PortList) renderPortItem(port models.PortInfo, isSelected bool) string {
	markers := pl.renderMarkers(port)

	portNum := pl.formatPortNumber(port.PortNumber)
	processName := pl.formatProcessName(port.ProcessName, 16)
	pid := pl.formatPID(port.PID)
	user := pl.formatUser(port.User, 10)
	command := pl.formatCommand(port.Command, 30)

	line := strings.Join([]string{
		markers,
		portNum,
		processName,
		pid,
		user,
		command,
	}, " ")

	if isSelected {
		return pl.styles.PortSelected.Render(line)
	}
	return pl.styles.PortItem.Render(line)
}

func (pl *PortList) renderMarkers(port models.PortInfo) string {
	var markers []string

	if port.IsDocker {
		markers = append(markers, ui.RenderDockerMarker(pl.styles))
	}

	if port.IsRecommended() {
		markers = append(markers, ui.RenderRecommendedMarker(pl.styles))
	}

	if port.ShouldDisplayWarning() {
		markers = append(markers, ui.RenderSystemMarker(pl.styles))
	}

	if len(markers) == 0 {
		return "   "
	}

	return strings.Join(markers, " ")
}

func (pl *PortList) formatPortNumber(portNum int) string {
	s := intToString(portNum)
	return padRight(s, 6)
}

func (pl *PortList) formatProcessName(name string, maxWidth int) string {
	if len(name) > maxWidth {
		return name[:maxWidth-3] + "..."
	}
	return padRight(name, maxWidth)
}

func (pl *PortList) formatPID(pid int) string {
	s := intToString(pid)
	return padRight(s, 7)
}

func (pl *PortList) formatUser(user string, maxWidth int) string {
	if user == "" {
		user = "-"
	}
	if len(user) > maxWidth {
		return user[:maxWidth]
	}
	return padRight(user, maxWidth)
}

func (pl *PortList) formatCommand(command string, maxWidth int) string {
	if command == "" {
		return "-"
	}
	if len(command) > maxWidth {
		return command[:maxWidth-3] + "..."
	}
	return command
}

func padRight(s string, width int) string {
	for len(s) < width {
		s += " "
	}
	return s
}

func intToString(i int) string {
	if i == 0 {
		return "0"
	}
	var result string
	for i > 0 {
		result = string(rune('0'+i%10)) + result
		i /= 10
	}
	return result
}
