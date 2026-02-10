package components

import (
	"fmt"
	"strings"

	"github.com/manson/port-chaser/internal/ui"
)

type Header struct {
	styles *ui.Styles
	title  string
}

func NewHeader(styles *ui.Styles, title string) *Header {
	return &Header{
		styles: styles,
		title:  title,
	}
}

func (h *Header) Render(portCount int, showDockerOnly bool, width int) string {
	titleLine := h.renderTitleLine(portCount, showDockerOnly)
	divider := strings.Repeat("─", width-2)

	return h.styles.Header.Render(titleLine + "\n" + divider)
}

func (h *Header) renderTitleLine(portCount int, showDockerOnly bool) string {
	title := h.styles.HeaderTitle.Render(h.title)

	portText := fmt.Sprintf("%d ports", portCount)
	portCountDisplay := h.styles.HeaderSubtitle.Render(portText)

	var filterDisplay string
	if showDockerOnly {
		filterDisplay = h.styles.DockerMarker.Render("[Docker only]")
	} else {
		filterDisplay = h.styles.Muted.Render("[All]")
	}

	helpHint := h.styles.StatusKey.Render("[?]")

	left := title + " " + portCountDisplay + " " + filterDisplay
	right := helpHint

	availableWidth := 80
	leftWidth := len(left)
	rightWidth := len(right)
	padding := availableWidth - leftWidth - rightWidth

	if padding < 1 {
		padding = 1
	}

	middle := strings.Repeat(" ", padding)

	return left + middle + right
}

func (h *Header) SetTitle(title string) {
	h.title = title
}
