package process

import (
	"context"
	"fmt"
	"syscall"
	"time"

	"github.com/manson/port-chaser/internal/models"
)

// KillResult contains the outcome of a process termination attempt.
// It provides detailed information about how the process was terminated.
type KillResult struct {
	Success  bool          // true if process was terminated successfully
	Method   KillMethod    // termination method used (SIGTERM, SIGKILL, or FAILED)
	Message  string        // human-readable result message
	Duration time.Duration // time taken to terminate the process
}

// KillMethod represents the method used to terminate a process.
type KillMethod string

const (
	KillMethodSIGTERM KillMethod = "SIGTERM" // graceful termination signal
	KillMethodSIGKILL KillMethod = "SIGKILL" // force termination signal
	KillMethodFailed  KillMethod = "FAILED"  // termination failed
)

// Killer defines the interface for process termination operations.
// Implementations can provide platform-specific process killing behavior.
type Killer interface {
	// Kill attempts to terminate the process with the given PID using default grace period
	Kill(ctx context.Context, pid int, portInfo *models.PortInfo) (*KillResult, error)
	// KillWithTimeout attempts to terminate with a custom grace period before forcing
	KillWithTimeout(ctx context.Context, pid int, timeout time.Duration, portInfo *models.PortInfo) (*KillResult, error)
	// IsRunning checks if a process with the given PID is currently active
	IsRunning(pid int) (bool, error)
}

// ProcessKiller implements the Killer interface with a two-phase termination strategy.
// First tries SIGTERM (graceful shutdown), then SIGKILL (force) if needed.
type ProcessKiller struct {
	GracePeriod             time.Duration // how long to wait for graceful shutdown
	SystemProcessProtection bool          // whether to prevent killing system processes
}

// NewProcessKiller creates a new ProcessKiller with default settings.
// Default grace period is 3 seconds, system process protection is enabled.
func NewProcessKiller() *ProcessKiller {
	return &ProcessKiller{
		GracePeriod:             3 * time.Second,
		SystemProcessProtection: true,
	}
}

// NewProcessKillerWithGracePeriod creates a ProcessKiller with a custom grace period.
// Use this when you need more or less time for processes to shut down gracefully.
func NewProcessKillerWithGracePeriod(gracePeriod time.Duration) *ProcessKiller {
	return &ProcessKiller{
		GracePeriod:             gracePeriod,
		SystemProcessProtection: true,
	}
}

// Kill attempts to terminate the process using the default grace period.
// It first tries SIGTERM, then SIGKILL if the process doesn't exit in time.
func (k *ProcessKiller) Kill(ctx context.Context, pid int, portInfo *models.PortInfo) (*KillResult, error) {
	return k.KillWithTimeout(ctx, pid, k.GracePeriod, portInfo)
}

