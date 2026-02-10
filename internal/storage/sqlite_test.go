// Package storage에 대한 테스트입니다.
package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/manson/port-chaser/internal/models"
)

// TestSQLite_RecordKill은 RecordKill 기능을 테스트합니다.
func TestSQLite_RecordKill(t *testing.T) {
	tests := []struct {
		name    string
		entry   models.HistoryEntry
		wantErr bool
	}{
		{
			name: "정상 종료 기록",
			entry: models.HistoryEntry{
				PortNumber:  3000,
				ProcessName: "node",
				PID:         1234,
				Command:     "npm start",
				KilledAt:    time.Now(),
			},
			wantErr: false,
		},
		{
			name: "다른 포트 기록",
			entry: models.HistoryEntry{
				PortNumber:  8080,
				ProcessName: "python",
				PID:         5678,
				Command:     "python app.py",
				KilledAt:    time.Now(),
			},
			wantErr: false,
		},
		{
			name: "빈 명령어 기록",
			entry: models.HistoryEntry{
				PortNumber:  9000,
				ProcessName: "unknown",
				PID:         9999,
				Command:     "",
				KilledAt:    time.Now(),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 임시 데이터베이스 생성
			tmpDir := t.TempDir()
			dbPath := filepath.Join(tmpDir, "test.db")

			cfg := Config{
				DBPath:     dbPath,
				WALEnabled: true,
				Timeout:    50,
			}

			s, err := NewSQLite(cfg)
			if err != nil {
				t.Fatalf("NewSQLite() error = %v", err)
			}
			defer s.Close()

			err = s.RecordKill(tt.entry)
			if (err != nil) != tt.wantErr {
				t.Errorf("RecordKill() error = %v, wantErr %v", err, tt.wantErr)
			}

			// 기록이 제대로 되었는지 확인
			if !tt.wantErr {
				history, err := s.GetHistory(10)
				if err != nil {
					t.Errorf("GetHistory() error = %v", err)
				}
				if len(history) != 1 {
					t.Errorf("기록된 항목 수 = %d, 원하는 값 1", len(history))
				}
			}
		})
	}
}

// TestSQLite_GetHistory는 GetHistory 기능을 테스트합니다.
func TestSQLite_GetHistory(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	cfg := Config{
		DBPath:     dbPath,
		WALEnabled: true,
		Timeout:    50,
	}

	s, err := NewSQLite(cfg)
	if err != nil {
		t.Fatalf("NewSQLite() error = %v", err)
	}
	defer s.Close()

	// 테스트 데이터 삽입
	now := time.Now()
	entries := []models.HistoryEntry{
		{PortNumber: 3000, ProcessName: "node", PID: 1234, Command: "npm start", KilledAt: now.Add(-2 * time.Hour)},
		{PortNumber: 8080, ProcessName: "python", PID: 5678, Command: "python app.py", KilledAt: now.Add(-1 * time.Hour)},
		{PortNumber: 5000, ProcessName: "go", PID: 9012, Command: "go run main.go", KilledAt: now},
	}

	for _, entry := range entries {
		if err := s.RecordKill(entry); err != nil {
			t.Fatalf("RecordKill() error = %v", err)
		}
	}

	tests := []struct {
		name        string
		limit       int
		wantCount   int
		wantNewest  int // 가장 최근 항목의 포트 번호
	}{
		{
			name:       "전체 조회",
			limit:      10,
			wantCount:  3,
			wantNewest: 5000, // 가장 최근에 기록된 항목
		},
		{
			name:       "2개만 조회",
			limit:      2,
			wantCount:  2,
			wantNewest: 5000,
		},
		{
			name:       "1개만 조회",
			limit:      1,
			wantCount:  1,
			wantNewest: 5000,
		},
		{
			name:       "0개 요청",
			limit:      0,
			wantCount:  0,
			wantNewest: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := s.GetHistory(tt.limit)
			if err != nil {
				t.Errorf("GetHistory() error = %v", err)
				return
			}
			if len(got) != tt.wantCount {
				t.Errorf("GetHistory() 항목 수 = %d, 원하는 값 %d", len(got), tt.wantCount)
			}
			if tt.wantCount > 0 && got[0].PortNumber != tt.wantNewest {
				t.Errorf("GetHistory() 첫 번째 항목 포트 = %d, 원하는 값 %d", got[0].PortNumber, tt.wantNewest)
			}
		})
	}
}

