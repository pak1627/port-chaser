package components

import (
	"fmt"
	"strings"

	"github.com/manson/port-chaser/internal/models"
	"github.com/manson/port-chaser/internal/ui"
)

type Dialog struct {
	styles *ui.Styles
}

func NewDialog(styles *ui.Styles) *Dialog {
	return &Dialog{
		styles: styles,
	}
}

func (d *Dialog) RenderConfirmKill(port *models.PortInfo) string {
	if port == nil {
		return ""
	}

	var lines []string

	title := d.styles.DialogTitle.Render("Confirm Kill Process")
	lines = append(lines, title)
	lines = append(lines, "")

	if port.ShouldDisplayWarning() {
		warning := ui.RenderWarning(d.styles, "System critical process!")
		lines = append(lines, warning)
		lines = append(lines, "")
	}

	lines = append(lines, d.formatPortInfo(port))
	lines = append(lines, "")

	prompt := d.styles.StatusKey.Render("[y]") + " confirm  " +
		d.styles.StatusDim.Render("[n/esc] cancel")
	lines = append(lines, prompt)

	content := strings.Join(lines, "\n")
	return d.styles.Dialog.Render(content)
}

func (d *Dialog) formatPortInfo(port *models.PortInfo) string {
	var info []string

	marker := d.getMarker(port)
	portLine := fmt.Sprintf("Port: %s %d", marker, port.PortNumber)
	info = append(info, portLine)

	processLine := fmt.Sprintf("Process: %s (PID: %d)", port.ProcessName, port.PID)
	info = append(info, processLine)

	if port.IsDocker {
		dockerLine := fmt.Sprintf("Docker: %s (%s)", port.ContainerName, port.ImageName)
		info = append(info, dockerLine)
	}

	if port.Command != "" {
		cmd := port.Command
		if len(cmd) > 50 {
			cmd = cmd[:47] + "..."
		}
		cmdLine := "Command: " + cmd
		info = append(info, cmdLine)
	}

	return strings.Join(info, "\n")
}

func (d *Dialog) getMarker(port *models.PortInfo) string {
	if port.IsDocker {
		return ui.RenderDockerMarker(d.styles)
	}
	if port.IsRecommended() {
		return ui.RenderRecommendedMarker(d.styles)
	}
	if port.ShouldDisplayWarning() {
		return ui.RenderSystemMarker(d.styles)
	}
	return ""
}
