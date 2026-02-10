// Package process는 프로세스 종료 기능을 제공합니다.
package process

import (
	"context"
	"fmt"
	"syscall"
	"time"

	"github.com/manson/port-chaser/internal/models"
)

// KillResult는 프로세스 종료 결과를 나타냅니다.
type KillResult struct {
	Success    bool      // 종료 성공 여부
	Method     KillMethod // 사용된 종료 방법
	Message    string    // 결과 메시지
	Duration   time.Duration // 종료까지 걸린 시간
}

// KillMethod는 사용된 종료 방법을 나타냅니다.
type KillMethod string

const (
	KillMethodSIGTERM KillMethod = "SIGTERM" // 정상 종료
	KillMethodSIGKILL KillMethod = "SIGKILL" // 강제 종료
	KillMethodFailed  KillMethod = "FAILED"  // 종료 실패
)

// Killer는 프로세스 종료 기능을 제공하는 인터페이스입니다.
type Killer interface {
	// Kill은 지정된 PID의 프로세스를 종료합니다.
	// 먼저 SIGTERM을 전송하고, 3초 내에 종료하지 않으면 SIGKILL을 전송합니다.
	Kill(ctx context.Context, pid int, portInfo *models.PortInfo) (*KillResult, error)

	// KillWithTimeout은 타임아웃을 지정하여 프로세스를 종료합니다.
	KillWithTimeout(ctx context.Context, pid int, timeout time.Duration, portInfo *models.PortInfo) (*KillResult, error)

	// IsRunning은 프로세스가 실행 중인지 확인합니다.
	IsRunning(pid int) (bool, error)
}

// ProcessKiller는 Killer 인터페이스의 표준 구현입니다.
type ProcessKiller struct {
	// GracePeriod는 SIGTERM 후 SIGKILL 전 대기 시간입니다.
	GracePeriod time.Duration

	// SystemProcessProtection은 시스템 프로세스 보호 기능입니다.
	SystemProcessProtection bool
}

// NewProcessKiller는 새로운 ProcessKiller를 생성합니다.
func NewProcessKiller() *ProcessKiller {
	return &ProcessKiller{
		GracePeriod:             3 * time.Second,
		SystemProcessProtection: true,
	}
}

// NewProcessKillerWithGracePeriod는 사용자 정의 대기 시간으로 ProcessKiller를 생성합니다.
func NewProcessKillerWithGracePeriod(gracePeriod time.Duration) *ProcessKiller {
	return &ProcessKiller{
		GracePeriod:             gracePeriod,
		SystemProcessProtection: true,
	}
}

// Kill은 프로세스를 종료합니다. SIGTERM → 3초 대기 → SIGKILL 순서로 진행합니다.
func (k *ProcessKiller) Kill(ctx context.Context, pid int, portInfo *models.PortInfo) (*KillResult, error) {
	return k.KillWithTimeout(ctx, pid, k.GracePeriod, portInfo)
}

