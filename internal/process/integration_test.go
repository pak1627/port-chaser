//go:build !windows
// +build !windows

package process

import (
	"context"
	"os"
	"os/exec"
	"sync"
	"testing"
	"time"

	"github.com/manson/port-chaser/internal/models"
)

func spawnTestProcess(t *testing.T) (int, *exec.Cmd) {
	t.Helper()

	cmd := exec.Command("sleep", "30")
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start test process: %v", err)
	}

	return cmd.Process.Pid, cmd
}

func TestIntegration_SIGTERM(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test (-short)")
	}

	pid, cmd := spawnTestProcess(t)
	defer cmd.Wait()

	killer := NewProcessKiller()
	running, err := killer.IsRunning(pid)
	if err != nil {
		t.Fatalf("IsRunning failed: %v", err)
	}
	if !running {
		t.Fatal("test process is not running")
	}

	ctx := context.Background()
	portInfo := &models.PortInfo{
		PID:      pid,
		IsSystem: false,
	}

	result, err := killer.Kill(ctx, pid, portInfo)
	if err != nil {
		t.Errorf("Kill failed: %v", err)
	}

	if !result.Success {
		t.Errorf("termination failed: Method=%v, Message=%v", result.Method, result.Message)
	}

	if result.Method != KillMethodSIGTERM {
		t.Errorf("should terminate with SIGTERM: got Method=%v", result.Method)
	}

	time.Sleep(100 * time.Millisecond)
	running, _ = killer.IsRunning(pid)
	if running {
		t.Error("process is still running")
	}
}

func TestIntegration_SIGKILL(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test (-short)")
	}

	script := `
		trap 'echo SIGTERM ignored' TERM
		echo "ignore process started (PID: $$)"
		sleep 30
	`

	tmpFile, err := os.CreateTemp("", "sigterm-ignore-*.sh")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(script); err != nil {
		t.Fatalf("failed to write script: %v", err)
	}
	tmpFile.Close()

	if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
		t.Fatalf("failed to set execute permission: %v", err)
	}

	cmd := exec.Command(tmpFile.Name())
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start test process: %v", err)
	}
	pid := cmd.Process.Pid
	defer cmd.Wait()

	killer := NewProcessKillerWithGracePeriod(500 * time.Millisecond)
	ctx := context.Background()

	portInfo := &models.PortInfo{
		PID:      pid,
		IsSystem: false,
	}

	result, err := killer.Kill(ctx, pid, portInfo)
	if err != nil {
		t.Errorf("Kill failed: %v", err)
	}

	if result.Method != KillMethodSIGKILL {
		t.Logf("Method=%v (expected SIGKILL)", result.Method)
	}

	if !result.Success {
		t.Errorf("termination failed: Message=%v", result.Message)
	}

	time.Sleep(200 * time.Millisecond)
	running, _ := killer.IsRunning(pid)
	if running {
		t.Error("process is still running (SIGKILL failed)")
	}
}

func TestIntegration_ConcurrentKills(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test (-short)")
	}

	const numProcesses = 5
	processes := make([]struct {
		pid int
		cmd *exec.Cmd
	}, numProcesses)

	for i := 0; i < numProcesses; i++ {
		pid, cmd := spawnTestProcess(t)
		processes[i].pid = pid
		processes[i].cmd = cmd
	}
	defer func() {
		for _, p := range processes {
			p.cmd.Wait()
		}
	}()

	killer := NewProcessKiller()
	ctx := context.Background()

	var wg sync.WaitGroup
	results := make(chan *KillResult, numProcesses)

	for i := 0; i < numProcesses; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			portInfo := &models.PortInfo{
				PID:      processes[idx].pid,
				IsSystem: false,
			}
			result, _ := killer.Kill(ctx, processes[idx].pid, portInfo)
			results <- result
		}(i)
	}

	wg.Wait()
	close(results)

	successCount := 0
	for result := range results {
		if result.Success {
			successCount++
		}
	}

	if successCount != numProcesses {
		t.Errorf("some process terminations failed: success=%d, total=%d", successCount, numProcesses)
	}

	time.Sleep(200 * time.Millisecond)
	for _, p := range processes {
		running, _ := killer.IsRunning(p.pid)
		if running {
			t.Errorf("PID %d is still running", p.pid)
		}
	}
}

