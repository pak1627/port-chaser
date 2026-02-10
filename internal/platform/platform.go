package platform

import (
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

type Manager interface {
	GetOS() string
	GetArch() string
	IsDarwin() bool
	IsLinux() bool
	IsWindows() bool
	SignalName(sigNum int) string
	HomeDir() string
	ConfigDir() string
	DataDir() string
}

type defaultManager struct {
	os   string
	arch string
}

func NewManager() Manager {
	return &defaultManager{
		os:   runtime.GOOS,
		arch: runtime.GOARCH,
	}
}

func (m *defaultManager) GetOS() string {
	return m.os
}

func (m *defaultManager) GetArch() string {
	return m.arch
}

func (m *defaultManager) IsDarwin() bool {
	return m.os == "darwin"
}

func (m *defaultManager) IsLinux() bool {
	return m.os == "linux"
}

func (m *defaultManager) IsWindows() bool {
	return m.os == "windows"
}

func (m *defaultManager) SignalName(sigNum int) string {
	signals := map[int]string{
		1:  "SIGHUP",
		2:  "SIGINT",
		3:  "SIGQUIT",
		4:  "SIGILL",
		5:  "SIGTRAP",
		6:  "SIGABRT",
		7:  "SIGBUS",
		8:  "SIGFPE",
		9:  "SIGKILL",
		10: "SIGUSR1",
		11: "SIGSEGV",
		12: "SIGUSR2",
		13: "SIGPIPE",
		14: "SIGALRM",
		15: "SIGTERM",
		16: "SIGSTKFLT",
		17: "SIGCHLD",
		18: "SIGCONT",
		19: "SIGSTOP",
		20: "SIGTSTP",
		21: "SIGTTIN",
		22: "SIGTTOU",
		23: "SIGURG",
		24: "SIGXCPU",
		25: "SIGXFSZ",
		26: "SIGVTALRM",
		27: "SIGPROF",
		28: "SIGWINCH",
		29: "SIGIO",
		30: "SIGPWR",
		31: "SIGSYS",
	}

	if name, ok := signals[sigNum]; ok {
		return name
	}
	return "SIG" + strconv.Itoa(sigNum)
}

func (m *defaultManager) HomeDir() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	if home := os.Getenv("USERPROFILE"); home != "" {
		return home
	}

	usr, err := user.Current()
	if err == nil && usr.HomeDir != "" {
		return usr.HomeDir
	}

	return "."
}

func (m *defaultManager) ConfigDir() string {
	home := m.HomeDir()

	switch m.os {
	case "darwin":
		return filepath.Join(home, "Library", "Application Support")
	case "linux":
		if config := os.Getenv("XDG_CONFIG_HOME"); config != "" {
			return config
		}
		return filepath.Join(home, ".config")
	case "windows":
		if appdata := os.Getenv("APPDATA"); appdata != "" {
			return appdata
		}
		return filepath.Join(home, "AppData", "Roaming")
	default:
		return filepath.Join(home, ".config")
	}
}

func (m *defaultManager) DataDir() string {
	home := m.HomeDir()

	switch m.os {
	case "darwin":
		return filepath.Join(home, "Library", "Application Support")
	case "linux":
		if data := os.Getenv("XDG_DATA_HOME"); data != "" {
			return data
		}
		return filepath.Join(home, ".local", "share")
	case "windows":
		if localappdata := os.Getenv("LOCALAPPDATA"); localappdata != "" {
			return localappdata
		}
		return filepath.Join(home, "AppData", "Local")
	default:
		return filepath.Join(home, ".local", "share")
	}
}

func GetAppName() string {
	return "port-chaser"
}

func GetConfigPath() string {
	manager := NewManager()
	return filepath.Join(manager.ConfigDir(), GetAppName())
}

func GetDataPath() string {
	manager := NewManager()
	return filepath.Join(manager.DataDir(), GetAppName())
}

func GetHistoryPath() string {
	return filepath.Join(GetDataPath(), "history.db")
}

func NormalizePath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home := os.Getenv("HOME")
		if home == "" {
			if usr, err := user.Current(); err == nil {
				home = usr.HomeDir
			}
		}
		if home != "" {
			return filepath.Join(home, path[2:])
		}
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	return abs
}
