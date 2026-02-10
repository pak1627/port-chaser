package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/manson/port-chaser/internal/models"
)

type SQLite struct {
	db *sql.DB
}

func NewSQLite(cfg Config) (*SQLite, error) {
	if cfg.DBPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to find home directory: %w", err)
		}
		cfg.DBPath = filepath.Join(homeDir, ".port-chaser", "history.db")
	}

	if err := os.MkdirAll(filepath.Dir(cfg.DBPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	db, err := sql.Open("sqlite3", cfg.DBPath+"?_timeout="+fmt.Sprintf("%d", cfg.Timeout))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	s := &SQLite{db: db}

	if err := s.createTables(); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	if cfg.WALEnabled {
		if err := s.enableWAL(); err != nil {
			fmt.Printf("WAL mode enable failed (ignored): %v\n", err)
		}
	}

	if err := os.Chmod(cfg.DBPath, 0600); err != nil {
		fmt.Printf("Failed to set file permissions (ignored): %v\n", err)
	}

	return s, nil
}

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

func (s *SQLite) enableWAL() error {
	_, err := s.db.Exec("PRAGMA journal_mode=WAL;")
	return err
}

func (s *SQLite) RecordKill(entry models.HistoryEntry) error {
	query := `
	INSERT INTO history (port_number, process_name, pid, command, killed_at)
	VALUES (?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(query, entry.PortNumber, entry.ProcessName, entry.PID, entry.Command, entry.KilledAt)
	if err != nil {
		return fmt.Errorf("failed to record history: %w", err)
	}

	return nil
}

func (s *SQLite) GetHistory(limit int) ([]models.HistoryEntry, error) {
	query := `
	SELECT id, port_number, process_name, pid, command, killed_at
	FROM history
	ORDER BY killed_at DESC
	LIMIT ?
	`

	rows, err := s.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get history: %w", err)
	}
	defer rows.Close()

	var entries []models.HistoryEntry
	for rows.Next() {
		var entry models.HistoryEntry
		err := rows.Scan(&entry.ID, &entry.PortNumber, &entry.ProcessName, &entry.PID, &entry.Command, &entry.KilledAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan history row: %w", err)
		}
		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating history rows: %w", err)
	}

	return entries, nil
}

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
		return 0, fmt.Errorf("failed to get kill count: %w", err)
	}

	return count, nil
}

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
		return time.Time{}, nil
	}
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to get last kill time: %w", err)
	}

	return killedAt, nil
}

func (s *SQLite) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}
