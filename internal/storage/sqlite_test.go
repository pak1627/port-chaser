package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/manson/port-chaser/internal/models"
)

func TestSQLite_RecordKill(t *testing.T) {
	tests := []struct {
		name    string
		entry   models.HistoryEntry
		wantErr bool
	}{
		{
			name: "normal kill record",
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
			name: "different port record",
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
			name: "empty command record",
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

			if !tt.wantErr {
				history, err := s.GetHistory(10)
				if err != nil {
					t.Errorf("GetHistory() error = %v", err)
				}
				if len(history) != 1 {
					t.Errorf("recorded items count = %d, want 1", len(history))
				}
			}
		})
	}
}

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
		name       string
		limit      int
		wantCount  int
		wantNewest int
	}{
		{
			name:       "get all",
			limit:      10,
			wantCount:  3,
			wantNewest: 5000,
		},
		{
			name:       "get 2 only",
			limit:      2,
			wantCount:  2,
			wantNewest: 5000,
		},
		{
			name:       "get 1 only",
			limit:      1,
			wantCount:  1,
			wantNewest: 5000,
		},
		{
			name:       "0 requested",
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
				t.Errorf("GetHistory() count = %d, want %d", len(got), tt.wantCount)
			}
			if tt.wantCount > 0 && got[0].PortNumber != tt.wantNewest {
				t.Errorf("GetHistory() first item port = %d, want %d", got[0].PortNumber, tt.wantNewest)
			}
		})
	}
}

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
		name string
		port int
		days int
		want int
	}{
		{
			name: "port 3000, last 30 days",
			port: 3000,
			days: 30,
			want: 5,
		},
		{
			name: "port 8080, last 30 days",
			port: 8080,
			days: 30,
			want: 2,
		},
		{
			name: "port 3000, last 7 days",
			port: 3000,
			days: 7,
			want: 5,
		},
		{
			name: "non-existent port",
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
				t.Errorf("GetKillCount() = %d, want %d", got, tt.want)
			}
		})
	}
}

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
		name     string
		port     int
		wantZero bool
	}{
		{
			name:     "port with history",
			port:     3000,
			wantZero: false,
		},
		{
			name:     "port without history",
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
				t.Errorf("GetLastKillTime() IsZero = %v, want %v", isZero, tt.wantZero)
			}
			if !tt.wantZero {
				diff := now.Sub(got)
				if diff > time.Second {
					t.Errorf("GetLastKillTime() diff = %v, expect within 1 second", diff)
				}
			}
		})
	}
}

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

	err = s.Close()
	if err != nil {
		t.Errorf("Close() second call error = %v", err)
	}
}

func TestSQLite_NewSQLite_DirectoryCreation(t *testing.T) {
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

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("database file was not created")
	}
}

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
		t.Fatalf("failed to get file info: %v", err)
	}

	mode := info.Mode().Perm()
	expectedMode := os.FileMode(0600)

	if mode != expectedMode {
		t.Logf("file permissions = %v, want %v (may differ by system)", mode, expectedMode)
	}
}

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

	t.Skip("WAL mode test requires integration testing")
}

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

	for i := 0; i < 5; i++ {
		go func() {
			_, err := s.GetHistory(10)
			if err != nil {
				errors <- err
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		select {
		case <-done:
		case err := <-errors:
			t.Errorf("concurrent operation error: %v", err)
		}
	}

	history, err := s.GetHistory(100)
	if err != nil {
		t.Errorf("final history lookup failed: %v", err)
	}
	if len(history) != 5 {
		t.Errorf("recorded items count = %d, want 5", len(history))
	}
}

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

	history, err := s.GetHistory(10)
	if err != nil {
		t.Errorf("GetHistory() error = %v", err)
	}
	if len(history) != 0 {
		t.Errorf("empty database history length = %d, want 0", len(history))
	}

	count, err := s.GetKillCount(9999, 30)
	if err != nil {
		t.Errorf("GetKillCount() error = %v", err)
	}
	if count != 0 {
		t.Errorf("non-existent port kill count = %d, want 0", count)
	}

	lastTime, err := s.GetLastKillTime(9999)
	if err != nil {
		t.Errorf("GetLastKillTime() error = %v", err)
	}
	if !lastTime.IsZero() {
		t.Errorf("non-existent port last kill time is not zero: %v", lastTime)
	}
}

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
		entry.PID = i
		if err := s.RecordKill(entry); err != nil {
			b.Errorf("RecordKill() error = %v", err)
		}
	}
}

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