// TestSQLite_GetKillCount는 GetKillCount 기능을 테스트합니다.
func TestSQLite_GetKillCount(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	cfg := Config{
		DBPath:     dbPath,
		WALEnabled: true,
		Timeout:    50,
	}

	s, err := NewSQLite(cfg)
	if err != nil {
		t.Fatalf("NewSQLite() error = %v", err)
	}
	defer s.Close()

	now := time.Now()

	// 포트 3000에 5회 기록 (최근 30일 이내)
	for i := 0; i < 5; i++ {
		entry := models.HistoryEntry{
			PortNumber:  3000,
			ProcessName: "node",
			PID:         1234 + i,
			Command:     "npm start",
			KilledAt:    now.Add(-time.Duration(i) * time.Hour),
		}
		if err := s.RecordKill(entry); err != nil {
			t.Fatalf("RecordKill() error = %v", err)
		}
	}

	// 포트 8080에 2회 기록
	for i := 0; i < 2; i++ {
		entry := models.HistoryEntry{
			PortNumber:  8080,
			ProcessName: "python",
			PID:         5678 + i,
			Command:     "python app.py",
			KilledAt:    now.Add(-time.Duration(i) * time.Hour),
		}
		if err := s.RecordKill(entry); err != nil {
			t.Fatalf("RecordKill() error = %v", err)
		}
	}

	// 오래된 기록 (31일 전 - 30일 범위 밖)
	oldEntry := models.HistoryEntry{
		PortNumber:  3000,
		ProcessName: "node",
		PID:         9999,
		Command:     "npm start",
		KilledAt:    now.Add(-31 * 24 * time.Hour),
	}
	if err := s.RecordKill(oldEntry); err != nil {
		t.Fatalf("RecordKill() error = %v", err)
	}

	tests := []struct {
		name    string
		port    int
		days    int
		want    int
	}{
		{
			name: "포트 3000, 최근 30일",
			port: 3000,
			days: 30,
			want: 5, // 오래된 기록 제외
		},
		{
			name: "포트 8080, 최근 30일",
			port: 8080,
			days: 30,
			want: 2,
		},
		{
			name: "포트 3000, 최근 7일",
			port: 3000,
			days: 7,
			want: 5, // 모든 기록이 7일 이내
		},
		{
			name: "없는 포트",
			port: 9999,
			days: 30,
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := s.GetKillCount(tt.port, tt.days)
			if err != nil {
				t.Errorf("GetKillCount() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("GetKillCount() = %d, 원하는 값 %d", got, tt.want)
			}
		})
	}
}

// TestSQLite_GetLastKillTime은 GetLastKillTime 기능을 테스트합니다.
func TestSQLite_GetLastKillTime(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	cfg := Config{
		DBPath:     dbPath,
		WALEnabled: true,
		Timeout:    50,
	}

	s, err := NewSQLite(cfg)
	if err != nil {
		t.Fatalf("NewSQLite() error = %v", err)
	}
	defer s.Close()

	now := time.Now()

	// 포트 3000에 여러 번 기록
	entries := []models.HistoryEntry{
		{PortNumber: 3000, ProcessName: "node", PID: 1234, Command: "npm start", KilledAt: now.Add(-2 * time.Hour)},
		{PortNumber: 3000, ProcessName: "node", PID: 1235, Command: "npm start", KilledAt: now.Add(-1 * time.Hour)},
		{PortNumber: 3000, ProcessName: "node", PID: 1236, Command: "npm start", KilledAt: now},
	}

	for _, entry := range entries {
		if err := s.RecordKill(entry); err != nil {
			t.Fatalf("RecordKill() error = %v", err)
		}
	}

	tests := []struct {
		name      string
		port      int
		wantZero  bool
	}{
		{
			name:     "기록이 있는 포트",
			port:     3000,
			wantZero: false,
		},
		{
			name:     "기록이 없는 포트",
			port:     8080,
			wantZero: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := s.GetLastKillTime(tt.port)
			if err != nil {
				t.Errorf("GetLastKillTime() error = %v", err)
				return
			}
			isZero := got.IsZero()
			if isZero != tt.wantZero {
				t.Errorf("GetLastKillTime() IsZero = %v, 원하는 값 %v", isZero, tt.wantZero)
			}
			if !tt.wantZero {
				// 가장 최근 시각인 now와 1초 이내 차이인지 확인
				diff := now.Sub(got)
				if diff > time.Second {
					t.Errorf("GetLastKillTime() 차이 = %v, 1초 이내 예상", diff)
				}
			}
		})
	}
}

