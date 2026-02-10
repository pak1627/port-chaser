// Package storage는 SQLite 기반 히스토리 저장소 구현을 제공합니다.
package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3" // SQLite 드라이버

	"github.com/manson/port-chaser/internal/models"
)

// SQLite는 Storage 인터페이스의 SQLite 구현입니다.
type SQLite struct {
	db *sql.DB
}

// NewSQLite는 새로운 SQLite 저장소 인스턴스를 생성합니다.
func NewSQLite(cfg Config) (*SQLite, error) {
	if cfg.DBPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("홈 디렉토리를 찾을 수 없습니다: %w", err)
		}
		cfg.DBPath = filepath.Join(homeDir, ".port-chaser", "history.db")
	}

	// 데이터베이스 디렉토리 생성
	if err := os.MkdirAll(filepath.Dir(cfg.DBPath), 0755); err != nil {
		return nil, fmt.Errorf("디렉토리 생성 실패: %w", err)
	}

	// 데이터베이스 연결
	// SQLite _timeout은 밀리초 단위의 숫자만 사용합니다
	db, err := sql.Open("sqlite3", cfg.DBPath+"?_timeout="+fmt.Sprintf("%d", cfg.Timeout))
	if err != nil {
		return nil, fmt.Errorf("데이터베이스 연결 실패: %w", err)
	}

	// 연결 테스트
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("데이터베이스 핑 실패: %w", err)
	}

	s := &SQLite{db: db}

	// 테이블 생성
	if err := s.createTables(); err != nil {
		return nil, fmt.Errorf("테이블 생성 실패: %w", err)
	}

	// WAL 모드 활성화
	if cfg.WALEnabled {
		if err := s.enableWAL(); err != nil {
			// WAL 활성화 실패는 치명적이지 않음, 로깅만 수행
			fmt.Printf("WAL 모드 활성화 실패 (무시됨): %v\n", err)
		}
	}

	// 파일 권한 설정 (600 - 사용자 전용)
	if err := os.Chmod(cfg.DBPath, 0600); err != nil {
		// 권한 설정 실패는 치명적이지 않음
		fmt.Printf("파일 권한 설정 실패 (무시됨): %v\n", err)
	}

	return s, nil
}

// createTables는 필요한 테이블을 생성합니다.
func (s *SQLite) createTables() error {
	query := `
	CREATE TABLE IF NOT EXISTS history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		port_number INTEGER NOT NULL,
		process_name TEXT NOT NULL,
		pid INTEGER NOT NULL,
		command TEXT,
		killed_at DATETIME NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_history_port ON history(port_number);
	CREATE INDEX IF NOT EXISTS idx_history_killed_at ON history(killed_at);
	`

	_, err := s.db.Exec(query)
	return err
}

// enableWAL은 WAL(Wal-Ahead Logging) 모드를 활성화합니다.
func (s *SQLite) enableWAL() error {
	_, err := s.db.Exec("PRAGMA journal_mode=WAL;")
	return err
}

// RecordKill은 프로세스 종료 이벤트를 기록합니다.
func (s *SQLite) RecordKill(entry models.HistoryEntry) error {
	query := `
	INSERT INTO history (port_number, process_name, pid, command, killed_at)
	VALUES (?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(query, entry.PortNumber, entry.ProcessName, entry.PID, entry.Command, entry.KilledAt)
	if err != nil {
		return fmt.Errorf("히스토리 기록 실패: %w", err)
	}

	return nil
}

// GetHistory는 최근 종료 이력을 조회합니다.
func (s *SQLite) GetHistory(limit int) ([]models.HistoryEntry, error) {
	query := `
	SELECT id, port_number, process_name, pid, command, killed_at
	FROM history
	ORDER BY killed_at DESC
	LIMIT ?
	`

	rows, err := s.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("히스토리 조회 실패: %w", err)
	}
	defer rows.Close()

	var entries []models.HistoryEntry
	for rows.Next() {
		var entry models.HistoryEntry
		err := rows.Scan(&entry.ID, &entry.PortNumber, &entry.ProcessName, &entry.PID, &entry.Command, &entry.KilledAt)
		if err != nil {
			return nil, fmt.Errorf("히스토리 행 스캔 실패: %w", err)
		}
		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("히스토리 행 반복 오류: %w", err)
	}

	return entries, nil
}

// GetKillCount는 지정된 일수 동안 특정 포트의 프로세스가 종료된 횟수를 반환합니다.
func (s *SQLite) GetKillCount(port int, days int) (int, error) {
	since := time.Now().AddDate(0, 0, -days)

	query := `
	SELECT COUNT(*)
	FROM history
	WHERE port_number = ? AND killed_at >= ?
	`

	var count int
	err := s.db.QueryRow(query, port, since).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("종료 횟수 조회 실패: %w", err)
	}

	return count, nil
}

// GetLastKillTime은 특정 포트의 마지막 종료 시각을 반환합니다.
func (s *SQLite) GetLastKillTime(port int) (time.Time, error) {
	query := `
	SELECT killed_at
	FROM history
	WHERE port_number = ?
	ORDER BY killed_at DESC
	LIMIT 1
	`

	var killedAt time.Time
	err := s.db.QueryRow(query, port).Scan(&killedAt)
	if err == sql.ErrNoRows {
		return time.Time{}, nil // 종료 기록 없음
	}
	if err != nil {
		return time.Time{}, fmt.Errorf("마지막 종료 시각 조회 실패: %w", err)
	}

	return killedAt, nil
}

// Close는 저장소 연결을 닫습니다.
func (s *SQLite) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}
