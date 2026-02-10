// +build darwin linux

// Package process는 POSIX 시스템용 프로세스 종료 구현을 제공합니다.
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

// osProcess는 표준 라이브러리의 os.Process 래퍼입니다.
type osProcess struct {
	*os.Process
}

// Release는 프로세스 리소스를 해제합니다.
func (p *osProcess) Release() error {
	// os.Process는 명시적인 Release가 필요 없음
	return nil
}

// Signal은 프로세스에 시스템 신호를 전송합니다.
// os.Process.Signal은 os.Signal을 받지만, 우리는 syscall.Signal로 변환합니다.
func (p *osProcess) Signal(sig syscall.Signal) error {
	// os.Process.Signal으로 전달 (os.Signal은 인터페이스, syscall.Signal은 int)
	return p.Process.Signal(sig)
}

// findProcessImpl은 POSIX 시스템용 프로세스 찾기 구현입니다.
func findProcessImpl(pid int) (Process, error) {
	process, err := os.FindProcess(pid)
	if err != nil {
		return nil, fmt.Errorf("프로세스 찾기 실패 (PID %d): %w", pid, err)
	}
	return &osProcess{Process: process}, nil
}
