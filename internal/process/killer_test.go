package process

import (
	"context"
	"testing"
	"time"

	"github.com/manson/port-chaser/internal/models"
)

func TestProcessKiller_NewProcessKiller(t *testing.T) {
	killer := NewProcessKiller()

	if killer == nil {
		t.Fatal("NewProcessKiller returned nil")
	}

	if killer.GracePeriod != 3*time.Second {
		t.Errorf("Default GracePeriod is not 3s: got=%v", killer.GracePeriod)
	}

	if !killer.SystemProcessProtection {
		t.Error("SystemProcessProtection should be true by default")
	}
}

func TestProcessKiller_NewProcessKillerWithGracePeriod(t *testing.T) {
	customGracePeriod := 5 * time.Second
	killer := NewProcessKillerWithGracePeriod(customGracePeriod)

	if killer.GracePeriod != customGracePeriod {
		t.Errorf("GracePeriod mismatch: got=%v, want=%v", killer.GracePeriod, customGracePeriod)
	}
}

func TestProcessKiller_SystemProcessProtection(t *testing.T) {
	tests := []struct {
		name        string
		pid         int
		portInfo    *models.PortInfo
		wantSuccess bool
		wantMethod  KillMethod
	}{
		{
			name: "system critical process (PID 1)",
			pid:  1,
			portInfo: &models.PortInfo{
				PID:      1,
				IsSystem: true,
			},
			wantSuccess: false,
			wantMethod:  KillMethodFailed,
		},
		{
			name: "low PID process (PID 50)",
			pid:  50,
			portInfo: &models.PortInfo{
				PID:      50,
				IsSystem: false,
			},
			wantSuccess: false,
			wantMethod:  KillMethodFailed,
		},
		{
			name: "normal user process",
			pid:  1234,
			portInfo: &models.PortInfo{
				PID:      1234,
				IsSystem: false,
			},
			wantSuccess: false,
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

			if tt.pid < 100 || tt.portInfo.IsSystem {
				if result.Method != KillMethodFailed {
					t.Errorf("System process should be protected: got Method = %v, want %v", result.Method, KillMethodFailed)
				}
			}
		})
	}
}

func TestProcessKiller_IsRunning(t *testing.T) {
	killer := NewProcessKiller()

	tests := []struct {
		name    string
		pid     int
		wantErr bool
	}{
		{
			name:    "current process (always running)",
			pid:     0,
			wantErr: false,
		},
		{
			name:    "non-existent PID (high number)",
			pid:     999999999,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.pid == 0 {
				t.Skip("current process test skipped")
				return
			}

			running, err := killer.IsRunning(tt.pid)

			if (err != nil) != tt.wantErr {
				t.Errorf("IsRunning() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.pid == 999999999 && running {
				t.Error("Non-existent PID should not be running")
			}
		})
	}
}

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

func TestProcessKiller_Timeout(t *testing.T) {
	killer := NewProcessKillerWithGracePeriod(100 * time.Millisecond)
	ctx := context.Background()

	portInfo := &models.PortInfo{
		PID:      999998,
		IsSystem: false,
	}

	_, err := killer.Kill(ctx, portInfo.PID, portInfo)
	if err != nil {
		t.Logf("Non-existent process kill attempt: %v", err)
	}
}

func TestProcessKiller_ContextCancellation(t *testing.T) {
	killer := NewProcessKiller()
	ctx, cancel := context.WithCancel(context.Background())

	cancel()

	portInfo := &models.PortInfo{
		PID:      1234,
		IsSystem: false,
	}

	result, err := killer.Kill(ctx, portInfo.PID, portInfo)

	if result.Method != KillMethodFailed {
		t.Errorf("Cancelled context should return FAILED: got=%v", result.Method)
	}

	if err == nil {
		t.Error("Cancelled context should return error")
	}

	if result.Message != "Operation cancelled" {
		t.Errorf("Cancel message mismatch: got=%v", result.Message)
	}
}

func Benchmark_Kill(b *testing.B) {
	killer := NewProcessKiller()
	ctx := context.Background()
	portInfo := &models.PortInfo{
		PID:      999999,
		IsSystem: false,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		killer.Kill(ctx, portInfo.PID, portInfo)
	}
}

func Benchmark_IsRunning(b *testing.B) {
	killer := NewProcessKiller()
	pid := 999999

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		killer.IsRunning(pid)
	}
}

func TestProcessKiller_NilPortInfo(t *testing.T) {
	killer := NewProcessKiller()
	ctx := context.Background()

	result, err := killer.Kill(ctx, 999997, nil)

	if result.Success {
		t.Error("Non-existent process should fail")
	}

	if err != nil {
		t.Logf("Expected error: %v", err)
	}
}

func TestProcessKiller_KillWithTimeout(t *testing.T) {
	tests := []struct {
		name          string
		timeout       time.Duration
		expectSIGKILL bool
	}{
		{
			name:          "very short timeout (immediate SIGKILL)",
			timeout:       1 * time.Millisecond,
			expectSIGKILL: true,
		},
		{
			name:          "long timeout",
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

			if err == nil && result.Success {
				t.Error("Non-existent process cannot succeed")
			}

			t.Logf("Timeout %v: Method=%v, Message=%v", tt.timeout, result.Method, result.Message)
		})
	}
}

