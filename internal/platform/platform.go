// Package platform은 크로스 플랫폼 유틸리티를 제공합니다.
package platform

import (
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// Manager는 플랫폼 관리자 인터페이스입니다.
type Manager interface {
	// GetOS는 운영체제를 반환합니다.
	GetOS() string

	// GetArch는 아키텍처를 반환합니다.
	GetArch() string

	// IsDarwin은 macOS인지 확인합니다.
	IsDarwin() bool

	// IsLinux는 Linux인지 확인합니다.
	IsLinux() bool

	// IsWindows는 Windows인지 확인합니다.
	IsWindows() bool

	// SignalName은 시그널 번호를 이름으로 변환합니다.
	SignalName(sigNum int) string

	// HomeDir은 홈 디렉토리 경로를 반환합니다.
	HomeDir() string

	// ConfigDir은 설정 디렉토리 경로를 반환합니다.
	ConfigDir() string

	// DataDir은 데이터 디렉토리 경로를 반환합니다.
	DataDir() string
}

// defaultManager는 기본 플랫폼 관리자 구현입니다.
type defaultManager struct {
	os   string
	arch string
}

// NewManager는 새로운 플랫폼 관리자를 생성합니다.
func NewManager() Manager {
	return &defaultManager{
		os:   runtime.GOOS,
		arch: runtime.GOARCH,
	}
}

// GetOS는 운영체제를 반환합니다.
func (m *defaultManager) GetOS() string {
	return m.os
}

// GetArch는 아키텍처를 반환합니다.
func (m *defaultManager) GetArch() string {
	return m.arch
}

// IsDarwin은 macOS인지 확인합니다.
func (m *defaultManager) IsDarwin() bool {
	return m.os == "darwin"
}

// IsLinux는 Linux인지 확인합니다.
func (m *defaultManager) IsLinux() bool {
	return m.os == "linux"
}

// IsWindows는 Windows인지 확인합니다.
func (m *defaultManager) IsWindows() bool {
	return m.os == "windows"
}

// SignalName은 시그널 번호를 이름으로 변환합니다.
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

// HomeDir은 홈 디렉토리 경로를 반환합니다.
func (m *defaultManager) HomeDir() string {
	// 홈 디렉토리 환경 변수 확인
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	if home := os.Getenv("USERPROFILE"); home != "" {
		return home
	}

	// 사용자 정보 조회
	usr, err := user.Current()
	if err == nil && usr.HomeDir != "" {
		return usr.HomeDir
	}

	// 폴백: 현재 디렉토리 반환
	return "."
}

// ConfigDir은 설정 디렉토리 경로를 반환합니다.
func (m *defaultManager) ConfigDir() string {
	home := m.HomeDir()

	switch m.os {
	case "darwin":
		// macOS: ~/Library/Application Support
		return filepath.Join(home, "Library", "Application Support")
	case "linux":
		// Linux: ~/.config 또는 XDG_CONFIG_HOME
		if config := os.Getenv("XDG_CONFIG_HOME"); config != "" {
			return config
		}
		return filepath.Join(home, ".config")
	case "windows":
		// Windows: %APPDATA% 또는 ~/AppData/Roaming
		if appdata := os.Getenv("APPDATA"); appdata != "" {
			return appdata
		}
		return filepath.Join(home, "AppData", "Roaming")
	default:
		// 기본: ~/.config
		return filepath.Join(home, ".config")
	}
}

// DataDir은 데이터 디렉토리 경로를 반환합니다.
func (m *defaultManager) DataDir() string {
	home := m.HomeDir()

	switch m.os {
	case "darwin":
		// macOS: ~/Library/Application Support
		return filepath.Join(home, "Library", "Application Support")
	case "linux":
		// Linux: ~/.local/share 또는 XDG_DATA_HOME
		if data := os.Getenv("XDG_DATA_HOME"); data != "" {
			return data
		}
		return filepath.Join(home, ".local", "share")
	case "windows":
		// Windows: %LOCALAPPDATA% 또는 ~/AppData/Local
		if localappdata := os.Getenv("LOCALAPPDATA"); localappdata != "" {
			return localappdata
		}
		return filepath.Join(home, "AppData", "Local")
	default:
		// 기본: ~/.local/share
		return filepath.Join(home, ".local", "share")
	}
}

// GetAppName은 애플리케이션 이름을 반환합니다.
func GetAppName() string {
	return "port-chaser"
}

// GetConfigPath는 설정 파일 경로를 반환합니다.
func GetConfigPath() string {
	manager := NewManager()
	return filepath.Join(manager.ConfigDir(), GetAppName())
}

// GetDataPath는 데이터 디렉토리 경로를 반환합니다.
func GetDataPath() string {
	manager := NewManager()
	return filepath.Join(manager.DataDir(), GetAppName())
}

// GetHistoryPath는 히스토리 데이터베이스 경로를 반환합니다.
func GetHistoryPath() string {
	return filepath.Join(GetDataPath(), "history.db")
}

// NormalizePath는 경로를 정규화합니다.
func NormalizePath(path string) string {
	// tilde 확장
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

	// 절대 경로로 변환
	abs, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	return abs
}
