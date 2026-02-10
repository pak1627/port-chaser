package components

import (
	"strings"

	"github.com/manson/port-chaser/internal/ui"
)

type StatusBar struct {
	styles    *ui.Styles
	bindings  *ui.KeyBindings
	message   string
	messageFn func() string
}

func NewStatusBar(styles *ui.Styles, bindings *ui.KeyBindings) *StatusBar {
	return &StatusBar{
		styles:   styles,
		bindings: bindings,
	}
}

func (sb *StatusBar) Render(width int) string {
	if sb.message != "" {
		return sb.renderMessage(width)
	}

	if sb.messageFn != nil {
		sb.message = sb.messageFn()
		sb.messageFn = nil
		return sb.renderMessage(width)
	}

	return sb.renderKeyGuide(width)
}

func (sb *StatusBar) renderMessage(width int) string {
	return sb.styles.StatusBar.Render(sb.message)
}

func (sb *StatusBar) renderKeyGuide(width int) string {
	help := []string{
		sb.formatKey("↑", "j") + "/" + sb.formatKey("↓", "k") + ": move",
		sb.formatKey("Enter") + ": kill",
		sb.formatKey("/") + ": search",
		sb.formatKey("d") + ": Docker",
		sb.formatKey("q") + ": quit",
	}

	guide := strings.Join(help, " │ ")

	return sb.styles.StatusBar.Render(guide)
}

func (sb *StatusBar) formatKey(keys ...string) string {
	var formatted []string
	for _, key := range keys {
		formatted = append(formatted, sb.styles.StatusKey.Render(key))
	}
	return strings.Join(formatted, "/")
}

func (sb *StatusBar) SetMessage(message string) {
	sb.message = message
}

func (sb *StatusBar) SetMessageFn(fn func() string) {
	sb.messageFn = fn
}

func (sb *StatusBar) ClearMessage() {
	sb.message = ""
	sb.messageFn = nil
}
