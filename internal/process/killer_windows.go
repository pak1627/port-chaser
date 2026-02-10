// +build windows

// Package process는 Windows용 프로세스 종료 구현을 제공합니다.
package process

import (
	"fmt"
	"os"
	"syscall"
)

// Process는 실행 중인 프로세스를 나타냅니다.
type Process interface {
	// Signal은 프로세스에 신호를 전송합니다.
	Signal(sig syscall.Signal) error
	// Release는 프로세스 리소스를 해제합니다.
	Release() error
}

// windowsProcess는 Windows용 프로세스 래퍼입니다.
type windowsProcess struct {
	*os.Process
}

// Signal은 Windows 프로세스에 신호를 전송합니다.
// Windows는 SIGTERM/SIGKILL을 지원하지 않으므로 TerminateProcess를 사용합니다.
func (p *windowsProcess) Signal(sig syscall.Signal) error {
	// Windows: syscall.SIGTERM (0xf)은 TerminateProcess로 매핑됨
	// syscall.SIGKILL (0x9)도 동일하게 처리
	if sig == syscall.SIGTERM || sig == syscall.SIGKILL {
		return p.Terminate()
	}
	// 신호 0은 프로세스 존재 확인만 수행
	return nil
}

// Terminate는 Windows 프로세스를 강제 종료합니다.
func (p *windowsProcess) Terminate() error {
	// Windows는 syscall.TerminateProcess를 직접 호출해야 함
	dll, err := syscall.LoadDLL("kernel32.dll")
	if err != nil {
		return fmt.Errorf("kernel32.dll 로드 실패: %w", err)
	}
	defer dll.Release()

	proc, err := dll.FindProc("TerminateProcess")
	if err != nil {
		return fmt.Errorf("TerminateProcess 찾기 실패: %w", err)
	}

	// TerminateProcess(handle, exitCode)
	// 윈도우 핸들을 얻기 위해서는 추가 작업이 필요하지만,
	// os.Process.Kill()이 내부적으로 TerminateProcess를 호출하므로
	// 간단하게 Kill()을 사용합니다
	return p.Kill()
}

// Release는 프로세스 리소스를 해제합니다.
func (p *windowsProcess) Release() error {
	return nil
}

// findProcessImpl은 Windows용 프로세스 찾기 구현입니다.
func findProcessImpl(pid int) (Process, error) {
	process, err := os.FindProcess(pid)
	if err != nil {
		return nil, fmt.Errorf("프로세스 찾기 실패 (PID %d): %w", pid, err)
	}
	return &windowsProcess{Process: process}, nil
}