func TestProcessKiller_SystemProcessProtectionDisabled(t *testing.T) {
	killer := NewProcessKiller()
	killer.SystemProcessProtection = false

	ctx := context.Background()
	portInfo := &models.PortInfo{
		PID:      1,
		IsSystem: true,
	}

	result, err := killer.Kill(ctx, portInfo.PID, portInfo)

	if result.Message == "System process (PID 1) is protected" {
		t.Error("Protection disabled should not show protection message")
	}

	t.Logf("Protection disabled result: Success=%v, Method=%v, Message=%v, Error=%v",
		result.Success, result.Method, result.Message, err)
}

func TestProcessKiller_NotRunning(t *testing.T) {
	killer := NewProcessKiller()
	ctx := context.Background()

	portInfo := &models.PortInfo{
		PID:      999995,
		IsSystem: false,
	}

	result, err := killer.Kill(ctx, portInfo.PID, portInfo)

	if result.Success {
		t.Error("Non-existent process cannot succeed")
	}

	if result.Message == "" {
		t.Error("Message is empty")
	}

	t.Logf("Message: %v, Error: %v", result.Message, err)
}

func TestProcessKiller_ResultFields(t *testing.T) {
	killer := NewProcessKiller()
	ctx := context.Background()

	portInfo := &models.PortInfo{
		PID:      999994,
		IsSystem: false,
	}

	startTime := time.Now()
	result, _ := killer.Kill(ctx, portInfo.PID, portInfo)

	if result.Duration == 0 {
		t.Error("Duration should be measured")
	}

	validMethods := map[KillMethod]bool{
		KillMethodSIGTERM: true,
		KillMethodSIGKILL: true,
		KillMethodFailed:  true,
	}
	if !validMethods[result.Method] {
		t.Errorf("Invalid Method: %v", result.Method)
	}

	elapsed := time.Since(startTime)
	if elapsed > 5*time.Second {
		t.Errorf("Termination took too long: %v", elapsed)
	}

	t.Logf("Result: Success=%v, Method=%v, Duration=%v", result.Success, result.Method, result.Duration)
}

func TestProcessKiller_EdgeCase_PIDZero(t *testing.T) {
	killer := NewProcessKiller()
	ctx := context.Background()

	portInfo := &models.PortInfo{
		PID:      0,
		IsSystem: false,
	}

	result, err := killer.Kill(ctx, portInfo.PID, portInfo)

	t.Logf("PID 0 result: Success=%v, Method=%v, Message=%v, Error=%v",
		result.Success, result.Method, result.Message, err)
}

func TestProcessKiller_EdgeCase_NegativePID(t *testing.T) {
	killer := NewProcessKiller()
	ctx := context.Background()

	portInfo := &models.PortInfo{
		PID:      -1,
		IsSystem: false,
	}

	result, err := killer.Kill(ctx, portInfo.PID, portInfo)

	if err == nil {
		t.Error("Negative PID should return error")
	}

	if result.Method != KillMethodFailed {
		t.Errorf("Negative PID should have FAILED Method: got=%v", result.Method)
	}

	t.Logf("Negative PID result: Message=%v, Error=%v", result.Message, err)
}

func TestProcessKiller_SIGTERMThenSIGKILL(t *testing.T) {
	killer := NewProcessKillerWithGracePeriod(200 * time.Millisecond)
	ctx := context.Background()

	portInfo := &models.PortInfo{
		PID:      999993,
		IsSystem: false,
	}

	result, _ := killer.Kill(ctx, portInfo.PID, portInfo)

	t.Logf("After SIGTERM: Method=%v, Message=%v", result.Method, result.Message)
}

func TestProcessKiller_CompetitionCondition(t *testing.T) {
	killer := NewProcessKiller()
	ctx := context.Background()

	portInfo := &models.PortInfo{
		PID:      999992,
		IsSystem: false,
	}

	for i := 0; i < 3; i++ {
		result, err := killer.Kill(ctx, portInfo.PID, portInfo)

		t.Logf("Attempt %d: Success=%v, Method=%v, Message=%v, Error=%v",
			i+1, result.Success, result.Method, result.Message, err)
	}
}

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
			name:             "short wait time",
			gracePeriod:      100 * time.Millisecond,
			expectedMaxDelay: 500 * time.Millisecond,
		},
		{
			name:             "long wait time",
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

			if result.Duration > elapsed {
				t.Errorf("Duration > actual elapsed: Duration=%v, Elapsed=%v",
					result.Duration, elapsed)
			}

			if elapsed > tt.expectedMaxDelay {
				t.Errorf("Elapsed time too long: Expected<%v, Got=%v",
					tt.expectedMaxDelay, elapsed)
			}

			t.Logf("GracePeriod=%v: Duration=%v, ActualElapsed=%v",
				tt.gracePeriod, result.Duration, elapsed)
		})
	}
}