func TestIntegration_KillWithStorage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test (-short)")
	}

	pid, cmd := spawnTestProcess(t)
	defer cmd.Wait()

	killer := NewProcessKiller()
	ctx := context.Background()

	portInfo := &models.PortInfo{
		PortNumber:  3000,
		ProcessName: "sleep",
		PID:         pid,
		Command:     "sleep 30",
		IsSystem:    false,
	}

	result, err := killer.Kill(ctx, pid, portInfo)
	if err != nil {
		t.Errorf("Kill failed: %v", err)
	}

	if !result.Success {
		t.Errorf("termination failed: %v", result.Message)
	}

	t.Logf("process termination complete: Method=%v, Duration=%v", result.Method, result.Duration)
}

func TestIntegration_RealProcessScenario(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test (-short)")
	}

	scenarios := []struct {
		name    string
		portNum int
		command string
		args    []string
	}{
		{"HTTP server simulation", 8080, "sleep", []string{"30"}},
		{"database simulation", 5432, "sleep", []string{"30"}},
		{"application simulation", 3000, "sleep", []string{"30"}},
	}

	killer := NewProcessKiller()
	ctx := context.Background()

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			cmd := exec.Command(sc.command, sc.args...)
			if err := cmd.Start(); err != nil {
				t.Fatalf("failed to start process: %v", err)
			}
			defer cmd.Wait()

			pid := cmd.Process.Pid

			portInfo := &models.PortInfo{
				PortNumber:  sc.portNum,
				ProcessName: sc.command,
				PID:         pid,
				Command:     sc.command + " " + sc.args[0],
				IsSystem:    false,
			}

			running, _ := killer.IsRunning(pid)
			if !running {
				t.Fatal("process is not running")
			}

			result, err := killer.Kill(ctx, pid, portInfo)
			if err != nil {
				t.Errorf("Kill failed: %v", err)
			}

			if !result.Success {
				t.Errorf("termination failed: %v", result.Message)
			}

			time.Sleep(100 * time.Millisecond)
			running, _ = killer.IsRunning(pid)
			if running {
				t.Error("process is still running")
			}

			t.Logf("%s complete: Method=%v, Duration=%v", sc.name, result.Method, result.Duration)
		})
	}
}

func TestIntegration_ProcessNotFound(t *testing.T) {
	killer := NewProcessKiller()
	ctx := context.Background()

	portInfo := &models.PortInfo{
		PID:      9999999,
		IsSystem: false,
	}

	result, err := killer.Kill(ctx, portInfo.PID, portInfo)

	if result.Success {
		t.Error("non-existent process cannot succeed")
	}

	if result.Method != KillMethodFailed {
		t.Errorf("should be FAILED Method: got=%v", result.Method)
	}

	t.Logf("expected failure: Message=%v, Error=%v", result.Message, err)
}

func TestIntegration_PermissionDenied(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test (-short)")
	}

	killer := NewProcessKiller()
	ctx := context.Background()

	portInfo := &models.PortInfo{
		PID:      1,
		IsSystem: true,
	}

	result, err := killer.Kill(ctx, portInfo.PID, portInfo)

	if result.Success {
		t.Error("system process should be protected")
	}

	if result.Message == "" {
		t.Error("message is empty")
	}

	t.Logf("system process protection: Message=%v, Error=%v", result.Message, err)
}

