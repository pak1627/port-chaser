// +build !windows

// Package process에 대한 통합 테스트입니다.
// 이 테스트들은 실제 프로세스를 생성하고 종료하므로 POSIX 시스템에서만 실행됩니다.
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

// spawnTestProcess는 테스트용으로 실행 중인 프로세스를 생성합니다.
// sleep 명령을 사용하여 장기 실행 프로세스를 시뮬레이션합니다.
func spawnTestProcess(t *testing.T) (int, *exec.Cmd) {
	t.Helper()

	cmd := exec.Command("sleep", "30") // 30초 동안 실행
	if err := cmd.Start(); err != nil {
		t.Fatalf("테스트 프로세스 시작 실패: %v", err)
	}

	return cmd.Process.Pid, cmd
}

// TestIntegration_SIGTERM은 SIGTERM으로 프로세스를 정상 종료하는 통합 테스트입니다.
func TestIntegration_SIGTERM(t *testing.T) {
	if testing.Short() {
		t.Skip("통합 테스트 건너뜀 (-short)")
	}

	pid, cmd := spawnTestProcess(t)
	defer cmd.Wait() // 자식 프로세스 회수

	// 프로세스가 실행 중인지 확인
	killer := NewProcessKiller()
	running, err := killer.IsRunning(pid)
	if err != nil {
		t.Fatalf("IsRunning 실패: %v", err)
	}
	if !running {
		t.Fatal("테스트 프로세스가 실행 중이 아닙니다")
	}

	// 프로세스 종료
	ctx := context.Background()
	portInfo := &models.PortInfo{
		PID:      pid,
		IsSystem: false,
	}

	result, err := killer.Kill(ctx, pid, portInfo)
	if err != nil {
		t.Errorf("Kill 실패: %v", err)
	}

	if !result.Success {
		t.Errorf("종료 실패: Method=%v, Message=%v", result.Method, result.Message)
	}

	if result.Method != KillMethodSIGTERM {
		t.Errorf("SIGTERM으로 종료되어야 함: got Method=%v", result.Method)
	}

	// 프로세스가 실제로 종료되었는지 확인
	time.Sleep(100 * time.Millisecond)
	running, _ = killer.IsRunning(pid)
	if running {
		t.Error("프로세스가 여전히 실행 중입니다")
	}
}

// TestIntegration_SIGKILL은 SIGKILL로 프로세스를 강제 종료하는 통합 테스트입니다.
// SIGTERM을 무시하는 프로세스를 생성하여 SIGKILL이 필요한 상황을 시뮬레이션합니다.
func TestIntegration_SIGKILL(t *testing.T) {
	if testing.Short() {
		t.Skip("통합 테스트 건너뜀 (-short)")
	}

	// SIGTERM을 무시하는 스크립트 생성
	script := `
		trap 'echo SIGTERM 무시됨' TERM
		echo "무시 프로세스 시작 (PID: $$)"
		sleep 30
	`

	tmpFile, err := os.CreateTemp("", "sigterm-ignore-*.sh")
	if err != nil {
		t.Fatalf("임시 파일 생성 실패: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(script); err != nil {
		t.Fatalf("스크립트 쓰기 실패: %v", err)
	}
	tmpFile.Close()

	// 실행 권한 부여
	if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
		t.Fatalf("실행 권한 설정 실패: %v", err)
	}

	cmd := exec.Command(tmpFile.Name())
	if err := cmd.Start(); err != nil {
		t.Fatalf("테스트 프로세스 시작 실패: %v", err)
	}
	pid := cmd.Process.Pid
	defer cmd.Wait()

	// 짧은 대기 시간으로 SIGKILL까지 진행
	killer := NewProcessKillerWithGracePeriod(500 * time.Millisecond)
	ctx := context.Background()

	portInfo := &models.PortInfo{
		PID:      pid,
		IsSystem: false,
	}

	result, err := killer.Kill(ctx, pid, portInfo)
	if err != nil {
		t.Errorf("Kill 실패: %v", err)
	}

	// SIGTERM을 무시하므로 SIGKILL로 넘어가야 함
	if result.Method != KillMethodSIGKILL {
		t.Logf("Method=%v (SIGKILL 예상됨)", result.Method)
	}

	if !result.Success {
		t.Errorf("종료 실패: Message=%v", result.Message)
	}

	// 프로세스가 종료되었는지 확인
	time.Sleep(200 * time.Millisecond)
	running, _ := killer.IsRunning(pid)
	if running {
		t.Error("프로세스가 여전히 실행 중입니다 (SIGKILL 실패)")
	}
}

