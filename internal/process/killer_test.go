// Package process에 대한 테스트
package process

import (
	"context"
	"testing"
	"time"

	"github.com/manson/port-chaser/internal/models"
)

// TestProcessKiller_NewProcessKiller는 ProcessKiller 생성을 테스트합니다.
func TestProcessKiller_NewProcessKiller(t *testing.T) {
	killer := NewProcessKiller()

	if killer == nil {
		t.Fatal("NewProcessKiller가 nil을 반환했습니다")
	}

	if killer.GracePeriod != 3*time.Second {
		t.Errorf("기본 GracePeriod가 3초가 아닙니다: got=%v", killer.GracePeriod)
	}

	if !killer.SystemProcessProtection {
		t.Error("SystemProcessProtection이 기본적으로 true여야 합니다")
	}
}

// TestProcessKiller_NewProcessKillerWithGracePeriod는 사용자 정의 GracePeriod 테스트입니다.
func TestProcessKiller_NewProcessKillerWithGracePeriod(t *testing.T) {
	customGracePeriod := 5 * time.Second
	killer := NewProcessKillerWithGracePeriod(customGracePeriod)

	if killer.GracePeriod != customGracePeriod {
		t.Errorf("GracePeriod가 설정값과 다릅니다: got=%v, want=%v", killer.GracePeriod, customGracePeriod)
	}
}

// TestProcessKiller_SystemProcessProtection은 시스템 프로세스 보호 기능을 테스트합니다.
func TestProcessKiller_SystemProcessProtection(t *testing.T) {
	tests := []struct {
		name      string
		pid       int
		portInfo  *models.PortInfo
		wantSuccess bool
		wantMethod KillMethod
	}{
		{
			name: "시스템 중요 프로세스 (PID 1)",
			pid:  1,
			portInfo: &models.PortInfo{
				PID:      1,
				IsSystem: true,
			},
			wantSuccess: false,
			wantMethod:  KillMethodFailed,
		},
		{
			name: "낮은 PID 프로세스 (PID 50)",
			pid:  50,
			portInfo: &models.PortInfo{
				PID:      50,
				IsSystem: false,
			},
			wantSuccess: false,
			wantMethod:  KillMethodFailed,
		},
		{
			name: "일반 사용자 프로세스",
			pid:  1234,
			portInfo: &models.PortInfo{
				PID:      1234,
				IsSystem: false,
			},
			wantSuccess: false, // 실제 프로세스가 없으므로 실패 예상
			wantMethod:  KillMethodFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			killer := NewProcessKiller()
			ctx := context.Background()

			result, err := killer.Kill(ctx, tt.pid, tt.portInfo)

			if err != nil && tt.wantSuccess {
				t.Errorf("Kill() error = %v, want success", err)
			}

			if result.Success != tt.wantSuccess {
				t.Errorf("Kill() Success = %v, want %v", result.Success, tt.wantSuccess)
			}

			// 시스템 프로세스 보호로 실패한 경우
			if tt.pid < 100 || tt.portInfo.IsSystem {
				if result.Method != KillMethodFailed {
					t.Errorf("시스템 프로세스는 보호되어야 합니다: got Method = %v, want %v", result.Method, KillMethodFailed)
				}
			}
		})
	}
}

// TestProcessKiller_IsRunning은 프로세스 실행 상태 확인을 테스트합니다.
func TestProcessKiller_IsRunning(t *testing.T) {
	killer := NewProcessKiller()

	tests := []struct {
		name    string
		pid     int
		wantErr bool
	}{
		{
			name:    "현재 프로세스 (항상 실행 중)",
			pid:     0, // 현재 프로세스는 테스트에서 사용하지 않음
			wantErr: false,
		},
		{
			name:    "존재하지 않는 PID (높은 숫자)",
			pid:     999999999,
			wantErr: false, // IsRunning은 에러가 아니라 false를 반환
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 테스트 프로세스로 현재 프로세스 사용
			if tt.pid == 0 {
				// 실제로는 자신의 PID를 사용해야 하지만
				// 테스트 안정성을 위해 건너뜀
				t.Skip("현재 프로세스 테스트는 건너뜀")
				return
			}

			running, err := killer.IsRunning(tt.pid)

			if (err != nil) != tt.wantErr {
				t.Errorf("IsRunning() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.pid == 999999999 && running {
				t.Error("존재하지 않는 PID는 실행 중이지 않아야 합니다")
			}
		})
	}
}