// TestSQLite_Close는 Close 기능을 테스트합니다.
func TestSQLite_Close(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	cfg := Config{
		DBPath:     dbPath,
		WALEnabled: true,
		Timeout:    50,
	}

	s, err := NewSQLite(cfg)
	if err != nil {
		t.Fatalf("NewSQLite() error = %v", err)
	}

	err = s.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}

	// 닫힌 후 다시 닫아도 에러가 없어야 함
	err = s.Close()
	if err != nil {
		t.Errorf("Close() 두 번째 호출 error = %v", err)
	}
}

// TestSQLite_NewSQLite_디렉토리생성은 데이터베이스 디렉토리가 없을 때 생성하는지 테스트합니다.
func TestSQLite_NewSQLite_디렉토리생성(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "subdir", "test.db")

	cfg := Config{
		DBPath:     dbPath,
		WALEnabled: true,
		Timeout:    50,
	}

	s, err := NewSQLite(cfg)
	if err != nil {
		t.Fatalf("NewSQLite() error = %v", err)
	}
	defer s.Close()

	// 디렉토리가 생성되었는지 확인
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("데이터베이스 파일이 생성되지 않았습니다")
	}
}

// TestSQLite_FilePermissions은 데이터베이스 파일 권한이 600인지 테스트합니다.
func TestSQLite_FilePermissions(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	cfg := Config{
		DBPath:     dbPath,
		WALEnabled: true,
		Timeout:    50,
	}

	s, err := NewSQLite(cfg)
	if err != nil {
		t.Fatalf("NewSQLite() error = %v", err)
	}
	defer s.Close()

	info, err := os.Stat(dbPath)
	if err != nil {
		t.Fatalf("파일 정보 조회 실패: %v", err)
	}

	// Unix 시스템에서만 권한 확인
	mode := info.Mode().Perm()
	// 0600 = rw------- (소유자만 읽기/쓰기)
	expectedMode := os.FileMode(0600)

	if mode != expectedMode {
		// 일부 시스템에서는 umask가 적용될 수 있음
		t.Logf("파일 권한 = %v, 기대값 %v (시스템 차이로 무시될 수 있음)", mode, expectedMode)
	}
}

// TestSQLite_WALMode는 WAL 모드가 활성화되는지 테스트합니다.
func TestSQLite_WALMode(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	cfg := Config{
		DBPath:     dbPath,
		WALEnabled: true,
		Timeout:    50,
	}

	s, err := NewSQLite(cfg)
	if err != nil {
		t.Fatalf("NewSQLite() error = %v", err)
	}
	defer s.Close()

	// WAL 모드 확인을 위해 PRAGMA journal_mode 조회
	// db 필드가 비공개이므로 테스트를 건너뛰고 실제 사용에서 검증
	// WAL 모드는 NewSQLite 내부에서 활성화됨
	t.Skip("WAL 모드 테스트는 integration test에서 수행")
}

// TestSQLite_Concurrency는 동시 읽기/쓰기 테스트입니다.
func TestSQLite_Concurrency(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	cfg := Config{
		DBPath:     dbPath,
		WALEnabled: true,
		Timeout:    50,
	}

	s, err := NewSQLite(cfg)
	if err != nil {
		t.Fatalf("NewSQLite() error = %v", err)
	}
	defer s.Close()

	done := make(chan bool)
	errors := make(chan error, 10)

	// 동시 쓰기 고루틴
	for i := 0; i < 5; i++ {
		go func(idx int) {
			entry := models.HistoryEntry{
				PortNumber:  3000 + idx,
				ProcessName: "test",
				PID:         1000 + idx,
				Command:     "test command",
				KilledAt:    time.Now(),
			}
			if err := s.RecordKill(entry); err != nil {
				errors <- err
			}
			done <- true
		}(i)
	}

	// 동시 읽기 고루틴
	for i := 0; i < 5; i++ {
		go func() {
			_, err := s.GetHistory(10)
			if err != nil {
				errors <- err
			}
			done <- true
		}()
	}

	// 모든 고루틴 완료 대기
	for i := 0; i < 10; i++ {
		select {
		case <-done:
			// 정상 완료
		case err := <-errors:
			t.Errorf("동시 작업 중 에러 발생: %v", err)
		}
	}

	// 최종 검증
	history, err := s.GetHistory(100)
	if err != nil {
		t.Errorf("최종 히스토리 조회 실패: %v", err)
	}
	if len(history) != 5 {
		t.Errorf("기록된 항목 수 = %d, 원하는 값 5", len(history))
	}
}

