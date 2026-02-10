package platform

import (
	"strings"
	"testing"
)

func TestNewManager(t *testing.T) {
	manager := NewManager()

	if manager == nil {
		t.Fatal("NewManager() should not return nil")
	}
}

func TestManager_GetOS(t *testing.T) {
	manager := NewManager()

	os := manager.GetOS()
	if os == "" {
		t.Error("GetOS() should not return empty string")
	}

	validOS := map[string]bool{
		"darwin":  true,
		"linux":   true,
		"windows": true,
	}

	if !validOS[os] {
		t.Errorf("GetOS() = %s, invalid OS", os)
	}
}

func TestManager_GetArch(t *testing.T) {
	manager := NewManager()

	arch := manager.GetArch()
	if arch == "" {
		t.Error("GetArch() should not return empty string")
	}

	validArch := map[string]bool{
		"amd64": true,
		"arm64": true,
		"386":   true,
		"arm":   true,
	}

	if !validArch[arch] {
		t.Errorf("GetArch() = %s, invalid architecture", arch)
	}
}

func TestManager_IsDarwin(t *testing.T) {
	manager := NewManager()

	isDarwin := manager.IsDarwin()

	if manager.GetOS() == "darwin" && !isDarwin {
		t.Error("IsDarwin() should be true on macOS")
	}

	if manager.GetOS() != "darwin" && isDarwin {
		t.Error("IsDarwin() should be false on non-macOS")
	}
}

func TestManager_IsLinux(t *testing.T) {
	manager := NewManager()

	isLinux := manager.IsLinux()

	if manager.GetOS() == "linux" && !isLinux {
		t.Error("IsLinux() should be true on Linux")
	}

	if manager.GetOS() != "linux" && isLinux {
		t.Error("IsLinux() should be false on non-Linux")
	}
}

func TestManager_IsWindows(t *testing.T) {
	manager := NewManager()

	isWindows := manager.IsWindows()

	if manager.GetOS() == "windows" && !isWindows {
		t.Error("IsWindows() should be true on Windows")
	}

	if manager.GetOS() != "windows" && isWindows {
		t.Error("IsWindows() should be false on non-Windows")
	}
}

func TestManager_SignalName(t *testing.T) {
	manager := NewManager()

	tests := []struct {
		sigNum int
		want   string
	}{
		{1, "SIGHUP"},
		{2, "SIGINT"},
		{3, "SIGQUIT"},
		{4, "SIGILL"},
		{5, "SIGTRAP"},
		{6, "SIGABRT"},
		{7, "SIGBUS"},
		{8, "SIGFPE"},
		{9, "SIGKILL"},
		{10, "SIGUSR1"},
		{11, "SIGSEGV"},
		{12, "SIGUSR2"},
		{13, "SIGPIPE"},
		{14, "SIGALRM"},
		{15, "SIGTERM"},
		{16, "SIGSTKFLT"},
		{17, "SIGCHLD"},
		{18, "SIGCONT"},
		{19, "SIGSTOP"},
		{20, "SIGTSTP"},
		{21, "SIGTTIN"},
		{22, "SIGTTOU"},
		{23, "SIGURG"},
		{24, "SIGXCPU"},
		{25, "SIGXFSZ"},
		{26, "SIGVTALRM"},
		{27, "SIGPROF"},
		{28, "SIGWINCH"},
		{29, "SIGIO"},
		{30, "SIGPWR"},
		{31, "SIGSYS"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := manager.SignalName(tt.sigNum)
			if got != tt.want {
				t.Errorf("SignalName(%d) = %s, want %s", tt.sigNum, got, tt.want)
			}
		})
	}
}

func TestManager_HomeDir(t *testing.T) {
	manager := NewManager()

	home := manager.HomeDir()
	if home == "" {
		t.Error("HomeDir() should not return empty string")
	}

	if home[0] != '/' && home[1] != ':' {
		t.Errorf("HomeDir() = %s, should be absolute path", home)
	}
}

func TestManager_ConfigDir(t *testing.T) {
	manager := NewManager()

	config := manager.ConfigDir()
	if config == "" {
		t.Error("ConfigDir() should not return empty string")
	}
}

func TestManager_DataDir(t *testing.T) {
	manager := NewManager()

	data := manager.DataDir()
	if data == "" {
		t.Error("DataDir() should not return empty string")
	}
}

func TestGetAppName(t *testing.T) {
	name := GetAppName()
	if name != "port-chaser" {
		t.Errorf("GetAppName() = %s, want port-chaser", name)
	}
}

func TestGetConfigPath(t *testing.T) {
	path := GetConfigPath()
	if path == "" {
		t.Error("GetConfigPath() should not return empty string")
	}
	if !strings.Contains(path, "port-chaser") {
		t.Errorf("GetConfigPath() = %s, should contain port-chaser", path)
	}
}

func TestGetDataPath(t *testing.T) {
	path := GetDataPath()
	if path == "" {
		t.Error("GetDataPath() should not return empty string")
	}
	if !strings.Contains(path, "port-chaser") {
		t.Errorf("GetDataPath() = %s, should contain port-chaser", path)
	}
}

func TestGetHistoryPath(t *testing.T) {
	path := GetHistoryPath()
	if path == "" {
		t.Error("GetHistoryPath() should not return empty string")
	}
	if !strings.Contains(path, "history.db") {
		t.Errorf("GetHistoryPath() = %s, should contain history.db", path)
	}
}

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains string
	}{
		{
			name:     "tilde path",
			input:    "~/test",
			contains: "/test",
		},
		{
			name:     "relative path",
			input:    "./test",
			contains: "test",
		},
		{
			name:     "current directory",
			input:    ".",
			contains: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizePath(tt.input)
			if tt.contains != "" && !strings.Contains(result, tt.contains) {
				t.Errorf("NormalizePath(%s) = %s, should contain %s", tt.input, result, tt.contains)
			}
		})
	}
}

func TestManager_SignalName_Unknown(t *testing.T) {
	manager := NewManager()

	name := manager.SignalName(9999)
	if !strings.HasPrefix(name, "SIG") {
		t.Errorf("unknown signal name = %s, should start with SIG", name)
	}
}