// KillWithTimeout attempts to terminate a process with a custom grace period.
// The termination strategy:
// 1. Check if process is protected (system process)
// 2. Verify process is actually running
// 3. Send SIGTERM for graceful shutdown
// 4. Wait for process to exit within grace period
// 5. If still running after timeout, send SIGKILL
// 6. Return the result with method used and duration
func (k *ProcessKiller) KillWithTimeout(ctx context.Context, pid int, timeout time.Duration, portInfo *models.PortInfo) (*KillResult, error) {
	startTime := time.Now()

	// Protect system processes from accidental termination
	if k.SystemProcessProtection && portInfo != nil && portInfo.ShouldDisplayWarning() {
		return &KillResult{
			Success: false,
			Method:  KillMethodFailed,
			Message: fmt.Sprintf("System process (PID %d) is protected", pid),
		}, nil
	}

	// Verify process is running before attempting to kill
	running, err := k.IsRunning(pid)
	if err != nil {
		return &KillResult{
			Success: false,
			Method:  KillMethodFailed,
			Message: fmt.Sprintf("Process status check failed: %v", err),
		}, err
	}
	if !running {
		return &KillResult{
			Success: false,
			Method:  KillMethodFailed,
			Message: fmt.Sprintf("Process PID %d is not running", pid),
		}, nil
	}

	// Phase 1: Send SIGTERM for graceful shutdown
	if err := k.sendSignal(pid, syscall.SIGTERM); err != nil {
		return &KillResult{
			Success: false,
			Method:  KillMethodFailed,
			Message: fmt.Sprintf("SIGTERM send failed: %v", err),
		}, err
	}

	// Phase 2: Wait for graceful shutdown or force with SIGKILL
	graceCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-graceCtx.Done():
			// Grace period expired - force kill
			return k.forceKill(ctx, pid, startTime)
		case <-ticker.C:
			// Check if process has exited
			running, err := k.IsRunning(pid)
			if err == nil && !running {
				return &KillResult{
					Success:  true,
					Method:   KillMethodSIGTERM,
					Message:  fmt.Sprintf("PID %d terminated gracefully", pid),
					Duration: time.Since(startTime),
				}, nil
			}
		case <-ctx.Done():
			// Operation was cancelled from outside
			return &KillResult{
				Success: false,
				Method:  KillMethodFailed,
				Message: "Operation cancelled",
			}, ctx.Err()
		}
	}
}

// forceKill sends SIGKILL to terminate a process that didn't exit gracefully.
// This is the final step after SIGTERM fails. It verifies termination and returns result.
func (k *ProcessKiller) forceKill(_ context.Context, pid int, startTime time.Time) (*KillResult, error) {
	// Check one more time in case process exited between last check and timeout
	running, err := k.IsRunning(pid)
	if err == nil && !running {
		return &KillResult{
			Success:  true,
			Method:   KillMethodSIGTERM,
			Message:  fmt.Sprintf("PID %d terminated (timing difference)", pid),
			Duration: time.Since(startTime),
		}, nil
	}

	// Send SIGKILL to force termination
	if err := k.sendSignal(pid, syscall.SIGKILL); err != nil {
		return &KillResult{
			Success: false,
			Method:  KillMethodFailed,
			Message: fmt.Sprintf("SIGKILL send failed: %v", err),
		}, err
	}

	// Give the process a moment to terminate after SIGKILL
	time.Sleep(50 * time.Millisecond)
	running, _ = k.IsRunning(pid)

	result := &KillResult{
		Success:  !running,
		Method:   KillMethodSIGKILL,
		Message:  fmt.Sprintf("SIGKILL sent to PID %d", pid),
		Duration: time.Since(startTime),
	}

	// Set appropriate message based on final state
	if !running {
		result.Message = fmt.Sprintf("PID %d force terminated", pid)
		result.Success = true
	} else {
		result.Message = fmt.Sprintf("PID %d termination failed", pid)
		result.Success = false
	}

	return result, nil
}

// IsRunning checks if a process with the given PID is currently active.
// It uses signal 0 which doesn't actually send a signal but checks process existence.
// Returns false if process doesn't exist, true if it exists.
func (k *ProcessKiller) IsRunning(pid int) (bool, error) {
	err := k.sendSignal(pid, 0)
	if err != nil {
		return false, nil
	}
	return true, nil
}

// sendSignal sends a signal to the process with the given PID.
// Signal 0 is used to check process existence without actually signaling.
// Actual implementation is platform-specific (see killer_posix.go and killer_windows.go).
func (k *ProcessKiller) sendSignal(pid int, sig syscall.Signal) error {
	process, err := findProcess(pid)
	if err != nil {
		return err
	}
	defer process.Release()
	return process.Signal(sig)
}

// findProcess locates a process by PID and returns a Process handle.
// The actual implementation is platform-specific and defined in killer_posix.go or killer_windows.go.
func findProcess(pid int) (Process, error) {
	return findProcessImpl(pid)
}