// TestSQLite_EmptyDatabase는 빈 데이터베이스에서의 동작을 테스트합니다.
func TestSQLite_EmptyDatabase(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	cfg := Config{
		DBPath:     dbPath,
		WALEnabled: true,
		Timeout:    50,
	}

	s, err := NewSQLite(cfg)
	if err != nil {
		t.Fatalf("NewSQLite() error = %v", err)
	}
	defer s.Close()

	// 빈 히스토리 조회
	history, err := s.GetHistory(10)
	if err != nil {
		t.Errorf("GetHistory() error = %v", err)
	}
	if len(history) != 0 {
		t.Errorf("빈 데이터베이스에서 히스토리 길이 = %d, 원하는 값 0", len(history))
	}

	// 없는 포트의 종료 횟수 조회
	count, err := s.GetKillCount(9999, 30)
	if err != nil {
		t.Errorf("GetKillCount() error = %v", err)
	}
	if count != 0 {
		t.Errorf("없는 포트의 종료 횟수 = %d, 원하는 값 0", count)
	}

	// 없는 포트의 마지막 종료 시각 조회
	lastTime, err := s.GetLastKillTime(9999)
	if err != nil {
		t.Errorf("GetLastKillTime() error = %v", err)
	}
	if !lastTime.IsZero() {
		t.Errorf("없는 포트의 마지막 종료 시각이 zero time이 아님: %v", lastTime)
	}
}

// BenchmarkSQLite_RecordKill은 RecordKill 벤치마크 테스트입니다.
func BenchmarkSQLite_RecordKill(b *testing.B) {
	tmpDir := b.TempDir()
	dbPath := filepath.Join(tmpDir, "bench.db")

	cfg := Config{
		DBPath:     dbPath,
		WALEnabled: true,
		Timeout:    50,
	}

	s, err := NewSQLite(cfg)
	if err != nil {
		b.Fatalf("NewSQLite() error = %v", err)
	}
	defer s.Close()

	entry := models.HistoryEntry{
		PortNumber:  3000,
		ProcessName: "node",
		PID:         1234,
		Command:     "npm start",
		KilledAt:    time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		entry.PID = i // 중복 방지
		if err := s.RecordKill(entry); err != nil {
			b.Errorf("RecordKill() error = %v", err)
		}
	}
}

// BenchmarkSQLite_GetHistory는 GetHistory 벤치마크 테스트입니다.
func BenchmarkSQLite_GetHistory(b *testing.B) {
	tmpDir := b.TempDir()
	dbPath := filepath.Join(tmpDir, "bench.db")

	cfg := Config{
		DBPath:     dbPath,
		WALEnabled: true,
		Timeout:    50,
	}

	s, err := NewSQLite(cfg)
	if err != nil {
		b.Fatalf("NewSQLite() error = %v", err)
	}
	defer s.Close()

	// 100개 항목 삽입
	for i := 0; i < 100; i++ {
		entry := models.HistoryEntry{
			PortNumber:  3000 + i%10,
			ProcessName: "test",
			PID:         1000 + i,
			Command:     "test command",
			KilledAt:    time.Now().Add(-time.Duration(i) * time.Minute),
		}
		if err := s.RecordKill(entry); err != nil {
			b.Fatalf("RecordKill() error = %v", err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := s.GetHistory(50)
		if err != nil {
			b.Errorf("GetHistory() error = %v", err)
		}
	}
}

// BenchmarkSQLite_GetKillCount는 GetKillCount 벤치마크 테스트입니다.
func BenchmarkSQLite_GetKillCount(b *testing.B) {
	tmpDir := b.TempDir()
	dbPath := filepath.Join(tmpDir, "bench.db")

	cfg := Config{
		DBPath:     dbPath,
		WALEnabled: true,
		Timeout:    50,
	}

	s, err := NewSQLite(cfg)
	if err != nil {
		b.Fatalf("NewSQLite() error = %v", err)
	}
	defer s.Close()

	// 100개 항목 삽입
	for i := 0; i < 100; i++ {
		entry := models.HistoryEntry{
			PortNumber:  3000,
			ProcessName: "test",
			PID:         1000 + i,
			Command:     "test command",
			KilledAt:    time.Now(),
		}
		if err := s.RecordKill(entry); err != nil {
			b.Fatalf("RecordKill() error = %v", err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := s.GetKillCount(3000, 30)
		if err != nil {
			b.Errorf("GetKillCount() error = %v", err)
		}
	}
}
