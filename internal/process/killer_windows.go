//go:build windows
// +build windows

package process

import (
	"fmt"
	"os"
	"syscall"
)

// Process defines the interface for interacting with system processes on Windows.
// This abstraction allows the killer to work with different process implementations.
type Process interface {
	// Signal sends a signal to the process (SIGTERM or SIGKILL for termination)
	Signal(sig syscall.Signal) error
	// Release releases any resources associated with the process handle
	Release() error
}

// windowsProcess wraps the standard library's os.Process for Windows-specific behavior.
// On Windows, signals work differently than POSIX systems.
type windowsProcess struct {
	*os.Process
}

// Signal sends a signal to the Windows process.
// On Windows, SIGTERM and SIGKILL both map to process termination.
// Signal 0 for existence checking is not supported on Windows.
func (p *windowsProcess) Signal(sig syscall.Signal) error {
	if sig == syscall.SIGTERM || sig == syscall.SIGKILL {
		return p.Terminate()
	}
	return nil
}

// Terminate terminates the Windows process using the Windows API.
// It attempts to use kernel32.dll's TerminateProcess function.
func (p *windowsProcess) Terminate() error {
	dll, err := syscall.LoadDLL("kernel32.dll")
	if err != nil {
		return fmt.Errorf("kernel32.dll load failed: %w", err)
	}
	defer dll.Release()

	proc, err := dll.FindProc("TerminateProcess")
	if err != nil {
		return fmt.Errorf("TerminateProcess find failed: %w", err)
	}

	// Note: The proc variable is unused but kept for potential future use
	_ = proc
	return p.Kill()
}

// Release is a no-op for windowsProcess since os.Process doesn't require explicit cleanup.
// This method exists to satisfy the Process interface.
func (p *windowsProcess) Release() error {
	return nil
}

// findProcessImpl is the Windows implementation of process finding.
// It uses os.FindProcess which creates a Process object for the given PID.
func findProcessImpl(pid int) (Process, error) {
	process, err := os.FindProcess(pid)
	if err != nil {
		return nil, fmt.Errorf("find process failed (PID %d): %w", pid, err)
	}
	return &windowsProcess{Process: process}, nil
}