func TestIntegration_DurationMeasurement(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test (-short)")
	}

	pid, cmd := spawnTestProcess(t)
	defer cmd.Wait()

	killer := NewProcessKillerWithGracePeriod(100 * time.Millisecond)
	ctx := context.Background()

	portInfo := &models.PortInfo{
		PID:      pid,
		IsSystem: false,
	}

	start := time.Now()
	result, _ := killer.Kill(ctx, pid, portInfo)
	totalElapsed := time.Since(start)

	if result.Duration == 0 {
		t.Error("Duration should be measured")
	}

	if result.Duration > totalElapsed {
		t.Errorf("Duration is greater than total time: Duration=%v, Total=%v",
			result.Duration, totalElapsed)
	}

	if result.Duration > 1*time.Second {
		t.Errorf("termination time too long: %v", result.Duration)
	}

	t.Logf("time measurement: Duration=%v, TotalElapsed=%v", result.Duration, totalElapsed)
}

func TestIntegration_ContextTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test (-short)")
	}

	pid, cmd := spawnTestProcess(t)
	defer cmd.Wait()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	killer := NewProcessKillerWithGracePeriod(10 * time.Second)
	portInfo := &models.PortInfo{
		PID:      pid,
		IsSystem: false,
	}

	start := time.Now()
	result, err := killer.Kill(ctx, pid, portInfo)
	elapsed := time.Since(start)

	if err == nil {
		t.Error("context timeout should return error")
	}

	if result.Method != KillMethodFailed {
		t.Errorf("should be FAILED on timeout: got=%v", result.Method)
	}

	if result.Message != "operation canceled" {
		t.Errorf("cancel message different from expected: got=%v", result.Message)
	}

	if elapsed > 200*time.Millisecond {
		t.Errorf("timeout too long: %v", elapsed)
	}

	t.Logf("context timeout: Elapsed=%v, Message=%v", elapsed, result.Message)
}

func TestIntegration_SignalZero(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test (-short)")
	}

	pid, cmd := spawnTestProcess(t)
	defer cmd.Wait()

	killer := NewProcessKiller()

	running, err := killer.IsRunning(pid)
	if err != nil {
		t.Errorf("IsRunning error: %v", err)
	}
	if !running {
		t.Error("running process reported as not running")
	}

	ctx := context.Background()
	portInfo := &models.PortInfo{PID: pid, IsSystem: false}
	killer.Kill(ctx, pid, portInfo)

	time.Sleep(100 * time.Millisecond)

	running, _ = killer.IsRunning(pid)
	if running {
		t.Error("terminated process reported as running")
	}
}

func TestIntegration_MultipleKillAttempts(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test (-short)")
	}

	pid, cmd := spawnTestProcess(t)
	defer cmd.Wait()

	killer := NewProcessKiller()
	ctx := context.Background()
	portInfo := &models.PortInfo{PID: pid, IsSystem: false}

	result1, err1 := killer.Kill(ctx, pid, portInfo)
	if !result1.Success {
		t.Errorf("first termination failed: %v", result1.Message)
	}

	time.Sleep(200 * time.Millisecond)

	result2, err2 := killer.Kill(ctx, pid, portInfo)

	if result2.Success {
		t.Error("already terminated process should fail")
	}

	t.Logf("first: Success=%v, Method=%v, Error=%v", result1.Success, result1.Method, err1)
	t.Logf("second: Success=%v, Method=%v, Message=%v, Error=%v",
		result2.Success, result2.Method, result2.Message, err2)
}

func TestIntegration_ZombieProcess(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test (-short)")
	}

	cmd := exec.Command("sh", "-c", "exit 0")
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start process: %v", err)
	}

	pid := cmd.Process.Pid

	killer := NewProcessKiller()

	running, err := killer.IsRunning(pid)
	t.Logf("zombie process state: Running=%v, Error=%v", running, err)

	cmd.Wait()

	running, _ = killer.IsRunning(pid)
	if running {
		t.Error("reported as running after cleanup")
	}
}

func BenchmarkIntegration_Kill(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping benchmark (-short)")
	}

	killer := NewProcessKiller()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cmd := exec.Command("sleep", "30")
		if err := cmd.Start(); err != nil {
			b.Fatalf("failed to start process: %v", err)
		}
		pid := cmd.Process.Pid

		portInfo := &models.PortInfo{PID: pid, IsSystem: false}
		killer.Kill(ctx, pid, portInfo)

		cmd.Wait()
	}
}