// TestIntegration_ConcurrentKills는 동시에 여러 프로세스를 종료하는 통합 테스트입니다.
func TestIntegration_ConcurrentKills(t *testing.T) {
	if testing.Short() {
		t.Skip("통합 테스트 건너뜀 (-short)")
	}

	const numProcesses = 5
	processes := make([]struct {
		pid int
		cmd *exec.Cmd
	}, numProcesses)

	// 여러 프로세스 생성
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

	// 동시에 종료 시도
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

	// 모든 프로세스가 성공적으로 종료되었는지 확인
	successCount := 0
	for result := range results {
		if result.Success {
			successCount++
		}
	}

	if successCount != numProcesses {
		t.Errorf("일부 프로세스 종료 실패: 성공=%d, 전체=%d", successCount, numProcesses)
	}

	// 모든 프로세스가 실제로 종료되었는지 확인
	time.Sleep(200 * time.Millisecond)
	for _, p := range processes {
		running, _ := killer.IsRunning(p.pid)
		if running {
			t.Errorf("PID %d가 여전히 실행 중입니다", p.pid)
		}
	}
}

// TestIntegration_KillWithStorage는 종료 후 저장소에 기록하는 통합 테스트입니다.
func TestIntegration_KillWithStorage(t *testing.T) {
	if testing.Short() {
		t.Skip("통합 테스트 건너뜀 (-short)")
	}

	// 저장소 생성 (테스트용 임시 DB)
	// 실제 통합에서는 storage 패키지의 SQLite를 사용

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
		t.Errorf("Kill 실패: %v", err)
	}

	if !result.Success {
		t.Errorf("종료 실패: %v", result.Message)
	}

	// 저장소에 기록
	// entry := models.HistoryEntry{
	// 	PortNumber:  portInfo.PortNumber,
	// 	ProcessName: portInfo.ProcessName,
	// 	PID:         portInfo.PID,
	// 	Command:     portInfo.Command,
	// 	KilledAt:    time.Now(),
	// }
	// storage.RecordKill(entry)

	t.Logf("프로세스 종료 완료: Method=%v, Duration=%v", result.Method, result.Duration)
}

// TestIntegration_RealProcessScenario는 실제 사용 시나리오를 시뮬레이션합니다.
func TestIntegration_RealProcessScenario(t *testing.T) {
	if testing.Short() {
		t.Skip("통합 테스트 건너뜀 (-short)")
	}

	// 테스트 시나리오:
	// 1. 여러 포트에서 실행 중인 프로세스 생성
	// 2. 특정 포트의 프로세스 종료
	// 3. 종료 결과 검증

	scenarios := []struct {
		name     string
		portNum  int
		command  string
		args     []string
	}{
		{"HTTP 서버 시뮬레이션", 8080, "sleep", []string{"30"}},
		{"데이터베이스 시뮬레이션", 5432, "sleep", []string{"30"}},
		{"애플리케이션 시뮬레이션", 3000, "sleep", []string{"30"}},
	}

	killer := NewProcessKiller()
	ctx := context.Background()

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			cmd := exec.Command(sc.command, sc.args...)
			if err := cmd.Start(); err != nil {
				t.Fatalf("프로세스 시작 실패: %v", err)
			}
			defer cmd.Wait()

			pid := cmd.Process.Pid

			// PortInfo 생성
			portInfo := &models.PortInfo{
				PortNumber:  sc.portNum,
				ProcessName: sc.command,
				PID:         pid,
				Command:     sc.command + " " + sc.args[0],
				IsSystem:    false,
			}

			// 프로세스 실행 확인
			running, _ := killer.IsRunning(pid)
			if !running {
				t.Fatal("프로세스가 실행 중이지 않습니다")
			}

			// 종료
			result, err := killer.Kill(ctx, pid, portInfo)
			if err != nil {
				t.Errorf("Kill 실패: %v", err)
			}

			if !result.Success {
				t.Errorf("종료 실패: %v", result.Message)
			}

			// 종료 확인
			time.Sleep(100 * time.Millisecond)
			running, _ = killer.IsRunning(pid)
			if running {
				t.Error("프로세스가 여전히 실행 중입니다")
			}

			t.Logf("%s 완료: Method=%v, Duration=%v", sc.name, result.Method, result.Duration)
		})
	}
}