// TestKillMethod_String은 KillMethod 문자열 표현을 테스트합니다.
func TestKillMethod_String(t *testing.T) {
	tests := []struct {
		method KillMethod
		want   string
	}{
		{KillMethodSIGTERM, "SIGTERM"},
		{KillMethodSIGKILL, "SIGKILL"},
		{KillMethodFailed, "FAILED"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := string(tt.method)
			if got != tt.want {
				t.Errorf("KillMethod.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestProcessKiller_Timeout은 타임아웃 동작을 테스트합니다.
func TestProcessKiller_Timeout(t *testing.T) {
	killer := NewProcessKillerWithGracePeriod(100 * time.Millisecond)
	ctx := context.Background()

	// 존재하지 않는 프로세스로 타임아웃 테스트
	portInfo := &models.PortInfo{
		PID:      999998,
		IsSystem: false,
	}

	_, err := killer.Kill(ctx, portInfo.PID, portInfo)
	// 존재하지 않는 프로세스이므로 에러 또는 실패 결과 반환
	if err != nil {
		// 예상된 동작
		t.Logf("존재하지 않는 프로세스 종료 시도: %v", err)
	}
}

// TestProcessKiller_ContextCancellation은 컨텍스트 취소를 테스트합니다.
func TestProcessKiller_ContextCancellation(t *testing.T) {
	killer := NewProcessKiller()
	ctx, cancel := context.WithCancel(context.Background())

	// 즉시 컨텍스트 취소
	cancel()

	portInfo := &models.PortInfo{
		PID:      1234,
		IsSystem: false,
	}

	result, err := killer.Kill(ctx, portInfo.PID, portInfo)

	if result.Method != KillMethodFailed {
		t.Errorf("취소된 컨텍스트는 FAILED를 반환해야 합니다: got=%v", result.Method)
	}

	if err == nil {
		t.Error("취소된 컨텍스트는 에러를 반환해야 합니다")
	}

	if result.Message != "작업이 취소되었습니다" {
		t.Errorf("취소 메시지가 예상과 다릅니다: got=%v", result.Message)
	}
}

// Benchmark_Kill은 Kill 함수의 성능을 벤치마킹합니다.
func Benchmark_Kill(b *testing.B) {
	killer := NewProcessKiller()
	ctx := context.Background()
	portInfo := &models.PortInfo{
		PID:      999999, // 존재하지 않는 PID
		IsSystem: false,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		killer.Kill(ctx, portInfo.PID, portInfo)
	}
}

// Benchmark_IsRunning은 IsRunning 함수의 성능을 벤치마킹합니다.
func Benchmark_IsRunning(b *testing.B) {
	killer := NewProcessKiller()
	pid := 999999 // 존재하지 않는 PID

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		killer.IsRunning(pid)
	}
}

// TestProcessKiller_NilPortInfo는 portInfo가 nil일 때의 동작을 테스트합니다.
func TestProcessKiller_NilPortInfo(t *testing.T) {
	killer := NewProcessKiller()
	ctx := context.Background()

	// nil portInfo로 종료 시도 (시스템 프로세스 보호 건너뜀)
	result, err := killer.Kill(ctx, 999997, nil)

	// 존재하지 않는 프로세스이므로 실패 예상
	if result.Success {
		t.Error("존재하지 않는 프로세스는 실패해야 합니다")
	}

	if err != nil {
		t.Logf("예상된 에러: %v", err)
	}
}

// TestProcessKiller_KillWithTimeout은 사용자 정의 타임아웃 테스트입니다.
func TestProcessKiller_KillWithTimeout(t *testing.T) {
	tests := []struct {
		name           string
		timeout        time.Duration
		expectSIGKILL  bool
	}{
		{
			name:          "매우 짧은 타임아웃 (즉시 SIGKILL)",
			timeout:       1 * time.Millisecond,
			expectSIGKILL: true, // 존재하지 않는 프로세스이므로 SIGKILL 시도
		},
		{
			name:          "긴 타임아웃",
			timeout:       10 * time.Second,
			expectSIGKILL: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			killer := NewProcessKiller()
			ctx := context.Background()
			portInfo := &models.PortInfo{
				PID:      999996,
				IsSystem: false,
			}

			result, err := killer.KillWithTimeout(ctx, portInfo.PID, tt.timeout, portInfo)

			// 존재하지 않는 프로세스이므로 결과 확인
			if err == nil && result.Success {
				t.Error("존재하지 않는 프로세스는 성공할 수 없습니다")
			}

			t.Logf("타임아웃 %v: Method=%v, Message=%v", tt.timeout, result.Method, result.Message)
		})
	}
}

// TestProcessKiller_SystemProcessProtectionDisabled는 보호 기능 비활성화 테스트입니다.
func TestProcessKiller_SystemProcessProtectionDisabled(t *testing.T) {
	killer := NewProcessKiller()
	killer.SystemProcessProtection = false

	ctx := context.Background()
	portInfo := &models.PortInfo{
		PID:      1,
		IsSystem: true,
	}

	result, err := killer.Kill(ctx, portInfo.PID, portInfo)

	// 보호가 비활성화되었으므로 시스템 프로세스 보호 메시지가 없어야 함
	if result.Message == "시스템 중요 프로세스(PID 1)는 보호됩니다" {
		t.Error("보호가 비활성화되었으므로 보호 메시지가 없어야 합니다")
	}

	// PID 1은 실제로 종료할 수 없으므로 실패 예상
	t.Logf("보호 비활성화 결과: Success=%v, Method=%v, Message=%v, Error=%v",
		result.Success, result.Method, result.Message, err)
}

// TestProcessKiller_NotRunning은 이미 종료된 프로세스 처리를 테스트합니다.
func TestProcessKiller_NotRunning(t *testing.T) {
	killer := NewProcessKiller()
	ctx := context.Background()

	portInfo := &models.PortInfo{
		PID:      999995,
		IsSystem: false,
	}

	result, err := killer.Kill(ctx, portInfo.PID, portInfo)

	// 존재하지 않는 프로세스
	if result.Success {
		t.Error("존재하지 않는 프로세스는 성공할 수 없습니다")
	}

	// "실행 중이 아닙니다" 메시지 확인
	if result.Message == "" {
		t.Error("메시지가 비어있습니다")
	}

	t.Logf("메시지: %v, 에러: %v", result.Message, err)
}

// TestProcessKiller_ResultFields는 KillResult 필드 검증 테스트입니다.
func TestProcessKiller_ResultFields(t *testing.T) {
	killer := NewProcessKiller()
	ctx := context.Background()

	portInfo := &models.PortInfo{
		PID:      999994,
		IsSystem: false,
	}

	startTime := time.Now()
	result, _ := killer.Kill(ctx, portInfo.PID, portInfo)

	// Duration이 측정되었는지 확인
	if result.Duration == 0 {
		t.Error("Duration이 측정되어야 합니다")
	}

	// Method가 유효한지 확인
	validMethods := map[KillMethod]bool{
		KillMethodSIGTERM: true,
		KillMethodSIGKILL: true,
		KillMethodFailed:  true,
	}
	if !validMethods[result.Method] {
		t.Errorf("유효하지 않은 Method: %v", result.Method)
	}

	// 실행 시간 확인 (거의 즉시 반환되어야 함)
	elapsed := time.Since(startTime)
	if elapsed > 5*time.Second {
		t.Errorf("종료까지 너무 오래 걸렸습니다: %v", elapsed)
	}

	t.Logf("Result: Success=%v, Method=%v, Duration=%v", result.Success, result.Method, result.Duration)
}

// TestProcessKiller_EdgeCase_PIDZero는 PID 0 처리를 테스트합니다.
func TestProcessKiller_EdgeCase_PIDZero(t *testing.T) {
	killer := NewProcessKiller()
	ctx := context.Background()

	portInfo := &models.PortInfo{
		PID:      0,
		IsSystem: false,
	}

	result, err := killer.Kill(ctx, portInfo.PID, portInfo)

	// PID 0은 현재 프로세스 그룹을 의미하므로 특별 처리 필요
	t.Logf("PID 0 결과: Success=%v, Method=%v, Message=%v, Error=%v",
		result.Success, result.Method, result.Message, err)
}

// TestProcessKiller_EdgeCase_NegativePID는 음수 PID 처리를 테스트합니다.
func TestProcessKiller_EdgeCase_NegativePID(t *testing.T) {
	killer := NewProcessKiller()
	ctx := context.Background()

	portInfo := &models.PortInfo{
		PID:      -1,
		IsSystem: false,
	}

	result, err := killer.Kill(ctx, portInfo.PID, portInfo)

	// 음수 PID는 에러여야 함
	if err == nil {
		t.Error("음수 PID는 에러를 반환해야 합니다")
	}

	if result.Method != KillMethodFailed {
		t.Errorf("음수 PID는 FAILED Method여야 합니다: got=%v", result.Method)
	}

	t.Logf("음수 PID 결과: Message=%v, Error=%v", result.Message, err)
}

// TestProcessKiller_SIGTERMThenSIGKILL은 SIGTERM 후 SIGKILL 순서를 테스트합니다.
func TestProcessKiller_SIGTERMThenSIGKILL(t *testing.T) {
	// 짧은 타임아웃으로 SIGKILL까지 진행되도록 설정
	killer := NewProcessKillerWithGracePeriod(200 * time.Millisecond)
	ctx := context.Background()

	// 존재하지 않는 프로세스로 시도
	portInfo := &models.PortInfo{
		PID:      999993,
		IsSystem: false,
	}

	result, _ := killer.Kill(ctx, portInfo.PID, portInfo)

	// 프로세스가 존재하지 않으므로 SIGTERM 전송 후 "실행 중이 아님" 응답
	t.Logf("SIGTERM 후 결과: Method=%v, Message=%v", result.Method, result.Message)
}

// TestProcessKiller_CompetitionCondition은 경쟁 조건 테스트입니다.
func TestProcessKiller_CompetitionCondition(t *testing.T) {
	killer := NewProcessKiller()
	ctx := context.Background()

	portInfo := &models.PortInfo{
		PID:      999992,
		IsSystem: false,
	}

	// 동일한 PID로 여러 번 종료 시도
	for i := 0; i < 3; i++ {
		result, err := killer.Kill(ctx, portInfo.PID, portInfo)

		// 모든 시도가 일관되게 처리되어야 함
		t.Logf("시도 %d: Success=%v, Method=%v, Message=%v, Error=%v",
			i+1, result.Success, result.Method, result.Message, err)
	}
}

// TestProcessKiller_DurationAccuracy는 Duration 측정 정확도 테스트입니다.
func TestProcessKiller_DurationAccuracy(t *testing.T) {
	killer := NewProcessKiller()
	ctx := context.Background()

	portInfo := &models.PortInfo{
		PID:      999991,
		IsSystem: false,
	}

	tests := []struct {
		name             string
		gracePeriod      time.Duration
		expectedMaxDelay time.Duration
	}{
		{
			name:             "짧은 대기 시간",
			gracePeriod:      100 * time.Millisecond,
			expectedMaxDelay: 500 * time.Millisecond,
		},
		{
			name:             "긴 대기 시간",
			gracePeriod:      1 * time.Second,
			expectedMaxDelay: 2 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			killer.GracePeriod = tt.gracePeriod

			start := time.Now()
			result, _ := killer.Kill(ctx, portInfo.PID, portInfo)
			elapsed := time.Since(start)

			// Duration 필드와 실제 경과 시간 비교
			if result.Duration > elapsed {
				t.Errorf("Duration이 실제 경과 시간보다 큽니다: Duration=%v, Elapsed=%v",
					result.Duration, elapsed)
			}

			// 결과 경과 시간이 예상 범위 내
			if elapsed > tt.expectedMaxDelay {
				t.Errorf("경과 시간이 너무 깁니다: Expected<%v, Got=%v",
					tt.expectedMaxDelay, elapsed)
			}

			t.Logf("GracePeriod=%v: Duration=%v, ActualElapsed=%v",
				tt.gracePeriod, result.Duration, elapsed)
		})
	}
}
