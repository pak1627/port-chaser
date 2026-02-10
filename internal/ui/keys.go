package ui

import tea "github.com/charmbracelet/bubbletea"

type KeyBinding struct {
	keys []string
	help struct {
		key  string
		desc string
	}
}

func NewBinding(opts ...BindingOpt) KeyBinding {
	kb := KeyBinding{}
	for _, opt := range opts {
		opt(&kb)
	}
	return kb
}

type BindingOpt func(*KeyBinding)

func WithKeys(keys ...string) BindingOpt {
	return func(kb *KeyBinding) {
		kb.keys = keys
	}
}

func WithHelp(key, desc string) BindingOpt {
	return func(kb *KeyBinding) {
		kb.help.key = key
		kb.help.desc = desc
	}
}

type Help struct {
	Key  string
	Desc string
}

func (kb KeyBinding) Help() Help {
	return Help{
		Key:  kb.help.key,
		Desc: kb.help.desc,
	}
}

type KeyBindings struct {
	Quit             KeyBinding
	NavigateUp       KeyBinding
	NavigateDown     KeyBinding
	NavigateTop      KeyBinding
	NavigateBottom   KeyBinding
	KillProcess      KeyBinding
	Search           KeyBinding
	Confirm          KeyBinding
	Cancel           KeyBinding
	ToggleDockerOnly KeyBinding
	ShowHelp         KeyBinding
	ShowHistory      KeyBinding
	Refresh          KeyBinding
}

func DefaultKeyBindings() *KeyBindings {
	kb := &KeyBindings{}

	kb.Quit = NewBinding(
		WithKeys("q", "ctrl+q"),
		WithHelp("q", "quit"),
	)

	kb.NavigateUp = NewBinding(
		WithKeys("up", "k"),
		WithHelp("↑/k", "up"),
	)

	kb.NavigateDown = NewBinding(
		WithKeys("down", "j"),
		WithHelp("↓/j", "down"),
	)

	kb.NavigateTop = NewBinding(
		WithKeys("g"),
		WithHelp("gg", "top"),
	)

	kb.NavigateBottom = NewBinding(
		WithKeys("G"),
		WithHelp("G", "bottom"),
	)

	kb.KillProcess = NewBinding(
		WithKeys("enter"),
		WithHelp("enter", "kill process"),
	)

	kb.Search = NewBinding(
		WithKeys("/"),
		WithHelp("/", "search"),
	)

	kb.Confirm = NewBinding(
		WithKeys("y", "Y"),
		WithHelp("y", "confirm"),
	)

	kb.Cancel = NewBinding(
		WithKeys("n", "N", "esc"),
		WithHelp("n/esc", "cancel"),
	)

	kb.ToggleDockerOnly = NewBinding(
		WithKeys("d"),
		WithHelp("d", "Docker filter"),
	)

	kb.ShowHelp = NewBinding(
		WithKeys("?"),
		WithHelp("?", "help"),
	)

	kb.ShowHistory = NewBinding(
		WithKeys("h"),
		WithHelp("h", "history"),
	)

	kb.Refresh = NewBinding(
		WithKeys("r", "ctrl+r"),
		WithHelp("r", "refresh"),
	)

	return kb
}

func (kb *KeyBindings) HelpText() string {
	help := []string{
		kb.NavigateUp.Help().Key + "/" + kb.NavigateDown.Help().Key + ": " + "move",
		kb.KillProcess.Help().Key + ": " + "kill",
		kb.Search.Help().Key + ": " + "search",
		kb.ToggleDockerOnly.Help().Key + ": " + "Docker filter",
		kb.Quit.Help().Key + ": " + "quit",
	}
	return " | " + joinStrings(help, " | ")
}

func (kb *KeyBindings) ShortHelp() []KeyBinding {
	return []KeyBinding{
		kb.NavigateUp,
		kb.NavigateDown,
		kb.KillProcess,
		kb.Search,
		kb.ToggleDockerOnly,
		kb.Quit,
	}
}

func (kb *KeyBindings) FullHelp() []KeyBinding {
	return []KeyBinding{
		kb.NavigateUp,
		kb.NavigateDown,
		kb.NavigateTop,
		kb.NavigateBottom,
		kb.KillProcess,
		kb.Search,
		kb.ToggleDockerOnly,
		kb.ShowHistory,
		kb.ShowHelp,
		kb.Refresh,
		kb.Quit,
	}
}

func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}

func (kb KeyBinding) Matches(msg tea.Msg) bool {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return false
	}
	for _, k := range kb.keys {
		if keyMsg.String() == k {
			return true
		}
	}
	return false
}
