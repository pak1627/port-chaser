//go:build darwin || linux
// +build darwin linux

package process

import (
	"fmt"
	"os"
	"syscall"
)

// Process defines the interface for interacting with system processes on POSIX systems.
// This abstraction allows the killer to work with different process implementations.
type Process interface {
	// Signal sends a signal to the process (SIGTERM, SIGKILL, or 0 for checking existence)
	Signal(sig syscall.Signal) error
	// Release releases any resources associated with the process handle
	Release() error
}

// osProcess wraps the standard library's os.Process to implement our Process interface.
// It adds the Release method which is a no-op for os.Process since it doesn't hold resources.
type osProcess struct {
	*os.Process
}

// Release is a no-op for os.Process since it doesn't require explicit cleanup.
// This method exists to satisfy the Process interface.
func (p *osProcess) Release() error {
	return nil
}

// Signal sends a signal to the underlying OS process.
// This delegates to the standard library's Process.Signal method.
func (p *osProcess) Signal(sig syscall.Signal) error {
	return p.Process.Signal(sig)
}

// findProcessImpl is the POSIX implementation of process finding.
// It uses os.FindProcess which on POSIX systems doesn't actually verify the process exists
// but creates a Process object that can be used with Signal(0) to check existence.
func findProcessImpl(pid int) (Process, error) {
	process, err := os.FindProcess(pid)
	if err != nil {
		return nil, fmt.Errorf("find process failed (PID %d): %w", pid, err)
	}
	return &osProcess{Process: process}, nil
}