// KillWithTimeout은 타임아웃을 지정하여 프로세스를 종료합니다.
func (k *ProcessKiller) KillWithTimeout(ctx context.Context, pid int, timeout time.Duration, portInfo *models.PortInfo) (*KillResult, error) {
	startTime := time.Now()

	// 1. 시스템 프로세스 보호 확인
	if k.SystemProcessProtection && portInfo != nil && portInfo.ShouldDisplayWarning() {
		return &KillResult{
			Success: false,
			Method:  KillMethodFailed,
			Message: fmt.Sprintf("시스템 중요 프로세스(PID %d)는 보호됩니다", pid),
		}, nil
	}

	// 2. 프로세스 실행 중 확인
	running, err := k.IsRunning(pid)
	if err != nil {
		return &KillResult{
			Success: false,
			Method:  KillMethodFailed,
			Message: fmt.Sprintf("프로세스 상태 확인 실패: %v", err),
		}, err
	}
	if !running {
		return &KillResult{
			Success: false,
			Method:  KillMethodFailed,
			Message: fmt.Sprintf("프로세스 PID %d가 실행 중이 아닙니다", pid),
		}, nil
	}

	// 3. SIGTERM 전송
	if err := k.sendSignal(pid, syscall.SIGTERM); err != nil {
		return &KillResult{
			Success: false,
			Method:  KillMethodFailed,
			Message: fmt.Sprintf("SIGTERM 전송 실패: %v", err),
		}, err
	}

	// 4. 그레이스풀 종료 대기
	graceCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-graceCtx.Done():
			// 타임아웃: SIGKILL 전송
			return k.forceKill(ctx, pid, startTime)
		case <-ticker.C:
			// 프로세스 종료 확인
			running, err := k.IsRunning(pid)
			if err == nil && !running {
				return &KillResult{
					Success:  true,
					Method:   KillMethodSIGTERM,
					Message:  fmt.Sprintf("PID %d 프로세스가 정상 종료되었습니다", pid),
					Duration: time.Since(startTime),
				}, nil
			}
		case <-ctx.Done():
			return &KillResult{
				Success: false,
				Method:  KillMethodFailed,
				Message: "작업이 취소되었습니다",
			}, ctx.Err()
		}
	}
}

// forceKill은 SIGKILL을 전송하여 프로세스를 강제 종료합니다.
func (k *ProcessKiller) forceKill(ctx context.Context, pid int, startTime time.Time) (*KillResult, error) {
	// 프로세스가 이미 종료되었는지 마지막 확인
	running, err := k.IsRunning(pid)
	if err == nil && !running {
		return &KillResult{
			Success:  true,
			Method:   KillMethodSIGTERM,
			Message:  fmt.Sprintf("PID %d 프로세스가 종료되었습니다 (타이밍 차이)", pid),
			Duration: time.Since(startTime),
		}, nil
	}

	// SIGKILL 전송
	if err := k.sendSignal(pid, syscall.SIGKILL); err != nil {
		return &KillResult{
			Success: false,
			Method:  KillMethodFailed,
			Message: fmt.Sprintf("SIGKILL 전송 실패: %v", err),
		}, err
	}

	// SIGKILL 후 즉시 확인
	time.Sleep(50 * time.Millisecond)
	running, _ = k.IsRunning(pid)

	result := &KillResult{
		Success:  !running,
		Method:   KillMethodSIGKILL,
		Message:  fmt.Sprintf("PID %d 프로세스에 SIGKILL을 전송했습니다", pid),
		Duration: time.Since(startTime),
	}

	if !running {
		result.Message = fmt.Sprintf("PID %d 프로세스가 강제 종료되었습니다", pid)
		result.Success = true
	} else {
		result.Message = fmt.Sprintf("PID %d 프로세스 종료에 실패했습니다", pid)
		result.Success = false
	}

	return result, nil
}

// IsRunning은 프로세스가 실행 중인지 확인합니다.
func (k *ProcessKiller) IsRunning(pid int) (bool, error) {
	// 프로세스에 신호 0을 전송하여 존재 확인
	// 에러가 nil이면 프로세스가 실행 중임
	err := k.sendSignal(pid, 0)
	if err != nil {
		return false, nil // 프로세스가 존재하지 않음
	}
	return true, nil
}

// sendSignal은 프로세스에 신호를 전송합니다.
func (k *ProcessKiller) sendSignal(pid int, sig syscall.Signal) error {
	// syscall.Kill은 프로세스에 신호를 전송합니다.
	// 신호가 0이면 실제 신호는 전송하지 않고 프로세스 존재 여부만 확인합니다.
	process, err := findProcess(pid)
	if err != nil {
		return err
	}
	defer process.Release()
	return process.Signal(sig)
}

// findProcess는 PID로 프로세스를 찾습니다.
// 플랫폼별로 다른 구현이 필요합니다.
func findProcess(pid int) (Process, error) {
	return findProcessImpl(pid)
}
