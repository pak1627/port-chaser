package ui

import "github.com/charmbracelet/lipgloss"

type Styles struct {
	Header lipgloss.Style

	HeaderTitle lipgloss.Style

	HeaderSubtitle lipgloss.Style

	PortList lipgloss.Style

	PortItem lipgloss.Style

	PortSelected lipgloss.Style

	DockerMarker lipgloss.Style

	RecommendedMarker lipgloss.Style

	SystemMarker lipgloss.Style

	StatusBar lipgloss.Style

	StatusKey lipgloss.Style

	StatusDim lipgloss.Style

	Dialog lipgloss.Style

	DialogTitle lipgloss.Style

	DialogBorder lipgloss.Style

	Warning lipgloss.Style

	Error lipgloss.Style

	Success lipgloss.Style

	Muted lipgloss.Style
}

func DefaultStyles() *Styles {
	s := &Styles{}

	blue := lipgloss.Color("86")
	yellow := lipgloss.Color("226")
	red := lipgloss.Color("196")
	green := lipgloss.Color("46")
	gray := lipgloss.Color("245")

	s.Header = lipgloss.NewStyle().
		Padding(0, 1).
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(lipgloss.Color("238"))

	s.HeaderTitle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("212")).
		Width(20)

	s.HeaderSubtitle = lipgloss.NewStyle().
		Faint(true).
		Foreground(gray)

	s.PortList = lipgloss.NewStyle().
		Padding(1, 2)

	s.PortItem = lipgloss.NewStyle().
		Padding(0, 1).
		Width(80)

	s.PortSelected = lipgloss.NewStyle().
		Background(lipgloss.Color("240")).
		Foreground(lipgloss.Color("15")).
		Bold(true).
		Padding(0, 1)

	s.DockerMarker = lipgloss.NewStyle().
		Foreground(blue).
		Bold(true)

	s.RecommendedMarker = lipgloss.NewStyle().
		Foreground(yellow).
		Bold(true)

	s.SystemMarker = lipgloss.NewStyle().
		Foreground(red).
		Bold(true)

	s.StatusBar = lipgloss.NewStyle().
		Padding(1, 2).
		Border(lipgloss.NormalBorder(), true, false, false, false).
		BorderForeground(lipgloss.Color("238"))

	s.StatusKey = lipgloss.NewStyle().
		Foreground(lipgloss.Color("212")).
		Bold(true)

	s.StatusDim = lipgloss.NewStyle().
		Faint(true).
		Foreground(gray)

	s.Dialog = lipgloss.NewStyle().
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("238")).
		Width(50)

	s.DialogTitle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("212")).
		MarginBottom(1)

	s.DialogBorder = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("238"))

	s.Warning = lipgloss.NewStyle().
		Foreground(yellow).
		Bold(true)

	s.Error = lipgloss.NewStyle().
		Foreground(red).
		Bold(true)

	s.Success = lipgloss.NewStyle().
		Foreground(green).
		Bold(true)

	s.Muted = lipgloss.NewStyle().
		Faint(true).
		Foreground(gray)

	return s
}

func RenderDockerMarker(styles *Styles) string {
	return styles.DockerMarker.Render("[D]")
}

func RenderRecommendedMarker(styles *Styles) string {
	return styles.RecommendedMarker.Render("[!]")
}

func RenderSystemMarker(styles *Styles) string {
	return styles.SystemMarker.Render("[S]")
}

func RenderWarning(styles *Styles, msg string) string {
	return styles.Warning.Render("⚠ " + msg)
}

func RenderError(styles *Styles, msg string) string {
	return styles.Error.Render("✗ " + msg)
}

func RenderSuccess(styles *Styles, msg string) string {
	return styles.Success.Render("✓ " + msg)
}