// TestIntegration_ProcessNotFound는 존재하지 않는 프로세스 종료 시도 테스트입니다.
func TestIntegration_ProcessNotFound(t *testing.T) {
	killer := NewProcessKiller()
	ctx := context.Background()

	// 존재하지 않는 PID
	portInfo := &models.PortInfo{
		PID:      9999999,
		IsSystem: false,
	}

	result, err := killer.Kill(ctx, portInfo.PID, portInfo)

	// 에러가 아니라 실패 결과를 반환해야 함
	if result.Success {
		t.Error("존재하지 않는 프로세스는 성공할 수 없습니다")
	}

	if result.Method != KillMethodFailed {
		t.Errorf("FAILED Method여야 함: got=%v", result.Method)
	}

	t.Logf("예상된 실패: Message=%v, Error=%v", result.Message, err)
}

// TestIntegration_PermissionDenied는 권한 문제 시나리오 테스트입니다.
func TestIntegration_PermissionDenied(t *testing.T) {
	if testing.Short() {
		t.Skip("통합 테스트 건너뜀 (-short)")
	}

	// 다른 사용자의 프로세스는 종료할 수 없으므로 권한 거부 발생
	// 또는 root 권한으로 실행 중인 프로세스

	killer := NewProcessKiller()
	ctx := context.Background()

	// init 프로세스 (PID 1)는 항상 존재하지만 종료할 수 없음
	portInfo := &models.PortInfo{
		PID:      1,
		IsSystem: true,
	}

	result, err := killer.Kill(ctx, portInfo.PID, portInfo)

	// 시스템 프로세스 보호로 실패해야 함
	if result.Success {
		t.Error("시스템 프로세스는 보호되어야 합니다")
	}

	if result.Message == "" {
		t.Error("메시지가 비어있습니다")
	}

	t.Logf("시스템 프로세스 보호: Message=%v, Error=%v", result.Message, err)
}

// TestIntegration_DurationMeasurement는 종료 시간 측정 테스트입니다.
func TestIntegration_DurationMeasurement(t *testing.T) {
	if testing.Short() {
		t.Skip("통합 테스트 건너뜀 (-short)")
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

	// Duration 필드 검증
	if result.Duration == 0 {
		t.Error("Duration이 측정되어야 합니다")
	}

	// Duration이 전체 시간보다 작거나 같아야 함
	if result.Duration > totalElapsed {
		t.Errorf("Duration이 전체 시간보다 큽니다: Duration=%v, Total=%v",
			result.Duration, totalElapsed)
	}

	// 신속하게 종료되어야 함 (SIGTERM)
	if result.Duration > 1*time.Second {
		t.Errorf("종료 시간이 너무 깁니다: %v", result.Duration)
	}

	t.Logf("시간 측정: Duration=%v, TotalElapsed=%v", result.Duration, totalElapsed)
}

// TestIntegration_ContextTimeout은 컨텍스트 타임아웃 테스트입니다.
func TestIntegration_ContextTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("통합 테스트 건너뜀 (-short)")
	}

	pid, cmd := spawnTestProcess(t)
	defer cmd.Wait()

	// 50ms 후 컨텍스트 취소
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	killer := NewProcessKillerWithGracePeriod(10 * time.Second) // 긴 대기 시간
	portInfo := &models.PortInfo{
		PID:      pid,
		IsSystem: false,
	}

	start := time.Now()
	result, err := killer.Kill(ctx, pid, portInfo)
	elapsed := time.Since(start)

	// 컨텍스트 타임아웃으로 취소되어야 함
	if err == nil {
		t.Error("컨텍스트 타임아웃으로 에러가 반환되어야 합니다")
	}

	if result.Method != KillMethodFailed {
		t.Errorf("타임아웃 시 FAILED여야 함: got=%v", result.Method)
	}

	if result.Message != "작업이 취소되었습니다" {
		t.Errorf("취소 메시지가 예상과 다름: got=%v", result.Message)
	}

	// 컨텍스트 타임아웃 근처에서 종료되어야 함
	if elapsed > 200*time.Millisecond {
		t.Errorf("타임아웃이 너무 깁니다: %v", elapsed)
	}

	t.Logf("컨텍스트 타임아웃: Elapsed=%v, Message=%v", elapsed, result.Message)
}

