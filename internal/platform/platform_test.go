// Package platform에 대한 테스트입니다.
package platform

import (
	"strings"
	"testing"
)

// TestNewManager는 매니저 생성을 테스트합니다.
func TestNewManager(t *testing.T) {
	manager := NewManager()

	if manager == nil {
		t.Fatal("NewManager()는 nil을 반환하면 안 됩니다")
	}
}

// TestManager_GetOS는 OS 감지를 테스트합니다.
func TestManager_GetOS(t *testing.T) {
	manager := NewManager()

	os := manager.GetOS()
	if os == "" {
		t.Error("GetOS()는 빈 문자열을 반환하면 안 됩니다")
	}

	// 유효한 OS 값인지 확인
	validOS := map[string]bool{
		"darwin":  true,
		"linux":   true,
		"windows": true,
	}

	if !validOS[os] {
		t.Errorf("GetOS() = %s, 유효하지 않은 OS입니다", os)
	}
}

// TestManager_GetArch는 아키텍처 감지를 테스트합니다.
func TestManager_GetArch(t *testing.T) {
	manager := NewManager()

	arch := manager.GetArch()
	if arch == "" {
		t.Error("GetArch()는 빈 문자열을 반환하면 안 됩니다")
	}

	// 유효한 아키텍처 값인지 확인
	validArch := map[string]bool{
		"amd64": true,
		"arm64": true,
		"386":   true,
		"arm":   true,
	}

	if !validArch[arch] {
		t.Errorf("GetArch() = %s, 유효하지 않은 아키텍처입니다", arch)
	}
}

// TestManager_IsDarwin은 macOS 감지를 테스트합니다.
func TestManager_IsDarwin(t *testing.T) {
	manager := NewManager()

	isDarwin := manager.IsDarwin()

	// 현재 플랫폼이 macOS인 경우에만 true여야 함
	if manager.GetOS() == "darwin" && !isDarwin {
		t.Error("macOS에서 IsDarwin()은 true여야 합니다")
	}

	if manager.GetOS() != "darwin" && isDarwin {
		t.Error("비 macOS에서 IsDarwin()은 false여야 합니다")
	}
}

// TestManager_IsLinux는 Linux 감지를 테스트합니다.
func TestManager_IsLinux(t *testing.T) {
	manager := NewManager()

	isLinux := manager.IsLinux()

	if manager.GetOS() == "linux" && !isLinux {
		t.Error("Linux에서 IsLinux()은 true여야 합니다")
	}

	if manager.GetOS() != "linux" && isLinux {
		t.Error("비 Linux에서 IsLinux()은 false여야 합니다")
	}
}

// TestManager_IsWindows는 Windows 감지를 테스트합니다.
func TestManager_IsWindows(t *testing.T) {
	manager := NewManager()

	isWindows := manager.IsWindows()

	if manager.GetOS() == "windows" && !isWindows {
		t.Error("Windows에서 IsWindows()는 true여야 합니다")
	}

	if manager.GetOS() != "windows" && isWindows {
		t.Error("비 Windows에서 IsWindows()는 false여야 합니다")
	}
}

// TestManager_SignalName은 시그널 이름 변환을 테스트합니다.
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

// TestManager_HomeDir은 홈 디렉토리 경로를 테스트합니다.
func TestManager_HomeDir(t *testing.T) {
	manager := NewManager()

	home := manager.HomeDir()
	if home == "" {
		t.Error("HomeDir()은 빈 문자열을 반환하면 안 됩니다")
	}

	// 절대 경로인지 확인
	if home[0] != '/' && home[1] != ':' {
		t.Errorf("HomeDir() = %s, 절대 경로여야 합니다", home)
	}
}

// TestManager_ConfigDir은 설정 디렉토리 경로를 테스트합니다.
func TestManager_ConfigDir(t *testing.T) {
	manager := NewManager()

	config := manager.ConfigDir()
	if config == "" {
		t.Error("ConfigDir()은 빈 문자열을 반환하면 안 됩니다")
	}
}

// TestManager_DataDir은 데이터 디렉토리 경로를 테스트합니다.
func TestManager_DataDir(t *testing.T) {
	manager := NewManager()

	data := manager.DataDir()
	if data == "" {
		t.Error("DataDir()은 빈 문자열을 반환하면 안 됩니다")
	}
}

// TestGetAppName은 앱 이름 반환을 테스트합니다.
func TestGetAppName(t *testing.T) {
	name := GetAppName()
	if name != "port-chaser" {
		t.Errorf("GetAppName() = %s, want port-chaser", name)
	}
}

// TestGetConfigPath는 설정 경로 반환을 테스트합니다.
func TestGetConfigPath(t *testing.T) {
	path := GetConfigPath()
	if path == "" {
		t.Error("GetConfigPath()는 빈 문자열을 반환하면 안 됩니다")
	}
	// 포트-chaser가 포함되어야 함
	if !strings.Contains(path, "port-chaser") {
		t.Errorf("GetConfigPath() = %s, port-chaser가 포함되어야 합니다", path)
	}
}

// TestGetDataPath는 데이터 경로 반환을 테스트합니다.
func TestGetDataPath(t *testing.T) {
	path := GetDataPath()
	if path == "" {
		t.Error("GetDataPath()는 빈 문자열을 반환하면 안 됩니다")
	}
	// 포트-chaser가 포함되어야 함
	if !strings.Contains(path, "port-chaser") {
		t.Errorf("GetDataPath() = %s, port-chaser가 포함되어야 합니다", path)
	}
}

// TestGetHistoryPath는 히스토리 경로 반환을 테스트합니다.
func TestGetHistoryPath(t *testing.T) {
	path := GetHistoryPath()
	if path == "" {
		t.Error("GetHistoryPath()는 빈 문자열을 반환하면 안 됩니다")
	}
	// history.db가 포함되어야 함
	if !strings.Contains(path, "history.db") {
		t.Errorf("GetHistoryPath() = %s, history.db가 포함되어야 합니다", path)
	}
}

// TestNormalizePath는 경로 정규화를 테스트합니다.
func TestNormalizePath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains string // 결과에 포함되어야 하는 문자열
	}{
		{
			name:     "tilde 경로",
			input:    "~/test",
			contains: "/test",
		},
		{
			name:     "상대 경로",
			input:    "./test",
			contains: "test",
		},
		{
			name:     "현재 디렉토리",
			input:    ".",
			contains: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizePath(tt.input)
			if tt.contains != "" && !strings.Contains(result, tt.contains) {
				t.Errorf("NormalizePath(%s) = %s, %s가 포함되어야 합니다", tt.input, result, tt.contains)
			}
		})
	}
}

// TestManager_SignalName_Unknown은 알 수 없는 시그널 테스트입니다.
func TestManager_SignalName_Unknown(t *testing.T) {
	manager := NewManager()

	// 존재하지 않는 시그널 번호
	name := manager.SignalName(9999)
	if !strings.HasPrefix(name, "SIG") {
		t.Errorf("알 수 없는 시그널 이름 = %s, SIG로 시작해야 합니다", name)
	}
}