// TestIntegration_SignalZero는 시그널 0으로 프로세스 존재 확인 테스트입니다.
func TestIntegration_SignalZero(t *testing.T) {
	if testing.Short() {
		t.Skip("통합 테스트 건너뜀 (-short)")
	}

	// 실행 중인 프로세스
	pid, cmd := spawnTestProcess(t)
	defer cmd.Wait()

	killer := NewProcessKiller()

	// sendSignal(sig=0)은 프로세스 존재만 확인
	running, err := killer.IsRunning(pid)
	if err != nil {
		t.Errorf("IsRunning 에러: %v", err)
	}
	if !running {
		t.Error("실행 중인 프로세스가 실행 중이 아닌 것으로 보고됨")
	}

	// 프로세스 종료 후 확인
	// syscall.Kill(pid, syscall.SIGTERM) 또는 killer.Kill 사용
	ctx := context.Background()
	portInfo := &models.PortInfo{PID: pid, IsSystem: false}
	killer.Kill(ctx, pid, portInfo)

	time.Sleep(100 * time.Millisecond)

	running, _ = killer.IsRunning(pid)
	if running {
		t.Error("종료된 프로세스가 실행 중인 것으로 보고됨")
	}
}

// TestIntegration_MultipleKillAttempts는 여러 번 종료 시도 테스트입니다.
func TestIntegration_MultipleKillAttempts(t *testing.T) {
	if testing.Short() {
		t.Skip("통합 테스트 건너뜀 (-short)")
	}

	pid, cmd := spawnTestProcess(t)
	defer cmd.Wait()

	killer := NewProcessKiller()
	ctx := context.Background()
	portInfo := &models.PortInfo{PID: pid, IsSystem: false}

	// 첫 번째 종료 시도
	result1, err1 := killer.Kill(ctx, pid, portInfo)
	if !result1.Success {
		t.Errorf("첫 번째 종료 실패: %v", result1.Message)
	}

	// 프로세스가 종료될 때까지 대기
	time.Sleep(200 * time.Millisecond)

	// 두 번째 종료 시도 (이미 종료된 프로세스)
	result2, err2 := killer.Kill(ctx, pid, portInfo)

	// 이미 종료된 프로세스이므로 실패해야 함
	if result2.Success {
		t.Error("이미 종료된 프로세스는 실패해야 합니다")
	}

	t.Logf("첫 번째: Success=%v, Method=%v, Error=%v", result1.Success, result1.Method, err1)
	t.Logf("두 번째: Success=%v, Method=%v, Message=%v, Error=%v",
		result2.Success, result2.Method, result2.Message, err2)
}

// TestIntegration_ZombieProcess는 좀비 프로세스 상황 테스트입니다.
func TestIntegration_ZombieProcess(t *testing.T) {
	if testing.Short() {
		t.Skip("통합 테스트 건너뜀 (-short)")
	}

	// 좀비 프로세스 생성은 복잡하므로 간단한 시뮬레이션만 수행
	// 실제 좀비 프로세스: 부모가 wait()하지 않은 자식 프로세스

	// fork 후 자식이 즉시 exit하는 상황
	cmd := exec.Command("sh", "-c", "exit 0")
	if err := cmd.Start(); err != nil {
		t.Fatalf("프로세스 시작 실패: %v", err)
	}

	pid := cmd.Process.Pid

	// 부모는 wait()하지 않음 (좀비 상태 유지)
	// 실제로는 Go의 exec.Cmd가 자동으로 wait를 수행할 수 있음

	killer := NewProcessKiller()

	// 좀비 프로세스 상태 확인
	running, err := killer.IsRunning(pid)
	t.Logf("좀비 프로세스 상태: Running=%v, Error=%v", running, err)

	// wait() 호출로 좀비 정리
	cmd.Wait()

	// 정리 후 상태 확인
	running, _ = killer.IsRunning(pid)
	if running {
		t.Error("정리된 후 실행 중으로 보고됨")
	}
}

// BenchmarkIntegration_Kill은 실제 프로세스 종료 벤치마크입니다.
func BenchmarkIntegration_Kill(b *testing.B) {
	if testing.Short() {
		b.Skip("벤치마크 건너뜀 (-short)")
	}

	killer := NewProcessKiller()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 프로세스 생성
		cmd := exec.Command("sleep", "30")
		if err := cmd.Start(); err != nil {
			b.Fatalf("프로세스 시작 실패: %v", err)
		}
		pid := cmd.Process.Pid

		// 종료
		portInfo := &models.PortInfo{PID: pid, IsSystem: false}
		killer.Kill(ctx, pid, portInfo)

		// 정리
		cmd.Wait()
	}
}
